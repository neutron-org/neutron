package wasmbinding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/interchainqueries/types"
	icatypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

// GetInterchainQueryResult is a function, not method, so the message_plugin can use it
func (qp *QueryPlugin) GetInterchainQueryResult(ctx sdk.Context, queryID uint64) (*types.QueryResult, error) {
	return qp.icqKeeper.GetQueryResultByID(ctx, queryID)
}

func (qp *QueryPlugin) GetInterchainAccountAddress(ctx sdk.Context, req *icatypes.QueryInterchainAccountAddressRequest) (*icatypes.QueryInterchainAccountAddressResponse, error) {
	return qp.icaControllerKeeper.InterchainAccountAddress(sdk.WrapSDKContext(ctx), req)
}

func (qp *QueryPlugin) GetRegisteredInterchainQueries(ctx sdk.Context, req *types.QueryRegisteredQueriesRequest) (*types.QueryRegisteredQueriesResponse, error) {
	return qp.icqKeeper.GetRegisteredQueries(ctx, req)
}

func (qp *QueryPlugin) GetRegisteredInterchainQuery(ctx sdk.Context, req *types.QueryRegisteredQueryRequest) (*types.RegisteredQuery, error) {
	return qp.icqKeeper.GetQueryByID(ctx, req.QueryId)
}
