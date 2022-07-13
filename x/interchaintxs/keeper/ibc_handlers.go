package keeper

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
)

type SudoMessageTimeout struct {
	Timeout struct {
		Request channeltypes.Packet `json:"request"`
	} `json:"timeout"`
}
type SudoMessageResponse struct {
	Response struct {
		Request channeltypes.Packet `json:"request"`
		Message []byte              `json:"data"`
	} `json:"response"`
}

type SudoMessageError struct {
	Error struct {
		Request channeltypes.Packet `json:"request"`
		Details string              `json:"details"`
	} `json:"error"`
}

type SudoMessageOpenAck struct {
	OpenAck OpenAckDetails `json:"open_ack"`
}

type OpenAckDetails struct {
	PortID                string `json:"port_id"`
	ChannelID             string `json:"channel_id"`
	CounterpartyChannelId string `json:"counterparty_channel_id"`
	CounterpartyVersion   string `json:"counterparty_version"`
}

// HandleAcknowledgement passes the acknowledgement data to the appropriate contract via a Sudo call.
func (k *Keeper) HandleAcknowledgement(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) error {
	icaOwner, err := types.ICAOwnerFromPort(packet.SourcePort)
	if err != nil {
		k.Logger(ctx).Error("failed to get ica owner from source port: %v", err)
		return sdkerrors.Wrap(err, "failed to get ica owner from port")
	}

	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 packet acknowledgement: %v", err)
	}

	// Actually we have only one kind of error returned from acknowledgement
	// maybe later we'll retrieve actual errors from events
	errorText := ack.GetError()
	if errorText != "" {
		_, err = k.SudoError(ctx, icaOwner.GetContract(), packet, errorText)
	} else {
		_, err = k.SudoResponse(ctx, icaOwner.GetContract(), packet, ack.GetResult())
	}

	if err != nil {
		k.Logger(ctx).Error("failed to Sudo contract on packet acknowledgement", err)
		return sdkerrors.Wrap(err, "failed to Sudo the contract on packet acknowledgement")
	}

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a Sudo call.
// Since all ICA channels are ORDERED, a single timeout shuts down a channel.
// The affected zone should be paused after a timeout.
func (k *Keeper) HandleTimeout(ctx sdk.Context, packet channeltypes.Packet) error {
	icaOwner, err := types.ICAOwnerFromPort(packet.SourcePort)
	if err != nil {
		k.Logger(ctx).Error("failed to get ica owner from source port: %v", err)
		return sdkerrors.Wrap(err, "failed to get ica owner from port")
	}

	_, err = k.SudoTimeout(ctx, icaOwner.GetContract(), packet)
	if err != nil {
		k.Logger(ctx).Error("failed to Sudo contract on packet timeout", err)
		return sdkerrors.Wrap(err, "failed to Sudo the contract on packet timeout")
	}

	return nil
}

// HandleChanOpenAck passes the data about a successfully created channel to the appropriate contract
// (== the data about a successfully registered interchain account).
func (k *Keeper) HandleChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterpartyChannelId,
	counterpartyVersion string,
) error {
	icaOwner, err := types.ICAOwnerFromPort(portID)
	if err != nil {
		k.Logger(ctx).Error("failed to get ica owner from source port: %v", err)
		return sdkerrors.Wrap(err, "failed to get ica owner from port")
	}

	_, err = k.SudoOpenAck(ctx, icaOwner.GetContract(), OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelId: counterpartyChannelId,
		CounterpartyVersion:   counterpartyVersion,
	})
	if err != nil {
		k.Logger(ctx).Error("failed to Sudo the contract on packet openAck", err)
		return sdkerrors.Wrap(err, "failed to Sudo the contract on packet openAck")
	}

	return nil
}

// Sudo allows privileged access to a contract. This can never be called by an external tx, but only by
// another native Go module directly, or on-chain governance (if sudo proposals are enabled). Thus, the keeper doesn't
// place any access controls on it, that is the responsibility or the app developer (who passes the wasm.Keeper in
// app.go).
//
func (k *Keeper) SudoResponse(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
	msg []byte,
) ([]byte, error) {
	k.Logger(ctx).Info("SudoResponse", "contractAddress", contractAddress, "request", request, "msg", msg)

	// basically just for unit tests right now. But i think we will have the same logic in the production
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	x := SudoMessageResponse{}
	x.Response.Message = msg
	x.Response.Request = request
	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal sudo message", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessage")
	}
	k.Logger(ctx).Info("SudoResponse sending request", "data", string(m))

	r, err := k.wasmKeeper.Sudo(ctx, contractAddress, m)
	k.Logger(ctx).Info("SudoResponse received response", "err", err, "response", string(r))

	return r, err
}

func (k *Keeper) SudoTimeout(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
) ([]byte, error) {
	k.Logger(ctx).Info("SudoTimeout", "contractAddress", contractAddress, "request", request)

	// basically just for unit tests right now. But i think we will have the same logic in the production
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	x := SudoMessageTimeout{}
	x.Timeout.Request = request
	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal sudo message", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessage")
	}
	k.Logger(ctx).Info("SudoTimeout sending request", "data", string(m))

	r, err := k.wasmKeeper.Sudo(ctx, contractAddress, m)
	k.Logger(ctx).Info("SudoTimeout received response", "err", err, "response", string(r))

	return r, err
}

func (k *Keeper) SudoError(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
	details string,
) ([]byte, error) {
	k.Logger(ctx).Info("SudoError", "contractAddress", contractAddress, "request", request)

	// basically just for unit tests right now. But i think we will have the same logic in the production
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	x := SudoMessageError{}
	x.Error.Request = request
	x.Error.Details = details
	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal sudo message", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessage")
	}
	k.Logger(ctx).Info("SudoError sending request", "data", string(m))

	r, err := k.wasmKeeper.Sudo(ctx, contractAddress, m)
	k.Logger(ctx).Info("SudoError received response", "err", err, "response", string(r))

	return r, err
}

func (k *Keeper) SudoOpenAck(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	details OpenAckDetails,
) ([]byte, error) {
	k.Logger(ctx).Info("SudoOpenAck", "contractAddress", contractAddress)

	// basically just for unit tests right now. But i think we will have the same logic in the production
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	x := SudoMessageOpenAck{}
	x.OpenAck = details
	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal sudo message", "error", err)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessage")
	}
	k.Logger(ctx).Info("SudoOpenAck sending request", "data", string(m))

	r, err := k.wasmKeeper.Sudo(ctx, contractAddress, m)
	k.Logger(ctx).Info("SudoOpenAck received response", "err", err, "response", string(r))

	return r, err
}
