package keeper

import (
	"github.com/neutron-org/neutron/neutronutils"
	neutronerrors "github.com/neutron-org/neutron/neutronutils/errors"
	"time"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

// HandleAcknowledgement passes the acknowledgement data to the appropriate contract via a Sudo call.
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

	k.feeKeeper.DistributeAcknowledgementFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))

	cacheCtx, writeFn := neutronutils.CreateCachedContext(ctx, k.contractManagerKeeper.GetParams(ctx).SudoCallGasLimit)
	func() {
		defer neutronerrors.OutOfGasRecovery(cacheCtx.GasMeter(), &err)
		// Actually we have only one kind of error returned from acknowledgement
		// maybe later we'll retrieve actual errors from events
		if ack.GetError() != "" {

			_, err = k.contractManagerKeeper.SudoError(cacheCtx, icaOwner.GetContract(), packet, ack.GetError())
		} else {
			_, err = k.contractManagerKeeper.SudoResponse(cacheCtx, icaOwner.GetContract(), packet, ack.GetResult())
		}
	}()
	if err != nil {
		// the contract either returned an error or panicked with `out of gas`
		k.contractManagerKeeper.AddContractFailure(ctx, &packet, icaOwner.GetContract().String(), contractmanagertypes.Ack, &ack)
		k.Logger(ctx).Debug("HandleAcknowledgement: failed to Sudo contract on packet acknowledgement", "error", err)
	} else {
		writeFn()
	}

	ctx.GasMeter().ConsumeGas(cacheCtx.GasMeter().GasConsumedToLimit(), "consume gas from cached context")

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a Sudo call.
// Since all ICA channels are ORDERED, a single timeout shuts down a channel.
func (k *Keeper) HandleTimeout(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelHandleTimeout)

	icaOwner, err := types.ICAOwnerFromPort(packet.SourcePort)
	k.Logger(ctx).Debug("HandleTimeout")
	if err != nil {
		k.Logger(ctx).Error("HandleTimeout: failed to get ica owner from source port", "error", err)
		return errors.Wrap(err, "failed to get ica owner from port")
	}

	k.feeKeeper.DistributeTimeoutFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))

	cacheCtx, writeFn := neutronutils.CreateCachedContext(ctx, k.contractManagerKeeper.GetParams(ctx).SudoCallGasLimit)
	func() {
		defer neutronerrors.OutOfGasRecovery(cacheCtx.GasMeter(), &err)
		_, err = k.contractManagerKeeper.SudoTimeout(cacheCtx, icaOwner.GetContract(), packet)
	}()
	if err != nil {
		// the contract either returned an error or panicked with `out of gas`
		k.contractManagerKeeper.AddContractFailure(ctx, &packet, icaOwner.GetContract().String(), contractmanagertypes.Timeout, nil)
		k.Logger(ctx).Error("HandleTimeout: failed to Sudo contract on packet timeout", "error", err)
	} else {
		writeFn()
	}

	ctx.GasMeter().ConsumeGas(cacheCtx.GasMeter().GasConsumedToLimit(), "consume gas from cached context")

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

	_, err = k.contractManagerKeeper.SudoOnChanOpenAck(ctx, icaOwner.GetContract(), contractmanagertypes.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   counterpartyVersion,
	})
	if err != nil {
		k.Logger(ctx).Debug("HandleChanOpenAck: failed to Sudo contract on packet timeout", "error", err)
		return errors.Wrap(err, "failed to Sudo the contract OnChanOpenAck")
	}

	return nil
}
