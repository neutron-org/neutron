package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/neutron-org/neutron/v3/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v3/testutil/dex/nullify"
	"github.com/neutron-org/neutron/v3/x/dex/types"
)

func TestPoolQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNPools(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryPoolRequest
		response *types.QueryPoolResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryPoolRequest{
				PairId:    "TokenA<>TokenB",
				TickIndex: msgs[0].CenterTickIndex(),
				Fee:       msgs[0].Fee(),
			},
			response: &types.QueryPoolResponse{Pool: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryPoolRequest{
				PairId:    "TokenA<>TokenB",
				TickIndex: msgs[1].CenterTickIndex(),
				Fee:       msgs[1].Fee(),
			},
			response: &types.QueryPoolResponse{Pool: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryPoolRequest{
				PairId:    "TokenA<>TokenB",
				TickIndex: 0,
				Fee:       100000,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Pool(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestPoolQueryByID(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNPools(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryPoolByIDRequest
		response *types.QueryPoolResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryPoolByIDRequest{
				PoolId: 0,
			},
			response: &types.QueryPoolResponse{Pool: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryPoolByIDRequest{
				PoolId: 1,
			},
			response: &types.QueryPoolResponse{Pool: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryPoolByIDRequest{
				PoolId: 100,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.PoolByID(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}
