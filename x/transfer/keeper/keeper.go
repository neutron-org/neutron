package transfer

import (
	"context"

	"cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"

	feetypes "github.com/neutron-org/neutron/v6/x/feerefunder/types"
	wrappedtypes "github.com/neutron-org/neutron/v6/x/transfer/types"
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

	isContract := k.SudoKeeper.HasContractInfo(ctx, senderAddr)

	if err := msg.Validate(isContract); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgTransfer")
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
	if isContract {
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

func (k KeeperTransferWrapper) UpdateParams(goCtx context.Context, msg *wrappedtypes.MsgUpdateParams) (*wrappedtypes.MsgUpdateParamsResponse, error) {
	newMsg := &types.MsgUpdateParams{
		Signer: msg.Signer,
		Params: msg.Params,
	}
	if _, err := k.Keeper.UpdateParams(goCtx, newMsg); err != nil {
		return nil, err
	}

	return &wrappedtypes.MsgUpdateParamsResponse{}, nil
}

// NewKeeper creates a new IBC transfer Keeper(KeeperTransferWrapper) instance
func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper, channelKeeper wrappedtypes.ChannelKeeper, portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, scopedKeeper capabilitykeeper.ScopedKeeper,
	feeKeeper wrappedtypes.FeeRefunderKeeper,
	sudoKeeper wrappedtypes.WasmKeeper, authority string,
) KeeperTransferWrapper {
	return KeeperTransferWrapper{
		channelKeeper: channelKeeper,
		Keeper: keeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper,
			authKeeper, bankKeeper, scopedKeeper, authority),
		FeeKeeper:  feeKeeper,
		SudoKeeper: sudoKeeper,
	}
}
