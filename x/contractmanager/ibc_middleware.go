package contractmanager

import (
	"fmt"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/neutron-org/neutron/neutronutils"
	neutronerrors "github.com/neutron-org/neutron/neutronutils/errors"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
)

type SudoLimitWrapper struct {
	contractmanagertypes.ContractManagerWrapper
}

// NewSudoLimitWrapper suppresses an error from a sudo contract handler and saves it to a store
func NewSudoLimitWrapper(keeper contractmanagertypes.ContractManagerWrapper) contractmanagertypes.ContractManagerWrapper {
	return SudoLimitWrapper{
		keeper,
	}
}

func (k SudoLimitWrapper) SudoResponse(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, ack channeltypes.Acknowledgement) ([]byte, error) {
	err := k.sudo(ctx, senderAddress, request, &ack)
	if err != nil {
		k.Logger(ctx).Debug("SudoLimitWrapper: failed to sudo contract", "error", err, "ackType", "Result")
	}
	return nil, nil
}

func (k SudoLimitWrapper) SudoError(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, ack channeltypes.Acknowledgement) ([]byte, error) {
	err := k.sudo(ctx, senderAddress, request, &ack)
	if err != nil {
		k.Logger(ctx).Debug("SudoLimitWrapper: failed to sudo contract", "error", err, "ackType", "ErrorSudoPayload")
	}
	return nil, nil
}

func (k SudoLimitWrapper) SudoTimeout(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) ([]byte, error) {
	err := k.sudo(ctx, senderAddress, request, nil)
	if err != nil {
		k.Logger(ctx).Debug("SudoLimitWrapper: failed to sudo contract", "error", err, "ackType", "Timeout")
	}
	return nil, nil
}

// sudo calls underlying sudo handlers with a limited amount of gas
// in case of `out of gas` panic it converts the panic into an error and stops `out of gas` panic propagation
func (k SudoLimitWrapper) sudo(ctx sdk.Context, sender sdk.AccAddress, packet channeltypes.Packet, ack *channeltypes.Acknowledgement) (err error) {
	ackType := contractmanagertypes.Ack
	cacheCtx, writeFn := neutronutils.CreateCachedContext(ctx, k.ContractManagerWrapper.GetParams(ctx).SudoCallGasLimit)
	func() {
		defer neutronerrors.OutOfGasRecovery(cacheCtx.GasMeter(), &err)
		// Actually we have only one kind of error returned from acknowledgement
		// maybe later we'll retrieve actual errors from events
		if ack == nil {
			ackType = contractmanagertypes.Timeout
			_, err = k.ContractManagerWrapper.SudoTimeout(cacheCtx, sender, packet)
		} else if ack.GetError() != "" {
			_, err = k.ContractManagerWrapper.SudoError(cacheCtx, sender, packet, *ack)
		} else {
			_, err = k.ContractManagerWrapper.SudoResponse(cacheCtx, sender, packet, *ack)
		}
	}()
	if err != nil {
		// the contract either returned an error or panicked with `out of gas`
		k.ContractManagerWrapper.AddContractFailure(ctx, &packet, sender.String(), ackType, ack)
	} else {
		writeFn()
	}

	ctx.GasMeter().ConsumeGas(cacheCtx.GasMeter().GasConsumedToLimit(), "consume gas from cached context")
	return
}

func (k SudoLimitWrapper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", contractmanagertypes.ModuleName))
}
