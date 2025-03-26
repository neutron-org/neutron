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

// Returns all ACTIVE limit order tranches for a given pairID/tokenIn combination
// Does NOT return inactiveLimitOrderTranches
func (k Keeper) LimitOrderTrancheAll(
	c context.Context,
	req *types.QueryAllLimitOrderTrancheRequest,
) (*types.QueryAllLimitOrderTrancheResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var limitOrderTranches []*types.LimitOrderTranche
	ctx := sdk.UnwrapSDKContext(c)

	pairID, err := types.NewPairIDFromCanonicalString(req.PairId)
	if err != nil {
		return nil, err
	}
	tradePairID := types.NewTradePairIDFromMaker(pairID, req.TokenIn)

	store := ctx.KVStore(k.storeKey)
	limitOrderTrancheStore := prefix.NewStore(store, types.TickLiquidityPrefix(tradePairID))

	pageRes, err := query.FilteredPaginate(
		limitOrderTrancheStore,
		req.Pagination, func(_, value []byte, accum bool) (hit bool, err error) {
			var tick types.TickLiquidity

			if err := k.cdc.Unmarshal(value, &tick); err != nil {
				return false, err
			}
			tranche := tick.GetLimitOrderTranche()
			// Check if this is a LimitOrderTranche and not PoolReserves
			if tranche != nil {
				if accum {
					limitOrderTranches = append(limitOrderTranches, tranche)
				}

				return true, nil
			}

			return false, nil
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllLimitOrderTrancheResponse{LimitOrderTranche: limitOrderTranches, Pagination: pageRes}, nil
}

// Returns a specific limit order tranche either from the tickLiquidity index or from the FillLimitOrderTranche index
func (k Keeper) LimitOrderTranche(
	c context.Context,
	req *types.QueryGetLimitOrderTrancheRequest,
) (*types.QueryGetLimitOrderTrancheResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	pairID, err := types.NewPairIDFromCanonicalString(req.PairId)
	if err != nil {
		return nil, err
	}
	tradePairID := types.NewTradePairIDFromMaker(pairID, req.TokenIn)
	val, _, found := k.FindLimitOrderTranche(
		ctx,
		&types.LimitOrderTrancheKey{
			TradePairId:           tradePairID,
			TickIndexTakerToMaker: req.TickIndex,
			TrancheKey:            req.TrancheKey,
		},
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetLimitOrderTrancheResponse{LimitOrderTranche: val}, nil
}
