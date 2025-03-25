package keeper

import (
	"time"

	"github.com/neutron-org/neutron/v6/x/contractmanager/keeper"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	contractmanagertypes "github.com/neutron-org/neutron/v6/x/contractmanager/types"
	feetypes "github.com/neutron-org/neutron/v6/x/feerefunder/types"
	"github.com/neutron-org/neutron/v6/x/interchaintxs/types"
)

// HandleAcknowledgement passes the acknowledgement data to the appropriate contract via a sudo call.
func (k *Keeper) HandleAcknowledgement(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelHandleAcknowledgment)
	k.Logger(ctx).Debug("Handling acknowledgement")
	icaOwner, err := types.ICAOwnerFromPort(packet.SourcePort)
	if err != nil {
		k.Logger(ctx).Error("HandleAcknowledgement: failed to get ica owner from source port", "error", err)
		return errors.Wrap(err, "failed to get ica owner from port")
	}

	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		k.Logger(ctx).Error("HandleAcknowledgement: cannot unmarshal ICS-27 packet acknowledgement", "error", err)
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 packet acknowledgement: %v", err)
	}
	msg, err := keeper.PrepareSudoCallbackMessage(packet, &ack)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal Packet/Acknowledgment: %v", err)
	}

	k.feeKeeper.DistributeAcknowledgementFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))

	// Actually we have only one kind of error returned from acknowledgement
	// maybe later we'll retrieve actual errors from events
	_, err = k.sudoKeeper.Sudo(ctx, icaOwner.GetContract(), msg)
	if err != nil {
		k.Logger(ctx).Debug("HandleAcknowledgement: failed to Sudo contract on packet acknowledgement", "error", err)
	}

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a sudo call.
// Since all ICA channels are ORDERED, a single timeout shuts down a channel.
func (k *Keeper) HandleTimeout(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelHandleTimeout)
	k.Logger(ctx).Debug("HandleTimeout")
	icaOwner, err := types.ICAOwnerFromPort(packet.SourcePort)
	if err != nil {
		k.Logger(ctx).Error("HandleTimeout: failed to get ica owner from source port", "error", err)
		return errors.Wrap(err, "failed to get ica owner from port")
	}

	msg, err := keeper.PrepareSudoCallbackMessage(packet, nil)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal Packet: %v", err)
	}

	k.feeKeeper.DistributeTimeoutFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))

	_, err = k.sudoKeeper.Sudo(ctx, icaOwner.GetContract(), msg)
	if err != nil {
		k.Logger(ctx).Debug("HandleTimeout: failed to Sudo contract on packet timeout", "error", err)
	}

	return nil
}

// HandleChanOpenAck passes the data about a successfully created channel to the appropriate contract
// (== the data about a successfully registered interchain account).
// Notice that in the case of an ICA channel - it is not yet in OPEN state here
// the last step of channel opening(confirm) happens on the host chain.
func (k *Keeper) HandleChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterpartyChannelID,
	counterpartyVersion string,
) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelLabelHandleChanOpenAck)

	k.Logger(ctx).Debug("HandleChanOpenAck", "port_id", portID, "channel_id", channelID, "counterparty_channel_id", counterpartyChannelID, "counterparty_version", counterpartyVersion)
	icaOwner, err := types.ICAOwnerFromPort(portID)
	if err != nil {
		k.Logger(ctx).Error("HandleChanOpenAck: failed to get ica owner from source port", "error", err)
		return errors.Wrap(err, "failed to get ica owner from port")
	}

	payload, err := keeper.PrepareOpenAckCallbackMessage(contractmanagertypes.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   counterpartyVersion,
	})
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal OpenAckDetails: %v", err)
	}

	_, err = k.sudoKeeper.Sudo(ctx, icaOwner.GetContract(), payload)
	if err != nil {
		k.Logger(ctx).Debug("HandleChanOpenAck: failed to sudo contract on channel open acknowledgement", "error", err)
	}

	return nil
}
