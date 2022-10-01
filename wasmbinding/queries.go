package wasmbinding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/neutron-org/neutron/wasmbinding/bindings"
	"github.com/neutron-org/neutron/x/interchainqueries/types"
	icatypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

func (qp *QueryPlugin) GetInterchainQueryResult(ctx sdk.Context, queryID uint64) (*bindings.QueryRegisteredQueryResultResponse, error) {
	grpcResp, err := qp.icqKeeper.GetQueryResultByID(ctx, queryID)
	if err != nil {
		return nil, err
	}
	resp := bindings.QueryResult{
		KvResults: make([]*bindings.StorageValue, 0, len(grpcResp.KvResults)),
		Height:    grpcResp.GetHeight(),
		Revision:  grpcResp.GetRevision(),
	}
	for _, grpcKv := range grpcResp.GetKvResults() {
		kv := bindings.StorageValue{
			StoragePrefix: grpcKv.GetStoragePrefix(),
			Key:           grpcKv.GetKey(),
			Value:         grpcKv.GetValue(),
		}
		resp.KvResults = append(resp.KvResults, &kv)
	}

	return &bindings.QueryRegisteredQueryResultResponse{Result: &resp}, nil
}

func (qp *QueryPlugin) GetInterchainAccountAddress(ctx sdk.Context, req *bindings.QueryInterchainAccountAddressRequest) (*bindings.QueryInterchainAccountAddressResponse, error) {
	grpcReq := icatypes.QueryInterchainAccountAddressRequest{
		OwnerAddress:        req.OwnerAddress,
		InterchainAccountId: req.InterchainAccountId,
		ConnectionId:        req.ConnectionId,
	}

	grpcResp, err := qp.icaControllerKeeper.InterchainAccountAddress(sdk.WrapSDKContext(ctx), &grpcReq)
	if err != nil {
		return nil, err
	}

	return &bindings.QueryInterchainAccountAddressResponse{InterchainAccountAddress: grpcResp.GetInterchainAccountAddress()}, nil
}

func (qp *QueryPlugin) GetRegisteredInterchainQueries(ctx sdk.Context, query *bindings.QueryRegisteredQueriesRequest) (*bindings.QueryRegisteredQueriesResponse, error) {
	grpcResp, err := qp.icqKeeper.GetRegisteredQueries(ctx, &types.QueryRegisteredQueriesRequest{
		Owners:       query.Owners,
		ConnectionId: query.ConnectionId,
		Pagination: &sdkquery.PageRequest{
			Key:        query.Pagination.Key,
			Offset:     query.Pagination.Offset,
			Limit:      query.Pagination.Limit,
			CountTotal: query.Pagination.CountTotal,
			Reverse:    query.Pagination.Reverse,
		},
	})
	if err != nil {
		return nil, err
	}

	resp := bindings.QueryRegisteredQueriesResponse{RegisteredQueries: make([]bindings.RegisteredQuery, 0, len(grpcResp.GetRegisteredQueries()))}
	for _, grpcQuery := range grpcResp.GetRegisteredQueries() {
		query := mapGRPCRegisteredQueryToWasmBindings(grpcQuery)
		resp.RegisteredQueries = append(resp.RegisteredQueries, query)
	}
	return &resp, nil
}

func (qp *QueryPlugin) GetRegisteredInterchainQuery(ctx sdk.Context, req *bindings.QueryRegisteredQueryRequest) (*bindings.QueryRegisteredQueryResponse, error) {
	grpcResp, err := qp.icqKeeper.GetQueryByID(ctx, req.QueryId)
	if err != nil {
		return nil, err
	}
	if grpcResp == nil {
		return nil, sdkerrors.Wrapf(types.ErrEmptyResult, "interchain query response empty for query id %d", req.QueryId)
	}
	query := mapGRPCRegisteredQueryToWasmBindings(*grpcResp)

	return &bindings.QueryRegisteredQueryResponse{RegisteredQuery: &query}, nil
}

func mapGRPCRegisteredQueryToWasmBindings(grpcQuery types.RegisteredQuery) bindings.RegisteredQuery {
	return bindings.RegisteredQuery{
		Id:                              grpcQuery.GetId(),
		Owner:                           grpcQuery.GetOwner(),
		Keys:                            grpcQuery.GetKeys(),
		TransactionsFilter:              grpcQuery.GetTransactionsFilter(),
		QueryType:                       grpcQuery.GetQueryType(),
		ConnectionId:                    grpcQuery.GetConnectionId(),
		UpdatePeriod:                    grpcQuery.GetUpdatePeriod(),
		LastSubmittedResultLocalHeight:  grpcQuery.GetLastSubmittedResultLocalHeight(),
		LastSubmittedResultRemoteHeight: grpcQuery.GetLastSubmittedResultRemoteHeight(),
	}
}
