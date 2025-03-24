package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/x/revenue/types"
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
	rating := performanceRating(bpr, ovpr, 5, 5, 1000)
	require.Truef(t, math.LegacyOneDec().Equal(rating), rating.String())

	// performance is just a bit beyond the allowed to miss
	rating = performanceRating(bpr, ovpr, 5, 6, 1000)
	require.Truef(t, math.LegacyOneDec().GT(rating), rating.String())
	require.Truef(t, math.LegacyZeroDec().LT(rating), rating.String())

	// 10% is missing => 0.0 for both criteria
	rating = performanceRating(bpr, ovpr, 100, 100, 1000)
	require.Truef(t, math.LegacyZeroDec().Equal(rating), rating.String())

	// all votes, 5% missing blocks => somewhere between 0.5 and 1.0
	rating = performanceRating(bpr, ovpr, 50, 0, 1000)
	require.Truef(t, math.LegacyOneDec().GT(rating), rating.String())
	require.Truef(t, math.LegacyNewDecWithPrec(5, 1).LT(rating), rating.String())

	// all blocks, 5% missing votes => somewhere between 0.5 and 1.0
	rating = performanceRating(bpr, ovpr, 0, 50, 1000)
	require.Truef(t, math.LegacyOneDec().GT(rating), rating.String())
	require.Truef(t, math.LegacyNewDecWithPrec(5, 1).LT(rating), rating.String())

	// all votes, 10% missing blocks => 0.5 for votes, 0.0 for block
	rating = performanceRating(bpr, ovpr, 100, 0, 1000)
	require.Truef(t, math.LegacyNewDecWithPrec(5, 1).Equal(rating), rating.String())

	// all votes, 10%+1 missing blocks => 0.0 because required at least is 90%
	rating = performanceRating(bpr, ovpr, 101, 0, 1000)
	require.Truef(t, math.LegacyZeroDec().Equal(rating), rating.String())

	// when missed 70/95 of evaluation window, perf rating is about 46%
	rating = performanceRating(bpr, ovpr, 75, 75, 1000)
	require.Truef(t, math.LegacyNewDecWithPrec(457063711911357340, 18).Equal(rating), rating.String())

	// everything's been missed, perf rating is 0.0
	rating = performanceRating(bpr, ovpr, 1000, 1000, 1000)
	require.Truef(t, math.LegacyZeroDec().Equal(rating), rating.String())

	// define new performance requirements
	bpr.AllowedToMiss = math.LegacyNewDecWithPrec(5, 3)    // 0.5%
	bpr.RequiredAtLeast = math.LegacyNewDecWithPrec(9, 1)  // 90%
	ovpr.AllowedToMiss = math.LegacyNewDecWithPrec(5, 2)   // 5%
	ovpr.RequiredAtLeast = math.LegacyNewDecWithPrec(8, 1) // 80%

	// performance is within the new allowed to miss
	rating = performanceRating(bpr, ovpr, 5, 50, 1000)
	require.Truef(t, math.LegacyOneDec().Equal(rating), rating.String())

	// all but one block and votes missed (10%-1 blocks, 20%-1 votes) => perf rating is close to zero (1.7%)
	rating = performanceRating(bpr, ovpr, 99, 199, 1000)
	require.Truef(t, math.LegacyNewDecWithPrec(17115358571868268, 18).Equal(rating), rating.String())

	// 3% of missed blocks (2.5% greater than allowed to miss, at least 90% required) and
	// 10% of missed votes (5% greater than allowed to miss, at least 80% required)
	// result in a pretty decent performance rating of 91% due to quadratic function
	rating = performanceRating(bpr, ovpr, 30, 100, 1000)
	require.Truef(t, math.LegacyNewDecWithPrec(909818405663281010, 18).Equal(rating), rating.String())
}
