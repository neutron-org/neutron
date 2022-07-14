package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
	"github.com/neutron-org/neutron/x/sudo"
)

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
		_, err = k.sudoHandler.SudoError(ctx, icaOwner.GetContract(), packet, errorText)
	} else {
		_, err = k.sudoHandler.SudoResponse(ctx, icaOwner.GetContract(), packet, ack.GetResult())
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

	_, err = k.sudoHandler.SudoTimeout(ctx, icaOwner.GetContract(), packet)
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

	_, err = k.sudoHandler.SudoOpenAck(ctx, icaOwner.GetContract(), sudo.OpenAckDetails{
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
