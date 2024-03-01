package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/neutron-org/neutron/v3/testutil/common/sample"
	keepertest "github.com/neutron-org/neutron/v3/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v3/testutil/dex/nullify"
	"github.com/neutron-org/neutron/v3/x/dex/types"
)

func TestLimitOrderTrancheUserQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNLimitOrderTrancheUser(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetLimitOrderTrancheUserRequest
		response *types.QueryGetLimitOrderTrancheUserResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetLimitOrderTrancheUserRequest{
				TrancheKey: msgs[0].TrancheKey,
				Address:    msgs[0].Address,
			},
			response: &types.QueryGetLimitOrderTrancheUserResponse{LimitOrderTrancheUser: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetLimitOrderTrancheUserRequest{
				TrancheKey: msgs[1].TrancheKey,
				Address:    msgs[1].Address,
			},
			response: &types.QueryGetLimitOrderTrancheUserResponse{LimitOrderTrancheUser: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetLimitOrderTrancheUserRequest{
				TrancheKey: "100000",
				Address:    strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.LimitOrderTrancheUser(wctx, tc.request)
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

func TestLimitOrderTrancheUserQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNLimitOrderTrancheUser(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllLimitOrderTrancheUserRequest {
		return &types.QueryAllLimitOrderTrancheUserRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LimitOrderTrancheUserAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LimitOrderTrancheUser), step)
			require.Subset(t,
				msgs,
				resp.LimitOrderTrancheUser,
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LimitOrderTrancheUserAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LimitOrderTrancheUser), step)
			require.Subset(t,
				msgs,
				resp.LimitOrderTrancheUser,
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.LimitOrderTrancheUserAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			msgs,
			resp.LimitOrderTrancheUser,
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LimitOrderTrancheUserAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

func TestLimitOrderTrancheUserAllByAddress(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	address := sample.AccAddress()
	msgs := createNLimitOrderTrancheUserWithAddress(keeper, ctx, address, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllUserLimitOrdersRequest {
		return &types.QueryAllUserLimitOrdersRequest{
			Address: address,
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LimitOrderTrancheUserAllByAddress(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LimitOrders), step)
			require.Subset(t,
				msgs,
				resp.LimitOrders,
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LimitOrderTrancheUserAllByAddress(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LimitOrders), step)
			require.Subset(t,
				msgs,
				resp.LimitOrders,
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.LimitOrderTrancheUserAllByAddress(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			msgs,
			resp.LimitOrders,
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LimitOrderTrancheUserAllByAddress(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
