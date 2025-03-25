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

// NOTE: For single queries of tick liquidity use explicty typed queries
// (ie. the k.LimitOrderTranche & k.PoolReserves)

func (k Keeper) TickLiquidityAll(
	c context.Context,
	req *types.QueryAllTickLiquidityRequest,
) (*types.QueryAllTickLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var tickLiquidities []*types.TickLiquidity
	ctx := sdk.UnwrapSDKContext(c)

	pairID, err := types.NewPairIDFromCanonicalString(req.PairId)
	if err != nil {
		return nil, err
	}

	tradePairID := types.NewTradePairIDFromMaker(pairID, req.TokenIn)

	store := ctx.KVStore(k.storeKey)
	tickLiquidityStore := prefix.NewStore(store, types.TickLiquidityPrefix(tradePairID))

	pageRes, err := query.Paginate(tickLiquidityStore, req.Pagination, func(_, value []byte) error {
		tickLiquidity := &types.TickLiquidity{}
		if err := k.cdc.Unmarshal(value, tickLiquidity); err != nil {
			return err
		}

		tickLiquidities = append(tickLiquidities, tickLiquidity)

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllTickLiquidityResponse{TickLiquidity: tickLiquidities, Pagination: pageRes}, nil
}
