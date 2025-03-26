package keeper

import (
	"cosmossdk.io/math"

	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

// evaluateValCommitment calculates validator's performance rating and expected compensation amount.
func evaluateValCommitment(
	params revenuetypes.Params,
	baseRevenueAmount math.Int,
	valInfo revenuetypes.ValidatorInfo,
	blocksInPeriod uint64,
) (perfRating math.LegacyDec, valCompensation math.Int) {
	// is calculated against the total number of blocks the val has been in the active valset for
	// in the current period.
	perfRating = performanceRating(
		params.BlocksPerformanceRequirement,
		params.OracleVotesPerformanceRequirement,
		valInfo.InActiveValsetForBlocksInPeriod-valInfo.CommitedBlocksInPeriod,
		valInfo.InActiveValsetForBlocksInPeriod-valInfo.CommitedOracleVotesInPeriod,
		valInfo.InActiveValsetForBlocksInPeriod,
	)

	if blocksInPeriod == 0 {
		valCompensation = math.ZeroInt()
		return
	}

	valCompensation = perfRating.
		MulInt(baseRevenueAmount).
		MulInt(math.NewIntFromUint64(valInfo.InActiveValsetForBlocksInPeriod)).
		QuoInt(math.NewIntFromUint64(blocksInPeriod)).
		TruncateInt()

	return
}

// performanceRating evaluates the performance of a validator based on its participation in block
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
func performanceRating(
	blocksPR *revenuetypes.PerformanceRequirement,
	oracleVotesPR *revenuetypes.PerformanceRequirement,
	missedBlocks uint64,
	missedOracleVotes uint64,
	totalBlocks uint64,
) math.LegacyDec {
	if totalBlocks == 0 {
		return math.LegacyZeroDec()
	}

	missedBlocksInt := math.NewIntFromUint64(missedBlocks)
	missedOracleVotesInt := math.NewIntFromUint64(missedOracleVotes)
	totalBlocksInt := math.NewIntFromUint64(totalBlocks)

	blocksPerfThreshold := math.LegacyOneDec().Sub(blocksPR.RequiredAtLeast)
	oracleVotesPerfThreshold := math.LegacyOneDec().Sub(oracleVotesPR.RequiredAtLeast)

	// if a validator has signed less blocks than required, the rating is zero
	missedBlocksShare := math.LegacyNewDecFromInt(missedBlocksInt).QuoInt(totalBlocksInt)
	if missedBlocksShare.GT(blocksPerfThreshold) {
		return math.LegacyZeroDec()
	}
	// if a validator has provided less oracle prices than required, the rating is zero
	missedOracleVotesShare := math.LegacyNewDecFromInt(missedOracleVotesInt).QuoInt(totalBlocksInt)
	if missedOracleVotesShare.GT(oracleVotesPerfThreshold) {
		return math.LegacyZeroDec()
	}

	// if a validator's performance is within the allowed bounds, they get the max rating
	if missedBlocksShare.LTE(blocksPR.AllowedToMiss) &&
		missedOracleVotesShare.LTE(oracleVotesPR.AllowedToMiss) {
		return math.LegacyOneDec()
	}

	missedBlocksPerfQuo := calCMissedPerfQuo(missedBlocksShare, blocksPR.AllowedToMiss, blocksPerfThreshold)
	missedOracleVotesPerfQuo := calCMissedPerfQuo(missedOracleVotesShare, oracleVotesPR.AllowedToMiss, oracleVotesPerfThreshold)

	// rating = 0.5 * ((1 - missedBlocksPerfQuo^2) + (1 - missedOracleVotesPerfQuo^2))
	rating := math.LegacyNewDecWithPrec(5, 1).Mul(
		math.LegacyOneDec().Sub(missedBlocksPerfQuo.Mul(missedBlocksPerfQuo)).
			Add(math.LegacyOneDec().Sub(missedOracleVotesPerfQuo.Mul(missedOracleVotesPerfQuo))),
	)
	return rating
}

// calCMissedPerfQuo calculates the negative coefficient based on the missed share, allowed to miss,
// and performance threshold. If the missed share is LTE allowed to miss, the returned value is
// 0.0, i.e. no negative coefficient for this criteria.
func calCMissedPerfQuo(
	missedShare math.LegacyDec,
	allowedToMiss math.LegacyDec,
	perfThreshold math.LegacyDec,
) math.LegacyDec {
	if missedShare.LTE(allowedToMiss) {
		return math.LegacyZeroDec()
	}

	finedMissedShare := missedShare.Sub(allowedToMiss)    // how much missed over the allowed value
	perfEvalWindow := perfThreshold.Sub(allowedToMiss)    // span of evaluation window
	missedPerfQuo := finedMissedShare.Quo(perfEvalWindow) // how much missed in the evaluation window

	return missedPerfQuo
}
