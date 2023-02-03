package keeper

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/neutron-org/neutron/x/feerefunder/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		bankKeeper    types.BankKeeper
		storeKey      storetypes.StoreKey
		memKey        storetypes.StoreKey
		paramstore    paramtypes.Subspace
		channelKeeper types.ChannelKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	channelKeeper types.ChannelKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		channelKeeper: channelKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) LockFees(ctx sdk.Context, payer sdk.AccAddress, packetID types.PacketID, fee types.Fee) error {
	k.Logger(ctx).Debug("Trying to lock fees", "packetID", packetID, "fee", fee)

	if _, ok := k.channelKeeper.GetChannel(ctx, packetID.PortId, packetID.ChannelId); !ok {
		return sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "channel with id %s and port %s not found", packetID.ChannelId, packetID.PortId)
	}

	if err := k.checkFees(ctx, fee); err != nil {
		return sdkerrors.Wrapf(err, "failed to lock fees")
	}

	feeInfo := types.FeeInfo{
		Payer:    payer.String(),
		Fee:      fee,
		PacketId: packetID,
	}
	k.StoreFeeInfo(ctx, feeInfo)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, payer, types.ModuleName, fee.Total()); err != nil {
		return sdkerrors.Wrapf(err, "failed to send coins during fees locking")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLockFees,
			sdk.NewAttribute(types.AttributeKeyPayer, payer.String()),
			sdk.NewAttribute(types.AttributeKeyPortID, packetID.PortId),
			sdk.NewAttribute(types.AttributeKeyChannelID, packetID.ChannelId),
			sdk.NewAttribute(types.AttributeKeySequence, strconv.FormatUint(packetID.Sequence, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})

	return nil
}

func (k Keeper) DistributeAcknowledgementFee(ctx sdk.Context, receiver sdk.AccAddress, packetID types.PacketID) {
	k.Logger(ctx).Debug("Trying to distribute ack fee", "packetID", packetID)
	feeInfo, err := k.GetFeeInfo(ctx, packetID)
	if err != nil {
		k.Logger(ctx).Error("no fee info", "error", err)
		panic(sdkerrors.Wrapf(err, "no fee info"))
	}

	// try to distribute ack fee
	if err := k.distributeFee(ctx, receiver, feeInfo.Fee.AckFee); err != nil {
		k.Logger(ctx).Error("error distributing ack fee", "receiver", receiver, "payer", feeInfo.Payer, "packet", packetID)
		panic(sdkerrors.Wrapf(err, "error distributing ack fee: receiver = %s, packetID=%v", receiver, packetID))
	}

	// try to return unused timeout fee
	if err := k.distributeFee(ctx, sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.Fee.TimeoutFee); err != nil {
		k.Logger(ctx).Error("error returning unused timeout fee", "receiver", feeInfo.Payer, "packet", packetID)
		panic(sdkerrors.Wrapf(err, "error distributing unused timeout fee: receiver = %s, packetID=%v", receiver, packetID))
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDistributeAcknowledgementFee,
			sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
			sdk.NewAttribute(types.AttributeKeyPortID, packetID.PortId),
			sdk.NewAttribute(types.AttributeKeyChannelID, packetID.ChannelId),
			sdk.NewAttribute(types.AttributeKeySequence, strconv.FormatUint(packetID.Sequence, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})

	k.removeFeeInfo(ctx, packetID)
}

func (k Keeper) DistributeTimeoutFee(ctx sdk.Context, receiver sdk.AccAddress, packetID types.PacketID) {
	k.Logger(ctx).Debug("Trying to distribute timeout fee", "packetID", packetID)
	feeInfo, err := k.GetFeeInfo(ctx, packetID)
	if err != nil {
		k.Logger(ctx).Error("no fee info", "error", err)
		panic(sdkerrors.Wrapf(err, "no fee info"))
	}

	// try to distribute timeout fee
	if err := k.distributeFee(ctx, receiver, feeInfo.Fee.TimeoutFee); err != nil {
		k.Logger(ctx).Error("error distributing timeout fee", "receiver", receiver, "payer", feeInfo.Payer, "packet", packetID)
		panic(sdkerrors.Wrapf(err, "error distributing timeout fee: receiver = %s, packetID=%v", receiver, packetID))
	}

	// try to return unused ack fee
	if err := k.distributeFee(ctx, sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.Fee.AckFee); err != nil {
		k.Logger(ctx).Error("error returning unused ack fee", "receiver", feeInfo.Payer, "packet", packetID)
		panic(sdkerrors.Wrapf(err, "error distributing unused ack fee: receiver = %s, packetID=%v", receiver, packetID))
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDistributeTimeoutFee,
			sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
			sdk.NewAttribute(types.AttributeKeyPortID, packetID.PortId),
			sdk.NewAttribute(types.AttributeKeyChannelID, packetID.ChannelId),
			sdk.NewAttribute(types.AttributeKeySequence, strconv.FormatUint(packetID.Sequence, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})

	k.removeFeeInfo(ctx, packetID)
}

func (k Keeper) GetFeeInfo(ctx sdk.Context, packetID types.PacketID) (*types.FeeInfo, error) {
	store := ctx.KVStore(k.storeKey)

	var feeInfo types.FeeInfo
	bzFeeInfo := store.Get(types.GetFeePacketKey(packetID))
	if bzFeeInfo == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "no fee info for the given channelID = %s, portID = %s and sequence = %d", packetID.ChannelId, packetID.PortId, packetID.Sequence)
	}
	k.cdc.MustUnmarshal(bzFeeInfo, &feeInfo)

	return &feeInfo, nil
}

func (k Keeper) GetAllFeeInfos(ctx sdk.Context) []types.FeeInfo {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.FeeKey)

	infos := make([]types.FeeInfo, 0)

	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info types.FeeInfo
		k.cdc.MustUnmarshal(iterator.Value(), &info)
		infos = append(infos, info)
	}

	return infos
}

func (k Keeper) StoreFeeInfo(ctx sdk.Context, feeInfo types.FeeInfo) {
	store := ctx.KVStore(k.storeKey)

	bzFeeInfo := k.cdc.MustMarshal(&feeInfo)
	store.Set(types.GetFeePacketKey(feeInfo.PacketId), bzFeeInfo)
}

func (k Keeper) removeFeeInfo(ctx sdk.Context, packetID types.PacketID) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.GetFeePacketKey(packetID))
}

func (k Keeper) checkFees(ctx sdk.Context, fees types.Fee) error {
	params := k.GetParams(ctx)

	if !fees.TimeoutFee.IsAnyGTE(params.MinFee.TimeoutFee) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "provided timeout fee is less than min governance set timeout fee: %v < %v", fees.TimeoutFee, params.MinFee.TimeoutFee)
	}

	if !fees.AckFee.IsAnyGTE(params.MinFee.AckFee) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "provided ack fee is less than min governance set ack fee: %v < %v", fees.AckFee, params.MinFee.AckFee)
	}

	// we don't allow users to set recv fees, because we can't refund relayers for such messages
	if !fees.RecvFee.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "recv fee must be zero")
	}

	return nil
}

func (k Keeper) distributeFee(ctx sdk.Context, receiver sdk.AccAddress, fee sdk.Coins) error {
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, fee)
	if err != nil {
		k.Logger(ctx).Error("error distributing fee", "receiver address", receiver, "fee", fee)
		return sdkerrors.Wrapf(err, "error distributing fee to a receiver: %s", receiver.String())
	}
	return nil
}
