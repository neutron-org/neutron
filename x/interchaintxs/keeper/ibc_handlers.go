package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
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
	CounterPartyChannelId string `json:"counter_party_channel_id"`
	CounterpartyVersion   string `json:"counterparty_version"`
}

// HandleAcknowledgement passes the acknowledgement data to the Hub contract via a Sudo call.
func (k *Keeper) HandleAcknowledgement(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) error {
	hubContractAddress, err := k.GetHubAddress(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to GetHubAddress", err)
		return sdkerrors.Wrap(err, "failed to GetHubAddress")
	}

	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 packet acknowledgement: %v", err)
	}

	// Actually we have only one kind of error returned from acknowledgement
	// maybe later we'll retrieve actual errors from events
	errorText := ack.GetError()
	if errorText != "" {
		_, err = k.SudoError(ctx, hubContractAddress, packet, errorText)
	} else {
		_, err = k.SudoResponse(ctx, hubContractAddress, packet, ack.GetResult())
	}

	if err != nil {
		k.Logger(ctx).Error("failed to Sudo the hub contract on packet acknowledgement", err)
		return sdkerrors.Wrap(err, "failed to Sudo the hub contract on packet acknowledgement")
	}

	return nil
}

// HandleTimeout passes the timeout data to the Hub contract via a Sudo call.
// Since all ICA channels are ORDERED, a single timeout shuts down a channel.
// The affected zone should be paused after a timeout.
func (k *Keeper) HandleTimeout(ctx sdk.Context, packet channeltypes.Packet) error {
	hubContractAddress, err := k.GetHubAddress(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to GetHubAddress", err)
		return sdkerrors.Wrap(err, "failed to GetHubAddress")
	}

	_, err = k.SudoTimeout(ctx, hubContractAddress, packet)
	if err != nil {
		k.Logger(ctx).Error("failed to Sudo the hub contract on packet timeout", err)
		return sdkerrors.Wrap(err, "failed to Sudo the hub contract on packet timeout")
	}

	return nil
}

// HandleChanOpenAck passes the data about a successfully created channel to the Hub contract
// (== the data about a successfully registered interchain account).
func (k *Keeper) HandleChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterPartyChannelId,
	counterpartyVersion string,
) error {
	hubContractAddress, err := k.GetHubAddress(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to GetHubAddress", err)
		return sdkerrors.Wrap(err, "failed to GetHubAddress")
	}

	_, err = k.SudoOpenAck(ctx, hubContractAddress, OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterPartyChannelId: counterPartyChannelId,
		CounterpartyVersion:   counterpartyVersion,
	})
	if err != nil {
		k.Logger(ctx).Error("failed to Sudo the hub contract on packet openAck", err)
		return sdkerrors.Wrap(err, "failed to Sudo the hub contract on packet openAck")
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
