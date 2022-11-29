package bindings

import (
	"encoding/json"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/neutron-org/neutron/x/interchainqueries/types"
)

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
	QueryID uint64 `json:"query_id,omitempty"`
}

type QueryInterchainAccountAddressRequest struct {
	// owner_address is the owner of the interchain account on the controller chain
	OwnerAddress string `json:"owner_address,omitempty"`
	// interchain_account_id is an identifier of your interchain account from which you want to execute msgs
	InterchainAccountID string `json:"interchain_account_id,omitempty"`
	// connection_id is an IBC connection identifier between Neutron and remote chain
	ConnectionID string `json:"connection_id,omitempty"`
}

type QueryRegisteredQueriesRequest struct {
	Owners       []string           `json:"owners,omitempty"`
	ConnectionID string             `json:"connection_id,omitempty"`
	Pagination   *query.PageRequest `json:"pagination,omitempty"`
}

type QueryRegisteredQueryRequest struct {
	QueryID uint64 `json:"query_id,omitempty"`
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
	ID uint64 `json:"id"`
	// The address that registered the query.
	Owner string `json:"owner"`
	// The KV-storage keys for which we want to get values from remote chain
	Keys []*types.KVKey `json:"keys"`
	// The filter for transaction search ICQ
	TransactionsFilter string `json:"transactions_filter"`
	// The query type identifier (i.e. 'kv' or 'tx' for now).
	QueryType string `json:"query_type"`
	// The IBC connection ID for getting ConsensusState to verify proofs.
	ConnectionID string `json:"connection_id"`
	// Parameter that defines how often the query must be updated.
	UpdatePeriod uint64 `json:"update_period"`
	// The local chain last block height when the query result was updated.
	LastSubmittedResultLocalHeight uint64 `json:"last_submitted_result_local_height"`
	// The remote chain last block height when the query result was updated.
	LastSubmittedResultRemoteHeight uint64 `json:"last_submitted_result_remote_height"`
	// Amount of coins deposited for the query.
	Deposit sdktypes.Coins `json:"deposit"`
	// Timeout before query becomes available for everybody to remove.
	SubmitTimeout uint64 `json:"submit_timeout"`
}

func (rq RegisteredQuery) MarshalJSON() ([]byte, error) {
	type AliasRQ RegisteredQuery

	a := struct {
		AliasRQ
	}{
		AliasRQ: (AliasRQ)(rq),
	}

	// We want keys be as empty array in Json ('[]'), not 'null'
	// It's easier to work with on smart-contracts side
	if a.Keys == nil {
		a.Keys = make([]*types.KVKey, 0)
	}

	return json.Marshal(a)
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
	Key           []byte `json:"key"`
	Value         []byte `json:"value"`
}

func (sv StorageValue) MarshalJSON() ([]byte, error) {
	type AliasSV StorageValue

	a := struct {
		AliasSV
	}{
		AliasSV: (AliasSV)(sv),
	}

	// We want Key and Value be as empty arrays in Json ('[]'), not 'null'
	// It's easier to work with on smart-contracts side
	if a.Key == nil {
		a.Key = make([]byte, 0)
	}
	if a.Value == nil {
		a.Value = make([]byte, 0)
	}

	return json.Marshal(a)
}
