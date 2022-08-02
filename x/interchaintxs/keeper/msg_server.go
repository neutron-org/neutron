package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"time"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	"github.com/neutron-org/neutron/x/interchaintxs/types"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

// InterchainTxTimeout defines the IBC timeout of the interchain transaction.
// TODO: move to module parameters.
const InterchainTxTimeout = time.Hour * 24 * 7

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k Keeper) RegisterInterchainAccount(goCtx context.Context, msg *ictxtypes.MsgRegisterInterchainAccount) (*ictxtypes.MsgRegisterInterchainAccountResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelRegisterInterchainAccount)
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("RegisterInterchainAccount", "connection_id", msg.ConnectionId, "from_address", msg.FromAddress, "interchain_accountt_id", msg.InterchainAccountId)

	icaOwner, err := types.NewICAOwner(msg.FromAddress, msg.InterchainAccountId)
	if err != nil {
		k.Logger(ctx).Error("RegisterInterchainAccount: failed to create RegisterInterchainAccount", "error", err)
		return nil, sdkerrors.Wrap(err, "failed to create ICA owner")
	}

	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, icaOwner.String()); err != nil {
		k.Logger(ctx).Debug("RegisterInterchainAccount: failed to create RegisterInterchainAccount:", "error", err, "owner", icaOwner.String(), "msg", &msg)
		return nil, sdkerrors.Wrap(err, "failed to RegisterInterchainAccount")
	}

	return &ictxtypes.MsgRegisterInterchainAccountResponse{}, nil
}

func (k Keeper) SubmitTx(goCtx context.Context, msg *ictxtypes.MsgSubmitTx) (*ictxtypes.MsgSubmitTxResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelSubmitTx)
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("Submit tx", "connection_id", msg.ConnectionId, "from_address", msg.FromAddress, "interchain_accountt_id", msg.InterchainAccountId)

	icaOwner, err := types.NewICAOwner(msg.FromAddress, msg.InterchainAccountId)
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

	data, err := icatypes.SerializeCosmosTx(k.cdc, sdkMsgs)
	if err != nil {
		k.Logger(ctx).Debug("SubmitTx failed: failed to SerializeCosmosTx", "error", err, "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(err, "failed to SerializeCosmosTx")
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: msg.Memo,
	}

	timeoutTimestamp := time.Now().Add(InterchainTxTimeout).UnixNano()
	_, err = k.icaControllerKeeper.SendTx(ctx, chanCap, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		k.Logger(ctx).Error("SubmitTx failed", "error", err, "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(err, "failed to SendTx")
	}

	return &types.MsgSubmitTxResponse{}, nil
}
