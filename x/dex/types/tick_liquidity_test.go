package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/assert"
)

func TestHasTokenEmptyReserves(t *testing.T) {
	// WHEN has no reserves
	tick := &types.PoolReserves{ReservesMakerDenom: sdk.ZeroInt()}
	assert.False(t, tick.HasToken())
}

func TestHasTokenEmptyLO(t *testing.T) {
	// WHEN has no limits orders
	tick := &types.LimitOrderTranche{ReservesMakerDenom: sdk.NewInt(0)}
	assert.False(t, tick.HasTokenIn())
}

func TestHasToken0HasReserves(t *testing.T) {
	// WHEN tick has Reserves
	tick := &types.PoolReserves{ReservesMakerDenom: sdk.NewInt(10)}

	assert.True(t, tick.HasToken())
}

func TestHasTokenHasLO(t *testing.T) {
	// WHEN has limit ordeers
	tick := &types.LimitOrderTranche{ReservesMakerDenom: sdk.NewInt(10)}
	assert.True(t, tick.HasTokenIn())
}
