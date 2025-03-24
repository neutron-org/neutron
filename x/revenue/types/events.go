package types

const (
	// EventTypeRevenueDistribution is emitted on every revenue distribution.
	EventTypeRevenueDistribution = "revenue_distribution"
	// EventTypeRevenueDistribution is emitted on revenue distribution errors.
	EventTypeRevenueDistributionError = "revenue_distribution_error"
	// EventTypeRevenueDistribution is emitted when revenue processing resulted in none revenue.
	EventTypeRevenueDistributionNone = "revenue_distribution_none"
	// EventAttributeValidator contains the validator address.
	EventAttributeValidator = "validator"
	// EventAttributeRevenueAmount contains the revenue amount sent to the validator.
	EventAttributeRevenueAmount = "revenue_amount"
	// EventAttributePerformanceRating contains the validator's performance rating.
	EventAttributePerformanceRating = "performance_rating"
	// EventAttributeInActiveValsetForBlocksInPeriod contains the number of blocks the validator has
	// remained in the active validator set for in the current payment period.
	EventAttributeInActiveValsetForBlocksInPeriod = "in_active_valset_for_blocks_in_period"
	// EventAttributeCommittedBlocksInPeriod contains the number of blocks committed by the validator
	// in the payment period.
	EventAttributeCommittedBlocksInPeriod = "committed_blocks_in_period"
	// EventAttributeCommittedOracleVotesInPeriod contains the number of blocks committed by the
	// validator in the payment period.
	EventAttributeCommittedOracleVotesInPeriod = "committed_oracle_votes_in_period"
	// EventAttributeEffectivePeriodProgress contains the revenue amount multiplier value that
	// corresponds to the effective payment period progress.
	EventAttributeEffectivePeriodProgress = "effective_period_progress"
	// EventAttributeTotalBlockInPeriod contains the total number of blocks in the payment period.
	EventAttributeTotalBlockInPeriod = "total_block_in_period"
	// EventAttributePaymentFailure is added to the event when a payment fails and contains the
	// payment error message.
	EventAttributePaymentFailure = "payment_failure"
)
