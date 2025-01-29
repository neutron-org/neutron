package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v5/x/revenue/types"
)

func TestPerformanceRating(t *testing.T) {
	bpr := &types.PerformanceRequirement{ // blocks performance requirement
		AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.5%
		RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 90%
	}
	ovpr := &types.PerformanceRequirement{ // oracle votes performance requirement
		AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.5%
		RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 90%
	}

	// performance is within the allowed to miss
	rating := PerformanceRating(bpr, ovpr, 5, 5, 1000)
	require.Equal(t, math.LegacyOneDec(), rating)

	// performance is just a bit beyond the allowed to miss
	rating = PerformanceRating(bpr, ovpr, 5, 6, 1000)
	require.True(t, math.LegacyOneDec().GT(rating))
	require.True(t, math.LegacyZeroDec().LT(rating))

	// 10% is missing => unacceptable
	rating = PerformanceRating(bpr, ovpr, 100, 100, 1000)
	require.Equal(t, math.LegacyZeroDec(), rating)

	// all votes, 10% missing blocks => unacceptable
	rating = PerformanceRating(bpr, ovpr, 100, 0, 1000)
	require.Equal(t, math.LegacyZeroDec(), rating)

	// when missed 70/95 of evaluation window, perf rating is about 46%
	rating = PerformanceRating(bpr, ovpr, 75, 75, 1000)
	require.Equal(t, math.LegacyNewDecWithPrec(457063711911357340, 18), rating)

	// define new performance requirements
	bpr.AllowedToMiss = math.LegacyNewDecWithPrec(5, 3)    // 0.5%
	bpr.RequiredAtLeast = math.LegacyNewDecWithPrec(9, 1)  // 90%
	ovpr.AllowedToMiss = math.LegacyNewDecWithPrec(5, 2)   // 5%
	ovpr.RequiredAtLeast = math.LegacyNewDecWithPrec(8, 1) // 80%

	// performance is within the new allowed to miss
	rating = PerformanceRating(bpr, ovpr, 5, 50, 1000)
	require.Equal(t, math.LegacyOneDec(), rating)

	// all but one block and votes missed (10%-1 blocks, 20%-1 votes) => perf rating is close to zero (1.7%)
	rating = PerformanceRating(bpr, ovpr, 99, 199, 1000)
	require.Equal(t, math.LegacyNewDecWithPrec(17115358571868268, 18), rating)

	// 3% of missed blocks (2.5% greater than allowed to miss, at least 90% required) and
	// 10% of missed votes (5% greater than allowed to miss, at least 80% required)
	// result in a pretty decent performance rating of 91% due to quadratic function
	rating = PerformanceRating(bpr, ovpr, 30, 100, 1000)
	require.Equal(t, math.LegacyNewDecWithPrec(909818405663281010, 18), rating)
}
