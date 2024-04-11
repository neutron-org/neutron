package transfer

import (
	"context"

	"cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"

	feetypes "github.com/neutron-org/neutron/v3/x/feerefunder/types"
	wrappedtypes "github.com/neutron-org/neutron/v3/x/transfer/types"
)

// KeeperTransferWrapper is a wrapper for original ibc keeper to override response for "Transfer" method
type KeeperTransferWrapper struct {
	keeper.Keeper
	channelKeeper wrappedtypes.ChannelKeeper
	FeeKeeper     wrappedtypes.FeeRefunderKeeper
	SudoKeeper    wrappedtypes.WasmKeeper
}

func (k KeeperTransferWrapper) Transfer(goCtx context.Context, msg *wrappedtypes.MsgTransfer) (*wrappedtypes.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		k.Logger(ctx).Debug("Transfer: failed to parse sender address", "sender", msg.Sender)
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.Sender)
	}

	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, msg.SourcePort, msg.SourceChannel)
	if !found {
		return nil, errors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", msg.SourcePort, msg.SourceChannel,
		)
	}

	// if the sender is a contract, lock fees.
	// Because contracts are required to pay fees for the acknowledgements
	if k.SudoKeeper.HasContractInfo(ctx, senderAddr) {
		if err := k.FeeKeeper.LockFees(ctx, senderAddr, feetypes.NewPacketID(msg.SourcePort, msg.SourceChannel, sequence), msg.Fee); err != nil {
			return nil, errors.Wrapf(err, "failed to lock fees to pay for transfer msg: %v", msg)
		}
	}

	transferMsg := types.NewMsgTransfer(msg.SourcePort, msg.SourceChannel, msg.Token, msg.Sender, msg.Receiver, msg.TimeoutHeight, msg.TimeoutTimestamp, msg.Memo)
	if _, err := k.Keeper.Transfer(goCtx, transferMsg); err != nil {
		return nil, err
	}

	return &wrappedtypes.MsgTransferResponse{
		SequenceId: sequence,
		Channel:    msg.SourceChannel,
	}, nil
}

// NewKeeper creates a new IBC transfer Keeper(KeeperTransferWrapper) instance
func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper, channelKeeper wrappedtypes.ChannelKeeper, portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, scopedKeeper capabilitykeeper.ScopedKeeper,
	feeKeeper wrappedtypes.FeeRefunderKeeper,
	sudoKeeper wrappedtypes.WasmKeeper,
) KeeperTransferWrapper {
	return KeeperTransferWrapper{
		channelKeeper: channelKeeper,
		Keeper: keeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper,
			authKeeper, bankKeeper, scopedKeeper),
		FeeKeeper:  feeKeeper,
		SudoKeeper: sudoKeeper,
	}
}
