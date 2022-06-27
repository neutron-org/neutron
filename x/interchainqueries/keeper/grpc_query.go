package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RegisteredQuery(goCtx context.Context, request *types.QueryRegisteredQueryRequest) (*types.QueryRegisteredQueryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	registeredQuery, err := k.GetQueryByID(ctx, request.QueryId)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidQueryID, "failed to get registered query by query id: %v", err)
	}

	return &types.QueryRegisteredQueryResponse{RegisteredQuery: registeredQuery}, nil
}

func (k Keeper) RegisteredQueries(goCtx context.Context, _ *types.QueryRegisteredQueriesRequest) (*types.QueryRegisteredQueriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RegisteredQueryKey)
	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	queries := make([]types.RegisteredQuery, 0)
	for ; iterator.Valid(); iterator.Next() {
		query := types.RegisteredQuery{}
		k.cdc.MustUnmarshal(iterator.Value(), &query)
		queries = append(queries, query)
	}

	return &types.QueryRegisteredQueriesResponse{RegisteredQueries: queries}, nil
}

func (k Keeper) QueryResult(goCtx context.Context, request *types.QueryRegisteredQueryResultRequest) (*types.QueryRegisteredQueryResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	result, err := k.GetQueryResultByID(ctx, request.QueryId)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidQueryID, "failed to get query result by query id: %v", err)
	}

	return &types.QueryRegisteredQueryResultResponse{Result: result}, nil
}

func (k Keeper) QueryTransactions(goCtx context.Context, request *types.QuerySubmittedTransactionsRequest) (*types.QuerySubmittedTransactionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	result, err := k.GetSubmittedTransactions(ctx, request.QueryId, request.Start, request.End)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidQueryID, "failed to get query result by query id: %v", err)
	}

	return &types.QuerySubmittedTransactionsResponse{Transactions: result}, nil
}
