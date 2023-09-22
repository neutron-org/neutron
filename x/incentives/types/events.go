package types

// Incentive module event types.
const (
	TypeEvtCreateGauge  = "create_gauge"
	TypeEvtAddToGauge   = "add_to_gauge"
	TypeEvtDistribution = "distribution"

	AttributeGaugeID  = "gauge_id"
	AttributeReceiver = "receiver"
	AttributeAmount   = "amount"

	TypeEvtStake   = "stake"
	TypeEvtUnstake = "unstake"

	AttributeStakeID        = "stake_id"
	AttributeStakeOwner     = "owner"
	AttributeStakeAmount    = "amount"
	AttributeStakeStakeTime = "stake_time"
	AttributeUnstakedCoins  = "unstaked_coins"
)
