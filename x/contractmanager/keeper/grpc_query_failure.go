package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/neutron-org/neutron/x/contractmanager/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) AllFailures(c context.Context, req *types.QueryAllFailureRequest) (*types.QueryAllFailureResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var failures []types.Failure
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	failureStore := prefix.NewStore(store, types.ContractFailuresKey)

	pageRes, err := query.Paginate(failureStore, req.Pagination, func(key []byte, value []byte) error {
		var failure types.Failure
		if err := k.cdc.Unmarshal(value, &failure); err != nil {
			return err
		}

		failures = append(failures, failure)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllFailureResponse{Failures: failures, Pagination: pageRes}, nil
}

func (k Keeper) Failure(c context.Context, req *types.QueryGetFailureRequest) (*types.QueryGetFailureResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse address: %s", req.Address)
	}

	ctx := sdk.UnwrapSDKContext(c)

	val := k.GetContractFailures(
		ctx,
		req.Address,
	)

	return &types.QueryGetFailureResponse{Failures: val}, nil
}
