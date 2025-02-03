package types

const (
	// EventTypeRevenueDistribution is emitted on every revenue distribution.
	EventTypeRevenueDistribution = "revenue_distribution"
	// EventAttributeValidator contains the validator address.
	EventAttributeValidator = "validator"
	// EventAttributeRevenueAmount contains the revenue amount sent to the validator.
	EventAttributeRevenueAmount = "revenue_amount"
	// EventAttributePerformanceRating contains the validator's performance rating.
	EventAttributePerformanceRating = "performance_rating"
	// EventAttributeCommittedBlocksInPeriod contains the number of blocks committed by the validator
	// in the payment period.
	EventAttributeCommittedBlocksInPeriod = "committed_blocks_in_period"
	// EventAttributeCommittedOracleVotesInPeriod contains the number of blocks committed by the
	// validator in the payment period.
	EventAttributeCommittedOracleVotesInPeriod = "committed_oracle_votes_in_period"
	// EventAttributeTotalBlockInPeriod contains the total number of blocks in the payment period.
	EventAttributeTotalBlockInPeriod = "total_block_in_period"
	// EventAttributePaymentFailure is added to the event when a payment fails and contains the
	// payment error message.
	EventAttributePaymentFailure = "payment_failure"
)
