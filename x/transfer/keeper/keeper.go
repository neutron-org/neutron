package transfer

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	feekeeper "github.com/neutron-org/neutron/x/fee/keeper"
	wrappedtypes "github.com/neutron-org/neutron/x/transfer/types"
)

// KeeperTransferWrapper is a wrapper for original ibc keeper to override response for "Transfer" method
type KeeperTransferWrapper struct {
	keeper.Keeper
	channelKeeper types.ChannelKeeper
	FeeKeeper     *feekeeper.Keeper
}

func (k KeeperTransferWrapper) Transfer(goCtx context.Context, msg *wrappedtypes.MsgTransfer) (*wrappedtypes.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		k.Logger(ctx).Debug("Transfer: failed to parse sender address", "sender", msg.Sender)
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.Sender)
	}

	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, msg.SourcePort, msg.SourceChannel)
	if !found {
		return nil, sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", msg.SourcePort, msg.SourceChannel,
		)
	}

	if msg.PayerFee == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fee can't be nil")
	}

	if err := k.FeeKeeper.LockFees(ctx, senderAddr, channeltypes.NewPacketId(msg.SourcePort, msg.SourceChannel, sequence), msg.PayerFee); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to lock fees to pay for transfer msg: %v", msg)
	}

	if _, err := k.Keeper.Transfer(goCtx, types.NewMsgTransfer(msg.SourcePort, msg.SourceChannel, msg.Token, msg.Sender, msg.Receiver, msg.TimeoutHeight, msg.TimeoutTimestamp)); err != nil {
		return nil, err
	}

	return &wrappedtypes.MsgTransferResponse{
		SequenceId: sequence,
		Channel:    msg.SourceChannel,
	}, nil
}

// NewKeeper creates a new IBC transfer Keeper(KeeperTransferWrapper) instance
func NewKeeper(
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	ics4Wrapper types.ICS4Wrapper, channelKeeper types.ChannelKeeper, portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, scopedKeeper capabilitykeeper.ScopedKeeper,
	feeKeeper *feekeeper.Keeper,
) KeeperTransferWrapper {
	return KeeperTransferWrapper{
		Keeper: keeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper,
			authKeeper, bankKeeper, scopedKeeper),
		channelKeeper: channelKeeper,
		FeeKeeper:     feeKeeper,
	}
}
