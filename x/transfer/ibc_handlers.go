package transfer

import (
	"cosmossdk.io/errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

// HandleAcknowledgement passes the acknowledgement data to the appropriate contract via a Sudo call.
func (im IBCModule) HandleAcknowledgement(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}
	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	senderAddress, err := sdk.AccAddressFromBech32(data.GetSender())
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to decode address from bech32: %v", err)
	}
	if !im.ContractManagerKeeper.HasContractInfo(ctx, senderAddress) {
		return nil
	}

	cacheCtx, writeFn, newGasMeter := im.createCachedContext(ctx)
	// consume all the gas from the cached context
	// we call the function this place function because we want to consume all the gas even in case of panic in SudoResponse/SudoError
	ctx.GasMeter().ConsumeGas(newGasMeter.Limit(), "consume full gas from cached context")
	defer im.outOfGasRecovery(ctx, newGasMeter, senderAddress, packet, data, "ack")

	// distribute fee
	im.wrappedKeeper.FeeKeeper.DistributeAcknowledgementFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))
	
	if ack.Success() {
		_, err = im.ContractManagerKeeper.SudoResponse(cacheCtx, senderAddress, packet, ack.GetResult())
	} else {
		// Actually we have only one kind of error returned from acknowledgement
		// maybe later we'll retrieve actual errors from events
		im.keeper.Logger(cacheCtx).Debug(ack.GetError(), "CheckTx", cacheCtx.IsCheckTx())
		_, err = im.ContractManagerKeeper.SudoError(cacheCtx, senderAddress, packet, ack.GetError())
	}

	if err != nil {
		im.ContractManagerKeeper.AddContractFailure(ctx, packet.SourceChannel, senderAddress.String(), packet.GetSequence(), "ack")
		im.keeper.Logger(ctx).Debug("failed to Sudo contract on packet acknowledgement", err)
	} else {
		writeFn()
	}

	im.keeper.Logger(ctx).Debug("acknowledgement received", "Packet data", data, "CheckTx", ctx.IsCheckTx())

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a Sudo call.
// Since all ICA channels are ORDERED, a single timeout shuts down a channel.
// The affected zone should be paused after a timeout.
func (im IBCModule) HandleTimeout(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	senderAddress, err := sdk.AccAddressFromBech32(data.GetSender())
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to decode address from bech32: %v", err)
	}
	if !im.ContractManagerKeeper.HasContractInfo(ctx, senderAddress) {
		return nil
	}

	cacheCtx, writeFn, newGasMeter := im.createCachedContext(ctx)
	// consume all the gas from the cached context
	// we call the function this place function because we want to consume all the gas even in case of panic in SudoTimeout
	ctx.GasMeter().ConsumeGas(newGasMeter.Limit(), "consume full gas from cached context")
	defer im.outOfGasRecovery(ctx, newGasMeter, senderAddress, packet, data, "timeout")

	// distribute fee
	im.wrappedKeeper.FeeKeeper.DistributeTimeoutFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))

	_, err = im.ContractManagerKeeper.SudoTimeout(cacheCtx, senderAddress, packet)
	if err != nil {
		im.ContractManagerKeeper.AddContractFailure(ctx, packet.SourceChannel, senderAddress.String(), packet.GetSequence(), "timeout")
		im.keeper.Logger(ctx).Debug("failed to Sudo contract on packet timeout", err)
	} else {
		writeFn()
	}

	return nil
}

func (im IBCModule) outOfGasRecovery(
	ctx sdk.Context,
	gasMeter sdk.GasMeter,
	senderAddress sdk.AccAddress,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	failureType string,
) {
	if r := recover(); r != nil {
		_, ok := r.(sdk.ErrorOutOfGas)
		if !ok || !gasMeter.IsOutOfGas() {
			panic(r)
		}

		im.keeper.Logger(ctx).Debug("Out of gas", "Gas meter", gasMeter.String(), "Packet data", data)
		im.ContractManagerKeeper.AddContractFailure(ctx, packet.SourceChannel, senderAddress.String(), packet.GetSequence(), failureType)
	}
}

// createCachedContext creates a cached context for handling Sudo calls to CosmWasm smart-contracts.
// If there is an error during Sudo call, we can safely revert changes made in cached context.
// panics if there is no enough gas for sudoCall
func (im *IBCModule) createCachedContext(ctx sdk.Context) (sdk.Context, func(), sdk.GasMeter) {
	cacheCtx, writeFn := ctx.CacheContext()

	sudoLimit := im.ContractManagerKeeper.GetParams(ctx).SudoCallGasLimit
	if ctx.GasMeter().GasRemaining() < sudoLimit {
		panic(sdk.ErrorOutOfGas{Descriptor: fmt.Sprintf("%dgas - reserve for sudo call", sudoLimit)})
	}

	gasMeter := sdk.NewGasMeter(sudoLimit)

	cacheCtx = cacheCtx.WithGasMeter(gasMeter)

	return cacheCtx, writeFn, gasMeter
}
