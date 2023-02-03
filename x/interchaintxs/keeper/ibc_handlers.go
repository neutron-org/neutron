package keeper

import (
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

const (
	// GasReserve is the amount of gas on the context gas meter we need to reserve in order to add contract failure to keeper
	GasReserve = 15000
)

func (k *Keeper) outOfGasRecovery(
	ctx sdk.Context,
	gasMeter sdk.GasMeter,
	senderAddress sdk.AccAddress,
	packet channeltypes.Packet,
	filureAckType string,
) {
	if r := recover(); r != nil {
		_, ok := r.(sdk.ErrorOutOfGas)
		if !ok || !gasMeter.IsOutOfGas() {
			panic(r)
		}

		k.Logger(ctx).Debug("Out of gas", "Gas meter", gasMeter.String())
		k.contractManagerKeeper.AddContractFailure(ctx, packet.SourceChannel, senderAddress.String(), packet.GetSequence(), filureAckType)
		// FIXME: add distribution call
	}
}

func (k *Keeper) createCachedContext(ctx sdk.Context) (cacheCtx sdk.Context, writeFn func(), newGasMeter sdk.GasMeter) {
	gasLeft := ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()

	var newLimit uint64
	if gasLeft < GasReserve {
		newLimit = 0
	} else {
		newLimit = gasLeft - GasReserve
	}

	newGasMeter = sdk.NewGasMeter(newLimit)

	cacheCtx, writeFn = ctx.CacheContext()
	if strings.HasPrefix(ctx.GasMeter().String(), "BasicGasMeter") {
		cacheCtx = ctx.WithGasMeter(newGasMeter)
	}

	return
}

// HandleAcknowledgement passes the acknowledgement data to the appropriate contract via a Sudo call.
func (k *Keeper) HandleAcknowledgement(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelHandleAcknowledgment)

	k.Logger(ctx).Debug("Handling acknowledgement")
	icaOwner, err := types.ICAOwnerFromPort(packet.SourcePort)
	if err != nil {
		k.Logger(ctx).Error("HandleAcknowledgement: failed to get ica owner from source port", "error", err)
		return sdkerrors.Wrap(err, "failed to get ica owner from port")
	}

	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		k.Logger(ctx).Error("HandleAcknowledgement: cannot unmarshal ICS-27 packet acknowledgement", "error", err)
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 packet acknowledgement: %v", err)
	}

	cacheCtx, writeFn, newGasMeter := k.createCachedContext(ctx)
	defer k.outOfGasRecovery(ctx, newGasMeter, icaOwner.GetContract(), packet, "ack")

	// Actually we have only one kind of error returned from acknowledgement
	// maybe later we'll retrieve actual errors from events
	errorText := ack.GetError()
	if errorText != "" {
		_, err = k.contractManagerKeeper.SudoError(cacheCtx, icaOwner.GetContract(), packet, errorText)
	} else {
		_, err = k.contractManagerKeeper.SudoResponse(cacheCtx, icaOwner.GetContract(), packet, ack.GetResult())
	}

	k.feeKeeper.DistributeAcknowledgementFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))

	if err != nil {
		k.contractManagerKeeper.AddContractFailure(ctx, packet.SourceChannel, icaOwner.GetContract().String(), packet.GetSequence(), "ack")
		k.Logger(ctx).Debug("HandleAcknowledgement: failed to Sudo contract on packet acknowledgement", "error", err)
	} else {
		writeFn()
	}

	ctx.GasMeter().ConsumeGas(newGasMeter.GasConsumed(), "consume from cached context")

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a Sudo call.
// Since all ICA channels are ORDERED, a single timeout shuts down a channel.
// The affected zone should be paused after a timeout.
func (k *Keeper) HandleTimeout(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelHandleTimeout)

	icaOwner, err := types.ICAOwnerFromPort(packet.SourcePort)
	k.Logger(ctx).Debug("HandleTimeout")
	if err != nil {
		k.Logger(ctx).Error("HandleTimeout: failed to get ica owner from source port", "error", err)
		return sdkerrors.Wrap(err, "failed to get ica owner from port")
	}

	cacheCtx, writeFn, newGasMeter := k.createCachedContext(ctx)
	defer k.outOfGasRecovery(ctx, newGasMeter, icaOwner.GetContract(), packet, "timeout")

	_, err = k.contractManagerKeeper.SudoTimeout(cacheCtx, icaOwner.GetContract(), packet)
	if err != nil {
		k.contractManagerKeeper.AddContractFailure(ctx, packet.SourceChannel, icaOwner.GetContract().String(), packet.GetSequence(), "timeout")
		k.Logger(ctx).Error("HandleTimeout: failed to Sudo contract on packet timeout", "error", err)
	} else {
		writeFn()
	}

	k.feeKeeper.DistributeTimeoutFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))
	ctx.GasMeter().ConsumeGas(newGasMeter.GasConsumed(), "consume from cached context")

	return nil
}

// HandleChanOpenAck passes the data about a successfully created channel to the appropriate contract
// (== the data about a successfully registered interchain account).
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
		return sdkerrors.Wrap(err, "failed to get ica owner from port")
	}

	_, err = k.contractManagerKeeper.SudoOnChanOpenAck(ctx, icaOwner.GetContract(), contractmanagertypes.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   counterpartyVersion,
	})
	if err != nil {
		k.Logger(ctx).Debug("HandleChanOpenAck: failed to Sudo contract on packet timeout", "error", err)
		return sdkerrors.Wrap(err, "failed to Sudo the contract OnChanOpenAck")
	}

	return nil
}
