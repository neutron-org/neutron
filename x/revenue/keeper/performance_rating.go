package keeper

import (
	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/v5/x/revenue/types"
)

// PerformanceRating calculates rating like a superlinearly function, e.g
// `f(a+b) > f(a) + f(b)`
// TODO: describe returned value: [0.0; 1.0]?
func PerformanceRating(
	params types.Params,
	missedBlocks uint64,
	missedOracleVotes uint64,
	totalBlocks uint64,
) math.LegacyDec {
	// missedShare = (missed_blocks + missed_votes) / (2 * total_blocks)
	// why `2 * total_blocks`? because max `missed_blocks + missed_votes` = 2 * total_blocks
	// missedShare is aggregate of missed blocks and missed votes
	missedShare := math.LegacyNewDec(int64(missedBlocks)).
		Add(math.LegacyNewDec(int64(missedOracleVotes))).
		QuoInt64(int64(totalBlocks * 2))
	if missedShare.LTE(params.AllowedMissed) {
		return math.LegacyOneDec()
	}
	if missedShare.GTE(params.PerformanceThreshold) {
		return math.LegacyZeroDec()
	}
	finedMissedShare := missedShare.Sub(params.AllowedMissed)
	// use parabolic function as superlinear at range [0;Threshold-Allowed],
	// TODO:
	a := math.LegacyOneDec().Quo(
		params.PerformanceThreshold.Sub(params.AllowedMissed).Mul(
			params.PerformanceThreshold.Sub(params.AllowedMissed),
		),
	)
	rating := math.LegacyOneDec().Sub(finedMissedShare.Mul(finedMissedShare).Mul(a))
	return rating
}
