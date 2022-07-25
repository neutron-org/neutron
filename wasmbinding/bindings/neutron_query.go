package bindings

import "github.com/neutron-org/neutron/x/interchainqueries/types"

// NeutronQuery contains neutron custom queries.
type NeutronQuery struct {
	/// Registered Interchain Query Result for specified
	InterchainQueryResult    *InterchainQueryResult    `json:"interchain_query_result,omitempty"`
	InterchainAccountAddress *InterchainAccountAddress `json:"interchain_account_address,omitempty"`
}

type InterchainQueryResult struct {
	QueryID uint64 `json:"query_id"`
}

type InterchainQueryResultResponse struct {
	Result *types.QueryResult `json:"result,omitempty"` // TODO: real result type
}

type InterchainAccountAddress struct {
	ConnectionID string `json:"connection_id"`
	OwnerAddress string `json:"owner_address"`
}

type InterchainAccountAddressResponse struct {
	InterchainAccountAddress *string `json:"interchain_account_address,omitempty"`
}
