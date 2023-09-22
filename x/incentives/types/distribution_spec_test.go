package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/neutron-org/neutron/x/incentives/types"
)

func TestDistributionSpec_Add(t *testing.T) {
	spec1 := types.DistributionSpec{
		"alice": sdk.Coins{sdk.NewInt64Coin("coin1", 100), sdk.NewInt64Coin("coin2", 200)},
		"bob":   sdk.Coins{sdk.NewInt64Coin("coin1", 300), sdk.NewInt64Coin("coin2", 400)},
	}

	spec2 := types.DistributionSpec{
		"alice": sdk.Coins{sdk.NewInt64Coin("coin1", 100), sdk.NewInt64Coin("coin2", 200)},
		"carol": sdk.Coins{sdk.NewInt64Coin("coin1", 500), sdk.NewInt64Coin("coin2", 600)},
	}

	expected := types.DistributionSpec{
		"alice": sdk.Coins{sdk.NewInt64Coin("coin1", 200), sdk.NewInt64Coin("coin2", 400)},
		"bob":   sdk.Coins{sdk.NewInt64Coin("coin1", 300), sdk.NewInt64Coin("coin2", 400)},
		"carol": sdk.Coins{sdk.NewInt64Coin("coin1", 500), sdk.NewInt64Coin("coin2", 600)},
	}

	result := spec1.Add(spec2)
	assert.Equal(t, expected, result)
}

func TestDistributionSpec_GetTotal(t *testing.T) {
	spec := types.DistributionSpec{
		"alice": sdk.Coins{sdk.NewInt64Coin("coin1", 100), sdk.NewInt64Coin("coin2", 200)},
		"bob":   sdk.Coins{sdk.NewInt64Coin("coin1", 300), sdk.NewInt64Coin("coin2", 400)},
		"carol": sdk.Coins{sdk.NewInt64Coin("coin1", 500), sdk.NewInt64Coin("coin2", 600)},
	}

	expected := sdk.Coins{sdk.NewInt64Coin("coin1", 900), sdk.NewInt64Coin("coin2", 1200)}

	total := spec.GetTotal()
	assert.Equal(t, expected, total)
}
