package types

// feerefunder module event types
const (
	EventTypeDistributeAcknowledgementFee = "distribute_ack_fee"
	EventTypeDistributeTimeoutFee         = "distribute_timeout_fee"
	EventTypeLockFees                     = "distribute_lock_fees"

	AttributeKeyReceiver  = "receiver"
	AttributeKeyChannelId = "channel_id"
	AttributeKeyPortId    = "port_id"
	AttributeKeySequence  = "sequence"
	AttributeKeyPayer     = "payer"
)
