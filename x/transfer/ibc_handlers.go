package transfer

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/neutron-org/neutron/neutronutils"

	neutronerrors "github.com/neutron-org/neutron/neutronutils/errors"
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

	cacheCtx, writeFn, newGasMeter := neutronutils.CreateCachedContext(ctx, im.ContractManagerKeeper.GetParams(ctx).SudoCallGasLimit)

	// distribute fee
	im.wrappedKeeper.FeeKeeper.DistributeAcknowledgementFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))

	func() {
		defer neutronerrors.OutOfGasRecovery(newGasMeter, &err)
		if ack.Success() {
			_, err = im.ContractManagerKeeper.SudoResponse(cacheCtx, senderAddress, packet, ack.GetResult())
		} else {
			// Actually we have only one kind of error returned from acknowledgement
			// maybe later we'll retrieve actual errors from events
			im.keeper.Logger(cacheCtx).Debug(ack.GetError(), "CheckTx", cacheCtx.IsCheckTx())
			_, err = im.ContractManagerKeeper.SudoError(cacheCtx, senderAddress, packet, ack.GetError())
		}
	}()

	if err != nil {
		// the contract either returned an error or panicked with `out of gas`
		im.ContractManagerKeeper.AddContractFailure(ctx, packet.SourceChannel, senderAddress.String(), packet.GetSequence(), "ack")
		im.keeper.Logger(ctx).Debug("failed to Sudo contract on packet acknowledgement", err)
	} else {
		writeFn()
	}

	ctx.GasMeter().ConsumeGas(newGasMeter.GasConsumedToLimit(), "consume gas from cached context")

	im.keeper.Logger(ctx).Debug("acknowledgement received", "Packet data", data, "CheckTx", ctx.IsCheckTx())

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a Sudo call.
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

	cacheCtx, writeFn, newGasMeter := neutronutils.CreateCachedContext(ctx, im.ContractManagerKeeper.GetParams(ctx).SudoCallGasLimit)

	// distribute fee
	im.wrappedKeeper.FeeKeeper.DistributeTimeoutFee(ctx, relayer, feetypes.NewPacketID(packet.SourcePort, packet.SourceChannel, packet.Sequence))
	func() {
		defer neutronerrors.OutOfGasRecovery(newGasMeter, &err)
		_, err = im.ContractManagerKeeper.SudoTimeout(cacheCtx, senderAddress, packet)
	}()

	if err != nil {
		// the contract either returned an error or panicked with `out of gas`
		im.ContractManagerKeeper.AddContractFailure(ctx, packet.SourceChannel, senderAddress.String(), packet.GetSequence(), "timeout")
		im.keeper.Logger(ctx).Debug("failed to Sudo contract on packet timeout", err)
	} else {
		writeFn()
	}

	ctx.GasMeter().ConsumeGas(newGasMeter.GasConsumedToLimit(), "consume gas from cached context")

	return nil
}
