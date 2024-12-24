package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/require"
)

func TestPerformanceRating(t *testing.T) {
	params := types.Params{
		DenomCompensation:    "untrn",
		OracleVoteWeight:     math.LegacyDec{},
		PerformanceThreshold: math.LegacyNewDecWithPrec(1, 1), // 0.1
		AllowedMissed:        math.LegacyNewDecWithPrec(5, 3), // 0.005
	}
	rating := PerformanceRating(params, 5, 5, 1000)
	require.Equal(t, math.LegacyOneDec(), rating)

	rating = PerformanceRating(params, 5, 6, 1000)
	require.True(t, math.LegacyOneDec().GT(rating))

	rating = PerformanceRating(params, 100, 100, 1000)
	require.Equal(t, math.LegacyZeroDec(), rating)

	rating = PerformanceRating(params, 100, 99, 1000)
	require.True(t, math.LegacyZeroDec().LT(rating))

	// when missed 70% of fined threshold (threshold - allowed), perf rating ~45%
	rating = PerformanceRating(params, 75, 75, 1000)
	require.Equal(t, rating, math.LegacyNewDecWithPrec(457063711911357341, 18))
}
