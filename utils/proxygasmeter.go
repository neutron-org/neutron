// https://github.com/MANTRA-Chain/mantrachain/pull/438

package utils

import (
	"fmt"
	"math"

	storetypes "cosmossdk.io/store/types"
)

var _ storetypes.GasMeter = &ProxyGasMeter{}

// ProxyGasMeter wraps another GasMeter, but enforces a lower gas limit.
// Gas consumption is delegated to the wrapped GasMeter, so it won't risk losing gas accounting compared to standalone
// gas meter.
type ProxyGasMeter struct {
	storetypes.GasMeter
	limit storetypes.Gas
}

// NewProxyGasMeter returns a new ProxyGasMeter which wraps the provided gas meter.
// The limit is the maximum gas that can be consumed on top of consumed gas of the wrapped gas meter.
//
// If limit is greater than or equal to the remaining gas, no wrapping is needed and the original gas meter is returned.
func NewProxyGasMeter(gasMeter storetypes.GasMeter, limit storetypes.Gas) storetypes.GasMeter {
	if limit >= gasMeter.GasRemaining() {
		return gasMeter
	}

	return &ProxyGasMeter{
		GasMeter: gasMeter,
		limit:    limit + gasMeter.GasConsumed(),
	}
}

func (pgm ProxyGasMeter) GasRemaining() storetypes.Gas {
	if pgm.IsPastLimit() {
		return 0
	}
	return pgm.limit - pgm.GasConsumed()
}

func (pgm ProxyGasMeter) Limit() storetypes.Gas {
	return pgm.limit
}

func (pgm ProxyGasMeter) IsPastLimit() bool {
	return pgm.GasConsumed() > pgm.limit
}

func (pgm ProxyGasMeter) IsOutOfGas() bool {
	return pgm.GasConsumed() >= pgm.limit
}

// addUint64Overflow performs the addition operation on two uint64 integers and
// returns a boolean on whether or not the result overflows.
func addUint64Overflow(a, b uint64) (uint64, bool) {
	if math.MaxUint64-a < b {
		return 0, true
	}

	return a + b, false
}

func (pgm ProxyGasMeter) ConsumeGas(amount storetypes.Gas, descriptor string) {
	consumed, overflow := addUint64Overflow(pgm.GasMeter.GasConsumed(), amount)
	if overflow {
		panic(storetypes.ErrorGasOverflow{Descriptor: descriptor})
	}

	if consumed > pgm.limit {
		panic(storetypes.ErrorOutOfGas{Descriptor: descriptor})
	}

	pgm.GasMeter.ConsumeGas(amount, descriptor)
}

func (pgm ProxyGasMeter) String() string {
	return fmt.Sprintf("ProxyGasMeter{consumed: %d, limit: %d}", pgm.GasConsumed(), pgm.limit)
}
