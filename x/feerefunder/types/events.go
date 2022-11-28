package types

// feerefunder module event types
const (
	EventTypeDistributeAcknowledgementFee = "distribute_ack_fee"
	EventTypeDistributeTimeoutFee         = "distribute_timeout_fee"
	EventTypeLockFees                     = "lock_fees"

	AttributeKeyReceiver  = "receiver"
	AttributeKeyChannelID = "channel_id"
	AttributeKeyPortID    = "port_id"
	AttributeKeySequence  = "sequence"
	AttributeKeyPayer     = "payer"
)
