package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

// HandleAcknowledgement passes the acknowledgement data to the Hub contract via a Sudo call.
func (k *Keeper) HandleAcknowledgement(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) error {
	hubContractAddress, err := k.GetHubAddress(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to GetHubAddress", err)
		return sdkerrors.Wrap(err, "failed to GetHubAddress")
	}

	acknowledgementPacketBz, err := packet.Marshal()
	if err != nil {
		k.Logger(ctx).Error("failed to marshal acknowledgement packet", "error", err, "sequence", packet.Sequence, "dst_channel",
			packet.DestinationChannel, "dst_port", packet.DestinationPort, "height", packet.TimeoutHeight)
		return sdkerrors.Wrap(err, "failed to marshal Timeout packet")
	}

	// TODO: we need to pass both the marshaled packet & the acknowledgement bytes in a single message,
	//  and it should be easy for the contract to parse it. We should possibly just use JSON for that.
	_, err = k.Sudo(ctx, hubContractAddress, append(acknowledgementPacketBz, acknowledgement...))
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

	timeoutPacketBz, err := packet.Marshal()
	if err != nil {
		k.Logger(ctx).Error("failed to marshal timeout packet", "error", err, "sequence", packet.Sequence, "dst_channel",
			packet.DestinationChannel, "dst_port", packet.DestinationPort, "height", packet.TimeoutHeight)
		return sdkerrors.Wrap(err, "failed to marshal Timeout packet")
	}

	_, err = k.Sudo(ctx, hubContractAddress, timeoutPacketBz)
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
	_, err = k.Sudo(ctx, hubContractAddress, []byte(portID+channelID+counterPartyChannelId+counterpartyVersion))
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
func (k *Keeper) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	return nil, nil
}
