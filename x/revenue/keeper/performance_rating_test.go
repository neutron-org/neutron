package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/require"
)

func TestPerformanceRating(t *testing.T) {
	params := types.Params{
		DenomCompensation: "untrn",
		BaseCompensation:  1500,
		BlocksPerformanceRequirement: &types.PerformanceRequirement{
			AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.5%
			RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 90%
		},
		OracleVotesPerformanceRequirement: &types.PerformanceRequirement{
			AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.5%
			RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 90%
		},
	}
	// performance is within the allowed to miss
	rating := PerformanceRating(params, 5, 5, 1000)
	require.Equal(t, math.LegacyOneDec(), rating)

	// performance is just a bit beyond the allowed to miss
	rating = PerformanceRating(params, 5, 6, 1000)
	require.True(t, math.LegacyOneDec().GT(rating))
	require.True(t, math.LegacyZeroDec().LT(rating))

	// 10% is missing => unacceptable
	rating = PerformanceRating(params, 100, 100, 1000)
	require.Equal(t, math.LegacyZeroDec(), rating)

	// all votes, 10% missing blocks => unacceptable
	rating = PerformanceRating(params, 100, 0, 1000)
	require.Equal(t, math.LegacyZeroDec(), rating)

	// when missed 70% of fined threshold (threshold - allowed), perf rating ~45%
	rating = PerformanceRating(params, 75, 75, 1000)
	require.Equal(t, rating, math.LegacyNewDecWithPrec(457063711911357341, 18))

	// define new performance requirements
	params = types.Params{
		DenomCompensation: "untrn",
		BaseCompensation:  1500,
		BlocksPerformanceRequirement: &types.PerformanceRequirement{
			AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.5%
			RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 90%
		},
		OracleVotesPerformanceRequirement: &types.PerformanceRequirement{
			AllowedToMiss:   math.LegacyNewDecWithPrec(5, 2), // 5%
			RequiredAtLeast: math.LegacyNewDecWithPrec(8, 1), // 80%
		},
	}

	// performance is within the new allowed to miss
	rating = PerformanceRating(params, 5, 50, 1000)
	require.Equal(t, math.LegacyOneDec(), rating)

	// 10%-1 blocks missed, 20%-1 votes missed => some rewards
	rating = PerformanceRating(params, 99, 199, 1000)
	require.True(t, math.LegacyOneDec().GT(rating))
	require.True(t, math.LegacyZeroDec().LT(rating))
}
