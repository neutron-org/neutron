package keeper

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

type SudoMessageTimeout struct {
	Timeout struct {
		Request           channeltypes.Packet `json:"request"`
		RequestPacketData RequestPacketData   `json:"request_packet_data"`
	} `json:"timeout"`
}
type SudoMessageResponse struct {
	Response struct {
		Request           channeltypes.Packet `json:"request"`
		RequestPacketData RequestPacketData   `json:"request_packet_data"`
		Message           []byte              `json:"data"`
	} `json:"response"`
}

type SudoMessageError struct {
	Error struct {
		Request           channeltypes.Packet `json:"request"`
		RequestPacketData RequestPacketData   `json:"request_packet_data"`
		Details           string              `json:"details"`
	} `json:"error"`
}

type RequestPacketData struct {
	Data      []byte `json:"data,omitempty"`
	Memo      string `json:"memo,omitempty"`
	Operation string `json:"operation,omitempty"`
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

	var packetData icatypes.InterchainAccountPacketData
	err = icatypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &packetData)
	if err != nil {
		k.Logger(ctx).Error("failed to unmarshal InterchainAccountPacketData", "error", err)
		return sdkerrors.Wrap(err, "failed to unmarshal InterchainAccountPacketData")
	}
	k.Logger(ctx).Info("Received PacketData", "packetData", packetData)

	// Actually we have only one kind of error returned from acknowledgement
	// maybe later we'll retrieve actual errors from events
	errorText := ack.GetError()
	if errorText != "" {
		_, err = k.SudoError(ctx, hubContractAddress, packet, packetData, errorText)
	} else {
		k.Logger(ctx).Info("HandleAcknowledgement", "ack", ack.String())
		_, err = k.SudoResponse(ctx, hubContractAddress, packet, packetData, ack.GetResult())
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
	counterpartyChannelId,
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
		CounterpartyChannelId: counterpartyChannelId,
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
	packetData icatypes.InterchainAccountPacketData,
	msg []byte,
) ([]byte, error) {
	k.Logger(ctx).Info("SudoResponse", "contractAddress", contractAddress, "request", request, "msg", msg)

	// basically just for unit tests right now. But i think we will have the same logic in the production
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	
	operation, memo, err := unpackMemo(packetData.Memo)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unpack memo from response")
	}

	x := SudoMessageResponse{}
	x.Response.Message = msg
	x.Response.Request = request
	x.Response.RequestPacketData = RequestPacketData{Data: packetData.Data, Memo: memo, Operation: operation}
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

	var packetData icatypes.InterchainAccountPacketData
	err := icatypes.ModuleCdc.UnmarshalJSON(request.GetData(), &packetData)
	if err != nil {
		k.Logger(ctx).Error("failed to unmarshal InterchainAccountPacketData", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to unmarshal InterchainAccountPacketData")
	}
	k.Logger(ctx).Info("Received PacketData", "packetData", packetData)

	operation, memo, err := unpackMemo(packetData.Memo)
	if err != nil {
		// TODO: do we need to return if cannot unpack memo?
		return nil, sdkerrors.Wrap(err, "failed to unpack memo from response")
	}

	x := SudoMessageTimeout{}
	x.Timeout.Request = request
	x.Timeout.RequestPacketData = RequestPacketData{Data: packetData.Data, Memo: memo, Operation: operation}
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
	packetData icatypes.InterchainAccountPacketData,
	details string,
) ([]byte, error) {
	k.Logger(ctx).Info("SudoError", "contractAddress", contractAddress, "request", request)

	// basically just for unit tests right now. But i think we will have the same logic in the production
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	operation, memo, err := unpackMemo(packetData.Memo)
	if err != nil {
		// TODO: do we need to return if cannot unpack memo?
		return nil, sdkerrors.Wrap(err, "failed to unpack memo from response")
	}

	x := SudoMessageError{}
	x.Error.Request = request
	x.Error.Details = details
	x.Error.RequestPacketData = RequestPacketData{Data: packetData.Data, Memo: memo, Operation: operation}
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
