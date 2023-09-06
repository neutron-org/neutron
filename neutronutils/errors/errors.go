package errors

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OutOfGasRecovery converts `out of gas` panic into an error
// leaving unprocessed any other kinds of panics
func OutOfGasRecovery(
	gasMeter sdk.GasMeter,
	err *error,
) {
	if r := recover(); r != nil {
		_, ok := r.(sdk.ErrorOutOfGas)
		if !ok || !gasMeter.IsOutOfGas() {
			panic(r)
		}
		*err = errors.Wrapf(errors.ErrPanic, "%v", r)
	}
}
