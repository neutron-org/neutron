package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (k Keeper) PoolReservesAll(
	goCtx context.Context,
	req *types.QueryAllPoolReservesRequest,
) (*types.QueryAllPoolReservesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pairID, err := types.NewPairIDFromCanonicalString(req.PairId)
	if err != nil {
		return nil, err
	}
	tradePairID := types.NewTradePairIDFromMaker(pairID, req.TokenIn)

	store := ctx.KVStore(k.storeKey)
	PoolReservesStore := prefix.NewStore(store, types.TickLiquidityPrefix(tradePairID))

	var poolReserves []*types.PoolReserves
	pageRes, err := query.FilteredPaginate(
		PoolReservesStore,
		req.Pagination,
		func(_, value []byte, accum bool) (hit bool, err error) {
			var tick types.TickLiquidity

			if err := k.cdc.Unmarshal(value, &tick); err != nil {
				return false, err
			}
			reserves := tick.GetPoolReserves()
			// Check if this is a LimitOrderTranche and not PoolReserves
			if reserves != nil {
				if accum {
					poolReserves = append(poolReserves, reserves)
				}

				return true, nil
			}

			return false, nil
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPoolReservesResponse{PoolReserves: poolReserves, Pagination: pageRes}, nil
}

func (k Keeper) PoolReserves(
	goCtx context.Context,
	req *types.QueryGetPoolReservesRequest,
) (*types.QueryGetPoolReservesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	pairID, err := types.NewPairIDFromCanonicalString(req.PairId)
	if err != nil {
		return nil, err
	}
	tradePairID := types.NewTradePairIDFromMaker(pairID, req.TokenIn)

	poolReservesID := &types.PoolReservesKey{
		TradePairId:           tradePairID,
		TickIndexTakerToMaker: req.TickIndex,
		Fee:                   req.Fee,
	}
	val, found := k.GetPoolReserves(ctx, poolReservesID)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetPoolReservesResponse{PoolReserves: val}, nil
}
