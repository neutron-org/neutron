package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	feetypes "github.com/neutron-org/neutron/v3/x/feerefunder/types"
	ictxtypes "github.com/neutron-org/neutron/v3/x/interchaintxs/types"
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
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.FromAddress)
	}

	if !k.sudoKeeper.HasContractInfo(ctx, senderAddr) {
		k.Logger(ctx).Debug("RegisterInterchainAccount: contract not found", "from_address", msg.FromAddress)
		return nil, errors.Wrapf(ictxtypes.ErrNotContract, "%s is not a contract address", msg.FromAddress)
	}

	// if contract is stored before [last] upgrade, we're not going charge fees for register ICA
	if k.sudoKeeper.GetContractInfo(ctx, senderAddr).CodeID >= k.GetICARegistrationFeeFirstCodeID(ctx) {
		if err := k.ChargeFee(ctx, senderAddr, msg.RegisterFee); err != nil {
			return nil, errors.Wrapf(err, "failed to charge fees to pay for RegisterInterchainAccount msg: %s", msg)
		}
	}

	icaOwner := ictxtypes.NewICAOwnerFromAddress(senderAddr, msg.InterchainAccountId)

	// FIXME: empty version string doesn't look good
	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, icaOwner.String(), ""); err != nil {
		k.Logger(ctx).Debug("RegisterInterchainAccount: failed to create RegisterInterchainAccount:", "error", err, "owner", icaOwner.String(), "msg", &msg)
		return nil, errors.Wrap(err, "failed to RegisterInterchainAccount")
	}

	return &ictxtypes.MsgRegisterInterchainAccountResponse{}, nil
}

func (k Keeper) SubmitTx(goCtx context.Context, msg *ictxtypes.MsgSubmitTx) (*ictxtypes.MsgSubmitTxResponse, error) {
	defer telemetry.ModuleMeasureSince(ictxtypes.ModuleName, time.Now(), LabelSubmitTx)

	if msg == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "nil msg is prohibited")
	}

	if msg.Msgs == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "empty Msgs field is prohibited")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("SubmitTx", "connection_id", msg.ConnectionId, "from_address", msg.FromAddress, "interchain_account_id", msg.InterchainAccountId)

	senderAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		k.Logger(ctx).Debug("SubmitTx: failed to parse sender address", "from_address", msg.FromAddress)
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.FromAddress)
	}

	if !k.sudoKeeper.HasContractInfo(ctx, senderAddr) {
		k.Logger(ctx).Debug("SubmitTx: contract not found", "from_address", msg.FromAddress)
		return nil, errors.Wrapf(ictxtypes.ErrNotContract, "%s is not a contract address", msg.FromAddress)
	}

	params := k.GetParams(ctx)
	if uint64(len(msg.Msgs)) > params.GetMsgSubmitTxMaxMessages() {
		k.Logger(ctx).Debug("SubmitTx: provided MsgSubmitTx contains more messages than allowed",
			"msg", msg,
			"has", len(msg.Msgs),
			"max", params.GetMsgSubmitTxMaxMessages(),
		)
		return nil, fmt.Errorf(
			"MsgSubmitTx contains more messages than allowed, has=%d, max=%d",
			len(msg.Msgs),
			params.GetMsgSubmitTxMaxMessages(),
		)
	}

	icaOwner := ictxtypes.NewICAOwnerFromAddress(senderAddr, msg.InterchainAccountId)

	portID, err := icatypes.NewControllerPortID(icaOwner.String())
	if err != nil {
		k.Logger(ctx).Error("SubmitTx: failed to create NewControllerPortID:", "error", err, "owner", icaOwner)
		return nil, errors.Wrap(err, "failed to create NewControllerPortID")
	}

	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, msg.ConnectionId, portID)
	if !found {
		k.Logger(ctx).Debug("SubmitTx: failed to GetActiveChannelID", "connection_id", msg.ConnectionId, "port_id", portID)
		return nil, errors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to GetActiveChannelID for port %s", portID)
	}

	data, err := SerializeCosmosTx(k.Codec, msg.Msgs)
	if err != nil {
		k.Logger(ctx).Debug("SubmitTx: failed to SerializeCosmosTx", "error", err, "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, errors.Wrap(err, "failed to SerializeCosmosTx")
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: msg.Memo,
	}

	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, portID, channelID)
	if !found {
		return nil, errors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", portID, channelID,
		)
	}

	if err := k.feeKeeper.LockFees(ctx, senderAddr, feetypes.NewPacketID(portID, channelID, sequence), msg.Fee); err != nil {
		return nil, errors.Wrapf(err, "failed to lock fees to pay for SubmitTx msg: %s", msg)
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Duration(msg.Timeout) * time.Second).UnixNano()
	// TODO: keeper's SendTx deprecated, replace it with MsgServer SendTx
	_, err = k.icaControllerKeeper.SendTx(ctx, nil, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		// usually we use DEBUG level for such errors, but in this case we have checked full input before running SendTX, so error here may be critical
		k.Logger(ctx).Error("SubmitTx", "error", err, "connection_id", msg.ConnectionId, "port_id", portID, "channel_id", channelID)
		return nil, errors.Wrap(err, "failed to SendTx")
	}

	return &ictxtypes.MsgSubmitTxResponse{
		SequenceId: sequence,
		Channel:    channelID,
	}, nil
}

// SerializeCosmosTx serializes a slice of *types.Any messages using the CosmosTx type. The proto marshaled CosmosTx
// bytes are returned. This differs from icatypes.SerializeCosmosTx in that it does not serialize sdk.Msgs, but
// simply uses the already serialized values.
func SerializeCosmosTx(cdc codec.BinaryCodec, msgs []*codectypes.Any) (bz []byte, err error) {
	// only ProtoCodec is supported
	if _, ok := cdc.(*codec.ProtoCodec); !ok {
		return nil, errors.Wrap(icatypes.ErrInvalidCodec,
			"only ProtoCodec is supported for receiving messages on the host chain")
	}

	cosmosTx := &icatypes.CosmosTx{
		Messages: msgs,
	}

	bz, err = cdc.Marshal(cosmosTx)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// UpdateParams updates the module parameters
func (k Keeper) UpdateParams(goCtx context.Context, req *ictxtypes.MsgUpdateParams) (*ictxtypes.MsgUpdateParamsResponse, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, err
	}
	authority := k.GetAuthority()
	if authority != req.Authority {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &ictxtypes.MsgUpdateParamsResponse{}, nil
}
