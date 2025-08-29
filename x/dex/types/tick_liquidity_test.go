package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	math_utils "github.com/neutron-org/neutron/v8/utils/math"
	"github.com/neutron-org/neutron/v8/x/dex/types"
)

func TestHasTokenEmptyReserves(t *testing.T) {
	// WHEN has no reserves
	tick := &types.PoolReserves{DecReservesMakerDenom: math_utils.ZeroPrecDec()}
	assert.False(t, tick.HasToken())
}

func TestHasTokenEmptyLO(t *testing.T) {
	// WHEN has no limits orders
	tick := &types.LimitOrderTranche{DecReservesMakerDenom: math_utils.NewPrecDec(0)}
	assert.False(t, tick.HasTokenIn())
}

func TestHasToken0HasReserves(t *testing.T) {
	// WHEN tick has Reserves
	tick := &types.PoolReserves{DecReservesMakerDenom: math_utils.NewPrecDec(10)}

	assert.True(t, tick.HasToken())
}

func TestHasTokenHasLO(t *testing.T) {
	// WHEN has limit ordeers
	tick := &types.LimitOrderTranche{DecReservesMakerDenom: math_utils.NewPrecDec(10)}
	assert.True(t, tick.HasTokenIn())
}
