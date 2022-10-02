package keeper

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"

	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

type msgServer struct {
	Keeper
}

var _ ictxtypes.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) ictxtypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k Keeper) RegisterInterchainAccount(goCtx context.Context, msg *ictxtypes.MsgRegisterInterchainAccount) (*ictxtypes.MsgRegisterInterchainAccountResponse, error) {
	defer telemetry.ModuleMeasureSince(ictxtypes.ModuleName, time.Now(), LabelRegisterInterchainAccount)

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("RegisterInterchainAccount", "connection_id", msg.ConnectionId, "from_address", msg.FromAddress, "interchain_account_id", msg.InterchainAccountId)

	senderAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		k.Logger(ctx).Debug("RegisterInterchainAccount: failed to parse sender address", "from_address", msg.FromAddress)
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.FromAddress)
	}

	if !k.wasmKeeper.HasContractInfo(ctx, senderAddr) {
		k.Logger(ctx).Debug("RegisterInterchainAccount: contract not found", "from_address", msg.FromAddress)
		return nil, sdkerrors.Wrapf(ictxtypes.ErrNotContract, "%s is not a contract address", msg.FromAddress)
	}

	icaOwner, err := ictxtypes.NewICAOwner(msg.FromAddress, msg.InterchainAccountId)
	if err != nil {
		k.Logger(ctx).Debug("RegisterInterchainAccount: failed to create RegisterInterchainAccount", "error", err)
		return nil, sdkerrors.Wrap(err, "failed to create ICA owner")
	}

	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, icaOwner.String()); err != nil {
		k.Logger(ctx).Debug("RegisterInterchainAccount: failed to create RegisterInterchainAccount:", "error", err, "owner", icaOwner.String(), "msg", &msg)
		return nil, sdkerrors.Wrap(err, "failed to RegisterInterchainAccount")
	}

	return &ictxtypes.MsgRegisterInterchainAccountResponse{}, nil
}

func (k Keeper) SubmitTx(goCtx context.Context, msg *ictxtypes.MsgSubmitTx) (*ictxtypes.MsgSubmitTxResponse, error) {
	defer telemetry.ModuleMeasureSince(ictxtypes.ModuleName, time.Now(), LabelSubmitTx)

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("SubmitTx", "connection_id", msg.ConnectionId, "from_address", msg.FromAddress, "interchain_account_id", msg.InterchainAccountId)

	senderAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		k.Logger(ctx).Debug("SubmitTx: failed to parse sender address", "from_address", msg.FromAddress)
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.FromAddress)
	}

	if !k.wasmKeeper.HasContractInfo(ctx, senderAddr) {
		k.Logger(ctx).Debug("SubmitTx: contract not found", "from_address", msg.FromAddress)
		return nil, sdkerrors.Wrapf(ictxtypes.ErrNotContract, "%s is not a contract address", msg.FromAddress)
	}

	icaOwner, err := ictxtypes.NewICAOwner(msg.FromAddress, msg.InterchainAccountId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to create ICA owner")
	}

	portID, err := icatypes.NewControllerPortID(icaOwner.String())
	if err != nil {
		k.Logger(ctx).Error("SubmitTx: failed to create NewControllerPortID:", "error", err, "owner", icaOwner)
		return nil, sdkerrors.Wrap(err, "failed to create NewControllerPortID")
	}

	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, msg.ConnectionId, portID)
	if !found {
		k.Logger(ctx).Debug("SubmitTx: failed to GetActiveChannelID", "connection_id", msg.ConnectionId, "port_id", portID)
		return nil, sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to GetActiveChannelID for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		k.Logger(ctx).Debug("SubmitTx: failed to GetCapability", "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "failed to GetCapability")
	}

	sdkMsgs, err := msg.GetTxMsgs()
	if err != nil {
		k.Logger(ctx).Debug("SubmitTx: failed to GetTxMsgs", "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(err, "failed to GetTxMsgs")
	}

	data, err := icatypes.SerializeCosmosTx(k.Codec, sdkMsgs)
	if err != nil {
		k.Logger(ctx).Debug("SubmitTx: failed to SerializeCosmosTx", "error", err, "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(err, "failed to SerializeCosmosTx")
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: msg.Memo,
	}

	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, portID, channelID)
	if !found {
		return nil, sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", portID, channelID,
		)
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Duration(msg.Timeout) * time.Second).UnixNano()
	_, err = k.icaControllerKeeper.SendTx(ctx, chanCap, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		// usually we use DEBUG level for such errors, but in this case we have checked full input before running SendTX, so error here may be critical
		k.Logger(ctx).Error("SubmitTx", "error", err, "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(err, "failed to SendTx")
	}

	return &ictxtypes.MsgSubmitTxResponse{
		SequenceId: sequence,
		Channel:    channelID,
	}, nil
}
