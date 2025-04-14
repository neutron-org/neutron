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

func (k Keeper) InactiveLimitOrderTrancheAll(
	c context.Context,
	req *types.QueryAllInactiveLimitOrderTrancheRequest,
) (*types.QueryAllInactiveLimitOrderTrancheResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var inactiveLimitOrderTranches []*types.LimitOrderTranche
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	inactiveLimitOrderTrancheStore := prefix.NewStore(store, types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix))

	pageRes, err := query.Paginate(inactiveLimitOrderTrancheStore, req.Pagination, func(_, value []byte) error {
		inactiveLimitOrderTranche := &types.LimitOrderTranche{}
		if err := k.cdc.Unmarshal(value, inactiveLimitOrderTranche); err != nil {
			return err
		}

		inactiveLimitOrderTranches = append(inactiveLimitOrderTranches, inactiveLimitOrderTranche)

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllInactiveLimitOrderTrancheResponse{
		InactiveLimitOrderTranche: inactiveLimitOrderTranches,
		Pagination:                pageRes,
	}, nil
}

func (k Keeper) InactiveLimitOrderTranche(
	c context.Context,
	req *types.QueryGetInactiveLimitOrderTrancheRequest,
) (*types.QueryGetInactiveLimitOrderTrancheResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	pairID, err := types.NewPairIDFromCanonicalString(req.PairId)
	if err != nil {
		return nil, err
	}
	tradePairID := types.NewTradePairIDFromMaker(pairID, req.TokenIn)
	val, found := k.GetInactiveLimitOrderTranche(
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

	return &types.QueryGetInactiveLimitOrderTrancheResponse{InactiveLimitOrderTranche: val}, nil
}
