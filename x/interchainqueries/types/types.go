package types

const (
	// AttributeKeyQueryID represents the key for event attribute delivering the query ID of a
	// registered interchain query.
	AttributeKeyQueryID = "query_id"
	// AttributeKeyOwner represents the key for event attribute delivering the address of the
	// registrator of an interchain query.
	AttributeKeyOwner = "owner"
	// AttributeKeyZoneID represents the key for event attribute delivering the zone ID where the
	// event has been produced.
	AttributeKeyZoneID = "zone_id"
	// AttributeKeyQueryType represents the key for event attribute delivering the query type
	// identifier (e.g. /cosmos.staking.v1beta1.Query/AllDelegations)
	AttributeKeyQueryType = "type"
	// AttributeKeyQueryParameters represents the key for event attribute delivering the parameters
	// of an interchain query.
	AttributeKeyQueryParameters = "parameters"

	// AttributeValueCategory represents the value for the 'module' event attribute.
	AttributeValueCategory = ModuleName
	// AttributeValueQuery represents the value for the 'action' event attribute.
	AttributeValueQuery = "query"
)
