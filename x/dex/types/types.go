package types

const (
	// EventTypeNeutronMessage defines the event type used by the Interchain Queries module events.
	EventTypeNeutronMessage = "neutron"

	AttributeDenom                   = "denom"
	AttributeWithdrawn               = "total_withdrawn"
	AttributeGasConsumed             = "gas_consumed"
	AttributeIncExpiredOrders        = "total_orders_expired_inc"
	AttributeDecExpiredOrders        = "total_orders_expired_dec"
	AttributeTotalLimitOrders        = "total_orders_limit"
	AttributeTotalTickLiquiditiesInc = "total_tick_liqidities_inc"
	AttributeTotalTickLiquiditiesDec = "total_tick_liqidities_dec"
)
