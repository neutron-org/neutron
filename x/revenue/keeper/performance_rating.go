package keeper

import (
	"cosmossdk.io/math"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
)

// PerformanceRating evaluates the performance of a validator based on its participation in block
// signing and oracle price voting. The function returns a normalized performance score, expressed
// as a decimal value between 0.0 and 1.0, where:
//
// 1.0 (100% performance): The validator meets or exceeds the performance requirements for both
// block signing and oracle price voting.
//
// 0.0 (0% performance): The validator fails to meet the minimum performance thresholds for either
// block signing or oracle price voting.
//
// A value between 0.0 and 1.0: The validator's performance partially meets the defined requirements,
// and the rating is calculated based on the extent to which the validator's performance deviates
// from the optimal values.
func PerformanceRating(
	blocksPR *revenuetypes.PerformanceRequirement,
	oracleVotesPR *revenuetypes.PerformanceRequirement,
	missedBlocks int64,
	missedOracleVotes int64,
	totalBlocks int64,
) math.LegacyDec {
	blocksPerfThreshold := math.LegacyOneDec().Sub(blocksPR.RequiredAtLeast)
	oracleVotesPerfThreshold := math.LegacyOneDec().Sub(oracleVotesPR.RequiredAtLeast)

	// if a validator has signed less blocks than required, the rating is zero
	missedBlocksShare := math.LegacyNewDec(missedBlocks).QuoInt64(totalBlocks)
	if missedBlocksShare.GTE(blocksPerfThreshold) {
		return math.LegacyZeroDec()
	}
	// if a validator has provided less oracle prices than required, the rating is zero
	missedOracleVotesShare := math.LegacyNewDec(missedOracleVotes).QuoInt64(totalBlocks)
	if missedOracleVotesShare.GTE(oracleVotesPerfThreshold) {
		return math.LegacyZeroDec()
	}

	// if a validator's performance is within the allowed bounds, they get the max rating
	if missedBlocksShare.LTE(blocksPR.AllowedToMiss) &&
		missedOracleVotesShare.LTE(oracleVotesPR.AllowedToMiss) {
		return math.LegacyOneDec()
	}

	// how much blocks/votes missed over the allowed value
	finedMissedBlocksShare := missedBlocksShare.Sub(blocksPR.AllowedToMiss)
	finedMissedOracleVotesShare := missedOracleVotesShare.Sub(oracleVotesPR.AllowedToMiss)

	// the missed blocks/votes span for (0.0;1.1) performance rating values
	blocksPerfEvalWindow := blocksPerfThreshold.Sub(blocksPR.AllowedToMiss)
	oracleVotesPerfEvalWindow := oracleVotesPerfThreshold.Sub(oracleVotesPR.AllowedToMiss)

	// calculated as how much blocks/votes missed in the eval window
	missedBlocksPerfQuo := finedMissedBlocksShare.Quo(blocksPerfEvalWindow)
	missedOracleVotesPerfQuo := finedMissedOracleVotesShare.Quo(oracleVotesPerfEvalWindow)

	// rating = 0.5 * ((1 - missedBlocksPerfQuo^2) + (1 - missedOracleVotesPerfQuo^2))
	rating := math.LegacyNewDecWithPrec(5, 1).Mul(
		math.LegacyOneDec().Sub(missedBlocksPerfQuo.Mul(missedBlocksPerfQuo)).
			Add(math.LegacyOneDec().Sub(missedOracleVotesPerfQuo.Mul(missedOracleVotesPerfQuo))),
	)
	return rating
}
