package utils_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v8/utils"
)

func TestSanitizeCoins(t *testing.T) {
	input := sdk.Coins{
		sdk.NewInt64Coin("atom", 10),
		sdk.NewInt64Coin("btc", 50),
		sdk.NewInt64Coin("atom", 5),
		sdk.NewInt64Coin("btc", 0),
	}
	expected := sdk.Coins{
		sdk.NewInt64Coin("atom", 15),
		sdk.NewInt64Coin("btc", 50),
	}

	result := utils.SanitizeCoins(input)
	require.Equal(t, expected, result)
}
