package keeper

import (
	"context"
	"time"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	"github.com/lidofinance/interchain-adapter/x/interchaintxs/types"
	proto "github.com/lidofinance/interchain-adapter/x/interchaintxs/types"
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

func (k Keeper) RegisterInterchainAccount(goCtx context.Context, msg *proto.MsgRegisterInterchainAccount) (*proto.MsgRegisterInterchainAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.Owner); err != nil {
		k.Logger(ctx).Error("failed to create RegisterInterchainAccount:", "error", err, "owner", msg.Owner, "connection_id", msg.ConnectionId)
		return nil, sdkerrors.Wrap(err, "failed to RegisterInterchainAccount")
	}

	return nil, nil
}

func (k Keeper) SubmitTx(goCtx context.Context, msg *proto.MsgSubmitTx) (*proto.MsgSubmitTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	portID, err := icatypes.NewControllerPortID(msg.Owner)
	if err != nil {
		k.Logger(ctx).Error("failed to create NewControllerPortID:", "error", err, "owner", msg.Owner)
		return nil, sdkerrors.Wrap(err, "failed to create NewControllerPortID")
	}

	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, msg.ConnectionId, portID)
	if !found {
		k.Logger(ctx).Error("failed to GetActiveChannelID", "connection_id", msg.ConnectionId, "port_id", portID)
		return nil, sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to GetActiveChannelID for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		k.Logger(ctx).Error("failed to GetCapability", "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "failed to GetCapability")
	}

	sdkMsgs, err := msg.GetTxMsgs()
	if err != nil {
		k.Logger(ctx).Error("failed to GetTxMsgs", "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(err, "failed to GetTxMsgs")
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, sdkMsgs)
	if err != nil {
		k.Logger(ctx).Error("failed to SerializeCosmosTx", "error", err, "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(err, "failed to SerializeCosmosTx")
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	timeoutTimestamp := time.Now().Add(InterchainTxTimeout).UnixNano()
	_, err = k.icaControllerKeeper.SendTx(ctx, chanCap, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		k.Logger(ctx).Error("failed to SendTx", "error", err, "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, sdkerrors.Wrap(err, "failed to SendTx")
	}

	return &types.MsgSubmitTxResponse{}, nil
}
