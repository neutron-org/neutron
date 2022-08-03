package bindings

// NeutronQuery contains neutron custom queries.
type NeutronQuery struct {
	/// Registered Interchain Query Result for specified QueryID
	InterchainQueryResult *QueryRegisteredQueryResultRequest `json:"interchain_query_result,omitempty"`
	/// Interchain account address for specified ConnectionID and OwnerAddress
	InterchainAccountAddress *QueryInterchainAccountAddressRequest `json:"interchain_account_address,omitempty"`
	/// RegisteredInterchainQueries
	RegisteredInterchainQueries *QueryRegisteredQueriesRequest `json:"registered_interchain_queries,omitempty"`
	/// RegisteredInterchainQuery
	RegisteredInterchainQuery *QueryRegisteredQueryRequest `json:"registered_interchain_query,omitempty"`
}

/* Requests */

type QueryRegisteredQueryResultRequest struct {
	QueryId uint64 `json:"query_id,omitempty"`
}

type QueryInterchainAccountAddressRequest struct {
	// owner_address is the owner of the interchain account on the controller chain
	OwnerAddress string `json:"owner_address,omitempty"`
	// interchain_account_id is an identifier of your interchain account from which you want to execute msgs
	InterchainAccountId string `json:"interchain_account_id,omitempty"`
	// connection_id is an IBC connection identifier between Neutron and remote chain
	ConnectionId string `json:"connection_id,omitempty"`
}

type QueryRegisteredQueriesRequest struct {
}

type QueryRegisteredQueryRequest struct {
	QueryId uint64 `json:"query_id,omitempty"`
}

/* Responses */

type QueryRegisteredQueryResponse struct {
	RegisteredQuery *RegisteredQuery `json:"registered_query,omitempty"`
}

type QueryRegisteredQueriesResponse struct {
	RegisteredQueries []RegisteredQuery `json:"registered_queries"`
}

type RegisteredQuery struct {
	// The unique id of the registered query.
	Id uint64 `json:"id,omitempty"`
	// The address that registered the query.
	Owner string `json:"owner,omitempty"`
	// The JSON encoded data of the query.
	QueryData string `json:"query_data,omitempty"`
	// The query type identifier (i.e. /cosmos.staking.v1beta1.Query/AllDelegations).
	QueryType string `json:"query_type,omitempty"`
	// The chain of interest identifier.
	ZoneId string `json:"zone_id,omitempty"`
	// The IBC connection ID for getting ConsensusState to verify proofs.
	ConnectionId string `json:"connection_id,omitempty"`
	// Parameter that defines how often the query must be updated.
	UpdatePeriod uint64 `json:"update_period,omitempty"`
	// The local height when the event to update the query result was emitted last time.
	LastEmittedHeight uint64 `json:"last_emitted_height,omitempty"`
	// The local chain last block height when the query result was updated.
	LastSubmittedResultLocalHeight uint64 `json:"last_submitted_result_local_height,omitempty"`
	// The remote chain last block height when the query result was updated.
	LastSubmittedResultRemoteHeight uint64 `json:"last_submitted_result_remote_height,omitempty"`
}

// Query response for an interchain account address
type QueryInterchainAccountAddressResponse struct {
	// The corresponding interchain account address on the host chain
	InterchainAccountAddress string `json:"interchain_account_address,omitempty"`
}

type QueryRegisteredQueryResultResponse struct {
	Result *QueryResult `json:"result,omitempty"`
}

type QueryResult struct {
	KvResults []*StorageValue `json:"kv_results,omitempty"`
	Height    uint64          `json:"height,omitempty"`
	Revision  uint64          `json:"revision,omitempty"`
}

type StorageValue struct {
	StoragePrefix string `json:"storage_prefix,omitempty"`
	Key           []byte `json:"key,omitempty"`
	Value         []byte `json:"value,omitempty"`
}
