package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

type SudoMessageType uint

const (
	SudoMessageTypeOpenAck SudoMessageType = iota
	SudoMessageTypeResponse
	SudoMessageTypeTimeout
	SudoMessageTypeError
)

type SudoMessage struct {
	MessageType SudoMessageType     `json:"type"`
	Request     channeltypes.Packet `json:"request,omitempty"`
	Message     string              `json:"result,omitempty"`
	Error       string              `json:"error,omitempty"`
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
		_, err = k.Sudo(ctx, SudoMessageTypeError, hubContractAddress, packet, []byte{}, []byte(errorText))
	} else {
		_, err = k.Sudo(ctx, SudoMessageTypeResponse, hubContractAddress, packet, ack.GetResult(), []byte{})
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

	_, err = k.Sudo(ctx, SudoMessageTypeTimeout, hubContractAddress, packet, []byte{}, []byte{})
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

	// TODO: we need to pass both the marshaled arguments bytes in a single message,
	//  and it should be easy for the contract to parse it. I don't want to use JSON (it's super ugly
	//  in this context); maybe we should generate a separate proto-message that will reference the
	//  channeltypes.Packet?
	_, err = k.Sudo(ctx, SudoMessageTypeOpenAck, hubContractAddress, channeltypes.Packet{}, []byte(portID+channelID+counterPartyChannelId+counterpartyVersion), []byte{})
	if err != nil {
		k.Logger(ctx).Error("failed to Sudo the hub contract on packet timeout", err)
		return sdkerrors.Wrap(err, "failed to Sudo the hub contract on packet timeout")
	}

	return nil
}

// Sudo allows privileged access to a contract. This can never be called by an external tx, but only by
// another native Go module directly, or on-chain governance (if sudo proposals are enabled). Thus, the keeper doesn't
// place any access controls on it, that is the responsibility or the app developer (who passes the wasm.Keeper in
// app.go).
//
// TODO: Sudo is actually a part of the wasmd keeper. When cosmos/wasmd is finalized, we need
// 	to import it and use the original sudo call.
func (k *Keeper) Sudo(
	ctx sdk.Context,
	messageType SudoMessageType,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
	msg []byte,
	error []byte,
) ([]byte, error) {

	k.Logger(ctx).Info("Sudo", "contractAddress", contractAddress, "request", request, "msg", msg)

	m, err := json.Marshal(SudoMessage{
		MessageType: messageType,
		Request:     request,
		Message:     string(msg),
		Error:       string(error),
	})

	if err != nil {
		k.Logger(ctx).Error("failed to marshal sudo message", "error", err, "sequence", request.Sequence, "dst_channel",
			request.DestinationChannel, "dst_port", request.DestinationPort, "height", request.TimeoutHeight)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessage")
	}

	return k.wasmKeeper.Sudo(ctx, contractAddress, m)
}
