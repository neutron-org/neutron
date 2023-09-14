package contractmanager

import (
	"fmt"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/neutronutils"
	neutronerrors "github.com/neutron-org/neutron/neutronutils/errors"
	"github.com/neutron-org/neutron/x/contractmanager/keeper"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
)

type SudoLimitWrapper struct {
	contractManager keeper.Keeper
	contractmanagertypes.WasmKeeper
}

// NewSudoLimitWrapper suppresses an error from a Sudo contract handler and saves it to a store
func NewSudoLimitWrapper(contractManager keeper.Keeper, sudoKeeper contractmanagertypes.WasmKeeper) contractmanagertypes.WasmKeeper {
	return SudoLimitWrapper{
		contractManager,
		sudoKeeper,
	}
}

// Sudo calls underlying Sudo handlers with a limited amount of gas
// in case of `out of gas` panic it converts the panic into an error and stops `out of gas` panic propagation
// if error happens during the Sudo call, we store the data that raised the error, and return the error
func (k SudoLimitWrapper) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (resp []byte, err error) {
	cacheCtx, writeFn := neutronutils.CreateCachedContext(ctx, k.contractManager.GetParams(ctx).SudoCallGasLimit)
	func() {
		defer neutronerrors.OutOfGasRecovery(cacheCtx.GasMeter(), &err)
		// Actually we have only one kind of error returned from acknowledgement
		// maybe later we'll retrieve actual errors from events
		resp, err = k.WasmKeeper.Sudo(cacheCtx, contractAddress, msg)

	}()
	if err != nil {
		// the contract either returned an error or panicked with `out of gas`
		k.contractManager.AddContractFailure(ctx, contractAddress.String(), msg)
	} else {
		writeFn()
	}

	ctx.GasMeter().ConsumeGas(cacheCtx.GasMeter().GasConsumedToLimit(), "consume gas from cached context")
	return
}

func (k SudoLimitWrapper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", contractmanagertypes.ModuleName))
}
