package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/neutron-org/neutron/x/cron/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ScheduleAll(c context.Context, req *types.QueryAllScheduleRequest) (*types.QueryAllScheduleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var schedules []types.Schedule
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	scheduleStore := prefix.NewStore(store, types.ScheduleKey) // TODO: works?

	pageRes, err := query.Paginate(scheduleStore, req.Pagination, func(key []byte, value []byte) error {
		var schedule types.Schedule
		if err := k.cdc.Unmarshal(value, &schedule); err != nil {
			return err
		}

		schedules = append(schedules, schedule)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllScheduleResponse{Schedule: schedules, Pagination: pageRes}, nil
}

func (k Keeper) Schedule(c context.Context, req *types.QueryGetScheduleRequest) (*types.QueryGetScheduleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetSchedule(
		ctx,
		req.Name,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetScheduleResponse{Schedule: *val}, nil
}
