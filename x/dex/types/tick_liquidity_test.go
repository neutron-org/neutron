package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func TestHasTokenEmptyReserves(t *testing.T) {
	// WHEN has no reserves
	tick := &types.PoolReserves{ReservesMakerDenom: math.ZeroInt()}
	assert.False(t, tick.HasToken())
}

func TestHasTokenEmptyLO(t *testing.T) {
	// WHEN has no limits orders
	tick := &types.LimitOrderTranche{ReservesMakerDenom: math.NewInt(0)}
	assert.False(t, tick.HasTokenIn())
}

func TestHasToken0HasReserves(t *testing.T) {
	// WHEN tick has Reserves
	tick := &types.PoolReserves{ReservesMakerDenom: math.NewInt(10)}

	assert.True(t, tick.HasToken())
}

func TestHasTokenHasLO(t *testing.T) {
	// WHEN has limit ordeers
	tick := &types.LimitOrderTranche{ReservesMakerDenom: math.NewInt(10)}
	assert.True(t, tick.HasTokenIn())
}
