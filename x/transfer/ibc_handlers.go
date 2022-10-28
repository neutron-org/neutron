package transfer

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

const (
	// We need to reserve this amount of gas on the context gas meter in order to add contract failure to keeper
	GasReserve = 15000
)

func (im IBCModule) outOfGasRecovery(
	ctx sdk.Context,
	gasMeter sdk.GasMeter,
	senderAddress sdk.AccAddress,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
) {
	if r := recover(); r != nil {
		_, ok := r.(sdk.ErrorOutOfGas)
		if !ok || !gasMeter.IsOutOfGas() {
			panic(r)
		}

		im.keeper.Logger(ctx).Debug("Out of gas", "Gas meter", gasMeter.String(), "Packet data", data)
		im.ContractmanagerKeeper.AddContractFailure(ctx, senderAddress.String(), packet.GetSequence(), "ack")

	}
}

func (im IBCModule) createCachedContext(ctx sdk.Context) (cacheCtx sdk.Context, writeFn func(), newGasMeter sdk.GasMeter) {
	gasLeft := ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()
	newLimit := gasLeft - GasReserve

	if newLimit > gasLeft {
		newLimit = 0
	}

	newGasMeter = sdk.NewGasMeter(newLimit)

	cacheCtx, writeFn = ctx.CacheContext()
	if strings.HasPrefix(ctx.GasMeter().String(), "BasicGasMeter") {
		cacheCtx = ctx.WithGasMeter(newGasMeter)
	}

	return
}

// HandleAcknowledgement passes the acknowledgement data to the appropriate contract via a Sudo call.
func (im IBCModule) HandleAcknowledgement(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) (err error) {
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}
	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	senderAddress, err := sdk.AccAddressFromBech32(data.GetSender())
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to decode address from bech32: %v", err)
	}

	cacheCtx, writeFn, newGasMeter := im.createCachedContext(ctx)
	defer im.outOfGasRecovery(ctx, newGasMeter, senderAddress, packet, data)

	if ack.Success() {
		_, err = im.ContractmanagerKeeper.SudoResponse(cacheCtx, senderAddress, packet, ack.GetResult())
	} else {
		// Actually we have only one kind of error returned from acknowledgement
		// maybe later we'll retrieve actual errors from events
		im.keeper.Logger(cacheCtx).Debug(ack.GetError(), "CheckTx", cacheCtx.IsCheckTx())
		_, err = im.ContractmanagerKeeper.SudoError(cacheCtx, senderAddress, packet, ack.GetError())
	}

	if err != nil {
		im.ContractmanagerKeeper.AddContractFailure(ctx, senderAddress.String(), packet.GetSequence(), "ack")
		im.keeper.Logger(ctx).Debug("failed to Sudo contract on packet acknowledgement", err)
	} else {
		writeFn()
	}

	ctx.GasMeter().ConsumeGas(newGasMeter.GasConsumed(), "consume from cached context")
	im.keeper.Logger(ctx).Debug("acknowledgement received", "Packet data", data, "CheckTx", ctx.IsCheckTx())

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a Sudo call.
// Since all ICA channels are ORDERED, a single timeout shuts down a channel.
// The affected zone should be paused after a timeout.
func (im IBCModule) HandleTimeout(ctx sdk.Context, packet channeltypes.Packet) error {
	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	senderAddress, err := sdk.AccAddressFromBech32(data.GetSender())
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to decode address from bech32: %v", err)
	}

	cacheCtx, writeFn, newGasMeter := im.createCachedContext(ctx)
	defer im.outOfGasRecovery(ctx, newGasMeter, senderAddress, packet, data)

	_, err = im.ContractmanagerKeeper.SudoTimeout(cacheCtx, senderAddress, packet)
	if err != nil {
		im.ContractmanagerKeeper.AddContractFailure(ctx, senderAddress.String(), packet.GetSequence(), "timeout")
		im.keeper.Logger(ctx).Debug("failed to Sudo contract on packet timeout", err)
	} else {
		writeFn()
	}

	ctx.GasMeter().ConsumeGas(newGasMeter.GasConsumed(), "consume from cached context")

	return nil
}
