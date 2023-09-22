package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/dex/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Pool(
	goCtx context.Context,
	req *types.QueryPoolRequest,
) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pairID, err := types.NewPairIDFromCanonicalString(req.PairID)
	if err != nil {
		return nil, err
	}

	pool, found := k.GetPool(ctx, pairID, req.TickIndex, req.Fee)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryPoolResponse{Pool: pool}, nil
}

func (k Keeper) PoolByID(
	goCtx context.Context,
	req *types.QueryPoolByIDRequest,
) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pool, found := k.GetPoolByID(ctx, req.PoolID)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryPoolResponse{Pool: pool}, nil
}
