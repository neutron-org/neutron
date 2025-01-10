package keeper

import (
	"log"

	"cosmossdk.io/math"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
)

// PerformanceRating calculates rating like a superlinearly function, e.g
// `f(a+b) > f(a) + f(b)`
// TODO: describe returned value: [0.0; 1.0]?
func PerformanceRating(
	params revenuetypes.Params,
	missedBlocks int64,
	missedOracleVotes int64,
	totalBlocks int64,
) math.LegacyDec {
	blocksPerfThreshold := math.LegacyOneDec().Sub(params.BlocksPerformanceRequirement.RequiredAtLeast)
	oracleVotesPerfThreshold := math.LegacyOneDec().Sub(params.OracleVotesPerformanceRequirement.RequiredAtLeast)

	log.Printf("params: %+v", params)
	log.Printf("missed: blocks %d, votes %d; total: %d", missedBlocks, missedOracleVotes, totalBlocks)
	log.Printf("thresholds: blocks %s, votes %s", blocksPerfThreshold, oracleVotesPerfThreshold)

	// if a validator has signed less blocks than required, the rating is zero
	missedBlocksShare := math.LegacyNewDec(missedBlocks).QuoInt64(totalBlocks)
	log.Printf("missed blocks share: %s", missedBlocksShare)
	if missedBlocksShare.GTE(blocksPerfThreshold) {
		return math.LegacyZeroDec()
	}
	// if a validator has provided less oracle prices than required, the rating is zero
	missedOracleVotesShare := math.LegacyNewDec(missedOracleVotes).QuoInt64(totalBlocks)
	log.Printf("missed votes share: %s", missedOracleVotesShare)
	if missedOracleVotesShare.GTE(oracleVotesPerfThreshold) {
		return math.LegacyZeroDec()
	}

	// if a validator has missed less blocks and prices than allowed, they get the max rating
	if missedBlocksShare.LT(params.BlocksPerformanceRequirement.AllowedToMiss) &&
		missedOracleVotesShare.LT(params.OracleVotesPerformanceRequirement.AllowedToMiss) {
		return math.LegacyOneDec()
	}

	finedMissedBlocksShare := missedBlocksShare.Sub(params.BlocksPerformanceRequirement.AllowedToMiss)
	finedMissedOracleVotesShare := missedOracleVotesShare.Sub(params.OracleVotesPerformanceRequirement.AllowedToMiss)
	finedMissedShareAvg := finedMissedBlocksShare.Add(finedMissedOracleVotesShare).Quo(math.LegacyNewDec(2))
	log.Printf("fined shares: blocks %s, votes %s, avg %s", finedMissedBlocksShare, finedMissedOracleVotesShare, finedMissedShareAvg)

	// use parabolic function as superlinear at range [0;Threshold-Allowed],
	// TODO:
	a := math.LegacyOneDec().Quo(
		blocksPerfThreshold.Sub(params.BlocksPerformanceRequirement.AllowedToMiss).Mul(
			oracleVotesPerfThreshold.Sub(params.OracleVotesPerformanceRequirement.AllowedToMiss),
		),
	)
	log.Printf("a: %s", a)
	log.Printf("a = 1 / (%s * %s)", blocksPerfThreshold.Sub(params.BlocksPerformanceRequirement.AllowedToMiss), oracleVotesPerfThreshold.Sub(params.OracleVotesPerformanceRequirement.AllowedToMiss))
	rating := math.LegacyOneDec().Sub(finedMissedShareAvg.Mul(finedMissedShareAvg).Mul(a))
	log.Printf("result rating: %s", rating)
	return rating
}
