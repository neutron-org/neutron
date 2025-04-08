package keeper

import (
	"context"
	"fmt"
	"strconv"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/neutron-org/neutron/v6/x/feerefunder/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		bankKeeper    types.BankKeeper
		storeKey      storetypes.StoreKey
		memKey        storetypes.StoreKey
		channelKeeper types.ChannelKeeper
		authority     string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	channelKeeper types.ChannelKeeper,
	bankKeeper types.BankKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		channelKeeper: channelKeeper,
		bankKeeper:    bankKeeper,
		authority:     authority,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) LockFees(ctx context.Context, payer sdk.AccAddress, packetID types.PacketID, fee types.Fee) error {
	c := sdk.UnwrapSDKContext(ctx)

	k.Logger(c).Debug("Trying to lock fees", "packetID", packetID, "fee", fee)

	if _, ok := k.channelKeeper.GetChannel(c, packetID.PortId, packetID.ChannelId); !ok {
		return errors.Wrapf(channeltypes.ErrChannelNotFound, "channel with id %s and port %s not found", packetID.ChannelId, packetID.PortId)
	}

	if err := k.checkFees(c, fee); err != nil {
		return errors.Wrapf(err, "failed to lock fees")
	}

	feeInfo := types.FeeInfo{
		Payer:    payer.String(),
		Fee:      fee,
		PacketId: packetID,
	}
	k.StoreFeeInfo(c, feeInfo)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, payer, types.ModuleName, fee.Total()); err != nil {
		return errors.Wrapf(err, "failed to send coins during fees locking")
	}

	c.EventManager().EmitEvents(sdk.Events{
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

func (k Keeper) DistributeAcknowledgementFee(ctx context.Context, receiver sdk.AccAddress, packetID types.PacketID) {
	c := sdk.UnwrapSDKContext(ctx)

	k.Logger(c).Debug("Trying to distribute ack fee", "packetID", packetID)
	feeInfo, err := k.GetFeeInfo(c, packetID)
	if err != nil {
		k.Logger(c).Error("no fee info", "error", err)
		panic(errors.Wrapf(err, "no fee info"))
	}

	// try to distribute ack fee
	if err := k.distributeFee(c, receiver, feeInfo.Fee.AckFee); err != nil {
		k.Logger(c).Error("error distributing ack fee", "receiver", receiver, "payer", feeInfo.Payer, "packet", packetID)
		panic(errors.Wrapf(err, "error distributing ack fee: receiver = %s, packetID=%v", receiver, packetID))
	}

	// try to return unused timeout fee
	if err := k.distributeFee(c, sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.Fee.TimeoutFee); err != nil {
		k.Logger(c).Error("error returning unused timeout fee", "receiver", feeInfo.Payer, "packet", packetID)
		panic(errors.Wrapf(err, "error distributing unused timeout fee: receiver = %s, packetID=%v", feeInfo.Payer, packetID))
	}

	c.EventManager().EmitEvents(sdk.Events{
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

	k.removeFeeInfo(c, packetID)
}

func (k Keeper) DistributeTimeoutFee(ctx context.Context, receiver sdk.AccAddress, packetID types.PacketID) {
	c := sdk.UnwrapSDKContext(ctx)

	k.Logger(c).Debug("Trying to distribute timeout fee", "packetID", packetID)
	feeInfo, err := k.GetFeeInfo(c, packetID)
	if err != nil {
		k.Logger(c).Error("no fee info", "error", err)
		panic(errors.Wrapf(err, "no fee info"))
	}

	// try to distribute timeout fee
	if err := k.distributeFee(c, receiver, feeInfo.Fee.TimeoutFee); err != nil {
		k.Logger(c).Error("error distributing timeout fee", "receiver", receiver, "payer", feeInfo.Payer, "packet", packetID)
		panic(errors.Wrapf(err, "error distributing timeout fee: receiver = %s, packetID=%v", receiver, packetID))
	}

	// try to return unused ack fee
	if err := k.distributeFee(c, sdk.MustAccAddressFromBech32(feeInfo.Payer), feeInfo.Fee.AckFee); err != nil {
		k.Logger(c).Error("error returning unused ack fee", "receiver", feeInfo.Payer, "packet", packetID)
		panic(errors.Wrapf(err, "error distributing unused ack fee: receiver = %s, packetID=%v", feeInfo.Payer, packetID))
	}

	c.EventManager().EmitEvents(sdk.Events{
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

	k.removeFeeInfo(c, packetID)
}

func (k Keeper) GetFeeInfo(ctx sdk.Context, packetID types.PacketID) (*types.FeeInfo, error) {
	store := ctx.KVStore(k.storeKey)

	var feeInfo types.FeeInfo
	bzFeeInfo := store.Get(types.GetFeePacketKey(packetID))
	if bzFeeInfo == nil {
		return nil, errors.Wrapf(sdkerrors.ErrKeyNotFound, "no fee info for the given channelID = %s, portID = %s and sequence = %d", packetID.ChannelId, packetID.PortId, packetID.Sequence)
	}
	k.cdc.MustUnmarshal(bzFeeInfo, &feeInfo)

	return &feeInfo, nil
}

func (k Keeper) GetAllFeeInfos(ctx sdk.Context) []types.FeeInfo {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.FeeKey)

	infos := make([]types.FeeInfo, 0)

	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
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

func (k Keeper) GetMinFee(ctx sdk.Context) types.Fee {
	params := k.GetParams(ctx)
	return params.GetMinFee()
}

func (k Keeper) removeFeeInfo(ctx sdk.Context, packetID types.PacketID) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.GetFeePacketKey(packetID))
}

func (k Keeper) checkFees(ctx sdk.Context, fees types.Fee) error {
	params := k.GetParams(ctx)

	if !fees.TimeoutFee.IsAnyGTE(params.MinFee.TimeoutFee) {
		return errors.Wrapf(sdkerrors.ErrInsufficientFee, "provided timeout fee is less than min governance set timeout fee: %v < %v", fees.TimeoutFee, params.MinFee.TimeoutFee)
	}

	if !fees.AckFee.IsAnyGTE(params.MinFee.AckFee) {
		return errors.Wrapf(sdkerrors.ErrInsufficientFee, "provided ack fee is less than min governance set ack fee: %v < %v", fees.AckFee, params.MinFee.AckFee)
	}

	if allowedCoins(fees.TimeoutFee, params.MinFee.TimeoutFee) {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "timeout fee cannot have coins other than in params")
	}

	if allowedCoins(fees.AckFee, params.MinFee.AckFee) {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "ack fee cannot have coins other than in params")
	}

	// we don't allow users to set recv fees, because we can't refund relayers for such messages
	if !fees.RecvFee.IsZero() {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "recv fee must be zero")
	}

	return nil
}

func (k Keeper) distributeFee(ctx sdk.Context, receiver sdk.AccAddress, fee sdk.Coins) error {
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, fee)
	if err != nil {
		k.Logger(ctx).Error("error distributing fee", "receiver address", receiver, "fee", fee)
		return errors.Wrapf(err, "error distributing fee to a receiver: %s", receiver.String())
	}
	return nil
}

// allowedCoins returns true if one or more coins from `fees` are not present in coins from `params`
// assumes that `params` is sorted
func allowedCoins(fees, params sdk.Coins) bool {
	for _, fee := range fees {
		if params.AmountOf(fee.Denom).IsZero() {
			return true
		}
	}
	return false
}
