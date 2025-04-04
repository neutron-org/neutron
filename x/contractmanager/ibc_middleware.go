package contractmanager

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/store/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	contractmanagerkeeper "github.com/neutron-org/neutron/v6/x/contractmanager/keeper"
	contractmanagertypes "github.com/neutron-org/neutron/v6/x/contractmanager/types"
)

type SudoLimitWrapper struct {
	contractManager contractmanagertypes.ContractManagerKeeper
	contractmanagertypes.WasmKeeper
}

// NewSudoLimitWrapper suppresses an error from a Sudo contract handler and saves it to a store
func NewSudoLimitWrapper(contractManager contractmanagertypes.ContractManagerKeeper, sudoKeeper contractmanagertypes.WasmKeeper) contractmanagertypes.WasmKeeper {
	return SudoLimitWrapper{
		contractManager,
		sudoKeeper,
	}
}

// Sudo calls underlying Sudo handlers with a limited amount of gas
// in case of `out of gas` panic it converts the panic into an error and stops `out of gas` panic propagation
// if error happens during the Sudo call, we store the data that raised the error, and return the error
func (k SudoLimitWrapper) Sudo(ctx context.Context, contractAddress sdk.AccAddress, msg []byte) (resp []byte, err error) {
	c := sdk.UnwrapSDKContext(ctx)

	cacheCtx, writeFn := createCachedContext(c, k.contractManager.GetParams(ctx).SudoCallGasLimit)
	func() {
		defer outOfGasRecovery(cacheCtx.GasMeter(), &err)
		// Actually we have only one kind of error returned from acknowledgement
		// maybe later we'll retrieve actual errors from events
		resp, err = k.WasmKeeper.Sudo(cacheCtx, contractAddress, msg)
	}()
	if err != nil { // the contract either returned an error or panicked with `out of gas`
		failure := k.contractManager.AddContractFailure(
			ctx,
			contractAddress.String(),
			msg,
			contractmanagerkeeper.RedactError(err).Error(),
		)
		c.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				wasmtypes.EventTypeSudo,
				sdk.NewAttribute(wasmtypes.AttributeKeyContractAddr, contractAddress.String()),
				sdk.NewAttribute(contractmanagertypes.AttributeKeySudoFailureID, fmt.Sprintf("%d", failure.Id)),
				sdk.NewAttribute(contractmanagertypes.AttributeKeySudoError, err.Error()),
			),
		})
	} else {
		writeFn()
	}

	c.GasMeter().ConsumeGas(cacheCtx.GasMeter().GasConsumedToLimit(), "consume gas from cached context")
	return resp, err
}

func (k SudoLimitWrapper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", contractmanagertypes.ModuleName))
}

// outOfGasRecovery converts `out of gas` panic into an error
// leaving unprocessed any other kinds of panics
func outOfGasRecovery(
	gasMeter types.GasMeter,
	err *error,
) {
	if r := recover(); r != nil {
		_, ok := r.(types.ErrorOutOfGas)
		if !ok || !gasMeter.IsOutOfGas() {
			panic(r)
		}
		*err = contractmanagertypes.ErrSudoOutOfGas
	}
}

// createCachedContext creates a cached context with a limited gas meter.
func createCachedContext(ctx sdk.Context, gasLimit uint64) (sdk.Context, func()) {
	cacheCtx, writeFn := ctx.CacheContext()
	gasMeter := types.NewGasMeter(gasLimit)
	cacheCtx = cacheCtx.WithGasMeter(gasMeter)
	return cacheCtx, writeFn
}
