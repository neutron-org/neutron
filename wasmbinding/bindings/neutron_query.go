package bindings

import (
	"github.com/neutron-org/neutron/x/interchainqueries/types"
	icatypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

// NeutronQuery contains neutron custom queries.
type NeutronQuery struct {
	/// Registered Interchain Query Result for specified QueryID
	InterchainQueryResult *types.QueryRegisteredQueryResultRequest `json:"interchain_query_result,omitempty"`
	/// Interchain account address for specified ConnectionID and OwnerAddress
	InterchainAccountAddress *icatypes.QueryInterchainAccountAddressRequest `json:"interchain_account_address,omitempty"`
	/// RegisteredInterchainQueries
	RegisteredInterchainQueries *types.QueryRegisteredQueriesRequest `json:"registered_interchain_queries,omitempty"`
	/// RegisteredInterchainQueries
	RegisteredInterchainQuery *types.QueryRegisteredQueryRequest `json:"registered_interchain_query,omitempty"`
}
