package transfer

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

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

	// We need to reserver this amount of gas (15000) on the context gas meter in order to add contract failure to keeper
	newGasMeter := sdk.NewGasMeter(ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed() - 15000)
	defer func() {
		if r := recover(); r != nil {
			_, ok := r.(sdk.ErrorOutOfGas)
			if !ok || !newGasMeter.IsOutOfGas() {
				panic(r)
			}

			im.keeper.Logger(ctx).Debug("Out of gas", "Gas meter", newGasMeter.String(), "Packet data", data)
			im.ContractmanagerKeeper.AddContractFailure(ctx, senderAddress.String(), packet.GetSequence(), "ack")

			err = nil
		}
	}()

	cacheCtx, writeFn := ctx.CacheContext()
	if strings.HasPrefix(ctx.GasMeter().String(), "BasicGasMeter") {
		cacheCtx = ctx.WithGasMeter(newGasMeter)
	}

	if ack.Success() {
		_, err = im.sudoHandler.SudoResponse(cacheCtx, senderAddress, packet, ack.GetResult())
	} else {
		// Actually we have only one kind of error returned from acknowledgement
		// maybe later we'll retrieve actual errors from events
		im.keeper.Logger(cacheCtx).Debug(ack.GetError(), "CheckTx", cacheCtx.IsCheckTx())
		_, err = im.sudoHandler.SudoError(cacheCtx, senderAddress, packet, ack.GetError())
	}

	if err != nil {
		im.ContractmanagerKeeper.AddContractFailure(ctx, senderAddress.String(), packet.GetSequence(), "ack")
		im.keeper.Logger(ctx).Debug("failed to Sudo contract on packet acknowledgement", err)
	} else {
		writeFn()
		ctx.GasMeter().ConsumeGas(newGasMeter.GasConsumed(), "consume from cached context")
	}

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

	_, err = im.sudoHandler.SudoTimeout(ctx, senderAddress, packet)
	if err != nil {
		im.keeper.Logger(ctx).Debug("failed to Sudo contract on packet timeout", err)
		return sdkerrors.Wrap(err, "failed to Sudo the contract on packet timeout")
	}

	return nil
}
