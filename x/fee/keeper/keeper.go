package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/neutron-org/neutron/x/fee/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		bankKeeper types.BankKeeper
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
		ibcKeeper  *ibckeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	ibcKeeper *ibckeeper.Keeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{

		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		ibcKeeper:  ibcKeeper,
		bankKeeper: bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) LockFees(ctx sdk.Context, payer sdk.AccAddress, packetID channeltypes.PacketId, fee *ibcfeetypes.Fee) error {
	store := ctx.KVStore(k.storeKey)

	if _, ok := k.ibcKeeper.ChannelKeeper.GetChannel(ctx, packetID.PortId, packetID.ChannelId); !ok {
		return sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "channel with id %s and port %s not found", packetID.ChannelId, packetID.PortId)
	}

	feeInfo := types.FeeInfo{
		Payer:    payer.String(),
		PayerFee: fee,
	}
	bzFeeInfo := k.cdc.MustMarshal(&feeInfo)
	store.Set(types.GetFeePacketKey(packetID.ChannelId, packetID.PortId, packetID.Sequence), bzFeeInfo)

	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, payer, types.ModuleName, fee.Total())
}

func (k Keeper) distributeFee(ctx sdk.Context, receiver, refundAccAddress sdk.AccAddress, fee sdk.Coins) error {
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, fee)
	if err != nil {
		if bytes.Equal(receiver, refundAccAddress) {
			k.Logger(ctx).Error("error distributing fee", "receiver address", receiver, "fee", fee)
			return sdkerrors.Wrapf(err, "error distributing fee to a receiver: %s", receiver.String())
		}

		// if an error is returned from x/bank and the receiver is not the refundAccAddress
		// then attempt to refund the fee to the original sender
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, refundAccAddress, fee)
		if err != nil {
			return sdkerrors.Wrapf(err, "error refunding fee to the original payer: %s", refundAccAddress.String())
		}
	}
	return nil
}

func (k Keeper) DistributeAcknowledgementFee(ctx sdk.Context, receiver sdk.AccAddress, packetID channeltypes.PacketId) {
	store := ctx.KVStore(k.storeKey)

	var feeInfo types.FeeInfo
	bzFeeInfo := store.Get(types.GetFeePacketKey(packetID.ChannelId, packetID.PortId, packetID.Sequence))
	if bzFeeInfo == nil {
		k.Logger(ctx).Error("no fee info", "channelID", packetID.ChannelId, "portID", packetID.PortId, "sequence", packetID.Sequence)
		return
	}
	k.cdc.MustUnmarshal(bzFeeInfo, &feeInfo)

	cacheCtx, writeFn := ctx.CacheContext()

	// try to distribute ack fee
	if err := k.distributeFee(cacheCtx, receiver, sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.PayerFee.AckFee); err != nil {
		k.Logger(ctx).Error("error distributing ack fee", "receiver", receiver, "payer", feeInfo.Payer, "packet", packetID)
		return
	}

	// try to return unused timeout and recv packet fee
	if err := k.distributeFee(cacheCtx, sdk.MustAccAddressFromBech32(feeInfo.Payer), sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.PayerFee.TimeoutFee); err != nil {
		k.Logger(ctx).Error("error returning unused timeout fee", "receiver", feeInfo.Payer, "packet", packetID)
		return
	}
	if err := k.distributeFee(cacheCtx, sdk.MustAccAddressFromBech32(feeInfo.Payer), sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.PayerFee.RecvFee); err != nil {
		k.Logger(ctx).Error("error returning unused recv fee", "receiver", feeInfo.Payer, "packet", packetID)
		return
	}

	ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())

	writeFn()

	k.removeFeeInfo(cacheCtx, packetID)
}

func (k Keeper) DistributeTimeoutFee(ctx sdk.Context, receiver sdk.AccAddress, packetID channeltypes.PacketId) {
	store := ctx.KVStore(k.storeKey)

	var feeInfo types.FeeInfo
	bzFeeInfo := store.Get(types.GetFeePacketKey(packetID.ChannelId, packetID.PortId, packetID.Sequence))
	if bzFeeInfo == nil {
		k.Logger(ctx).Error("no fee info", "channelID", packetID.ChannelId, "portID", packetID.PortId, "sequence", packetID.Sequence)
		return
	}
	k.cdc.MustUnmarshal(bzFeeInfo, &feeInfo)

	cacheCtx, writeFn := ctx.CacheContext()

	// try to distribute timeout fee
	if err := k.distributeFee(cacheCtx, receiver, sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.PayerFee.TimeoutFee); err != nil {
		k.Logger(ctx).Error("error distributing timeout fee", "receiver", receiver, "payer", feeInfo.Payer, "packet", packetID)
		return
	}

	// try to return unused ack and recv packet fee
	if err := k.distributeFee(cacheCtx, sdk.MustAccAddressFromBech32(feeInfo.Payer), sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.PayerFee.AckFee); err != nil {
		k.Logger(ctx).Error("error returning unused ack fee", "receiver", feeInfo.Payer, "packet", packetID)
		return
	}
	if err := k.distributeFee(cacheCtx, sdk.MustAccAddressFromBech32(feeInfo.Payer), sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.PayerFee.RecvFee); err != nil {
		k.Logger(ctx).Error("error returning unused recv fee", "receiver", feeInfo.Payer, "packet", packetID)
		return
	}

	ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())

	writeFn()

	k.removeFeeInfo(cacheCtx, packetID)
}

func (k Keeper) GetFeeInfo(ctx sdk.Context, packetID channeltypes.PacketId) (*types.FeeInfo, error) {
	store := ctx.KVStore(k.storeKey)

	var feeInfo types.FeeInfo
	bzFeeInfo := store.Get(types.GetFeePacketKey(packetID.ChannelId, packetID.PortId, packetID.Sequence))
	if bzFeeInfo == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "no fee info for the given channelID = %s, portID = %s and sequence = %d", packetID.ChannelId, packetID.PortId, packetID.Sequence)
	}
	k.cdc.MustUnmarshal(bzFeeInfo, &feeInfo)

	return &feeInfo, nil
}

func (k Keeper) removeFeeInfo(ctx sdk.Context, packetID channeltypes.PacketId) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.GetFeePacketKey(packetID.ChannelId, packetID.PortId, packetID.Sequence))
}
