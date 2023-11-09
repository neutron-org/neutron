package contractmanager

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
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
func (k SudoLimitWrapper) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (resp []byte, err error) {
	cacheCtx, writeFn := createCachedContext(ctx, k.contractManager.GetParams(ctx).SudoCallGasLimit)
	func() {
		defer outOfGasRecovery(cacheCtx.GasMeter(), &err)
		// Actually we have only one kind of error returned from acknowledgement
		// maybe later we'll retrieve actual errors from events
		resp, err = k.WasmKeeper.Sudo(cacheCtx, contractAddress, msg)
	}()
	if err != nil { // the contract either returned an error or panicked with `out of gas`
		failure := k.contractManager.AddContractFailure(ctx, contractAddress.String(), msg, redactError(err).Error())
		ctx.EventManager().EmitEvents(sdk.Events{
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

	ctx.GasMeter().ConsumeGas(cacheCtx.GasMeter().GasConsumedToLimit(), "consume gas from cached context")
	return
}

func (k SudoLimitWrapper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", contractmanagertypes.ModuleName))
}

// outOfGasRecovery converts `out of gas` panic into an error
// leaving unprocessed any other kinds of panics
func outOfGasRecovery(
	gasMeter sdk.GasMeter,
	err *error,
) {
	if r := recover(); r != nil {
		_, ok := r.(sdk.ErrorOutOfGas)
		if !ok || !gasMeter.IsOutOfGas() {
			panic(r)
		}
		*err = contractmanagertypes.ErrSudoOutOfGas
	}
}

// createCachedContext creates a cached context with a limited gas meter.
func createCachedContext(ctx sdk.Context, gasLimit uint64) (sdk.Context, func()) {
	cacheCtx, writeFn := ctx.CacheContext()
	gasMeter := sdk.NewGasMeter(gasLimit)
	cacheCtx = cacheCtx.WithGasMeter(gasMeter)
	return cacheCtx, writeFn
}

// redactError removes non-determenistic details from the error returning just codespace and core
// of the error. Returns full error for system errors.
//
// Copy+paste from https://github.com/neutron-org/wasmd/blob/5b59886e41ed55a7a4a9ae196e34b0852285503d/x/wasm/keeper/msg_dispatcher.go#L175-L190
func redactError(err error) error {
	// Do not redact system errors
	// SystemErrors must be created in x/wasm and we can ensure determinism
	if wasmvmtypes.ToSystemError(err) != nil {
		return err
	}

	// FIXME: do we want to hardcode some constant string mappings here as well?
	// Or better document them? (SDK error string may change on a patch release to fix wording)
	// sdk/11 is out of gas
	// sdk/5 is insufficient funds (on bank send)
	// (we can theoretically redact less in the future, but this is a first step to safety)
	codespace, code, _ := errorsmod.ABCIInfo(err, false)
	return fmt.Errorf("codespace: %s, code: %d", codespace, code)
}
