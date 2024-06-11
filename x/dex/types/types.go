package types

const (
	// EventTypeNeutronMessage defines the event type used by the Interchain Queries module events.
	EventTypeNeutronMessage = "neutron"

	AttributeDenom                = "denom"
	AttributeWithdrawn            = "total_withdrawn"
	AttributeGasConsumed          = "gas_consumed"
	AttributeLiquidityTickType    = "liquidity_tick_type"
	AttributeLp                   = "lp"
	AttributeLimitOrder           = "limit_order"
	AttributeIsExpiringLimitOrder = "is_expiring_limit_order"
	AttributeInc                  = "inc"
	AttributeDec                  = "dec"
	AttributePairID               = "pair_id"
)
