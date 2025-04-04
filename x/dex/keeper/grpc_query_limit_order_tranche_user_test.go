package keeper_test

import (
	"strconv"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"
	"github.com/neutron-org/neutron/v6/testutil/common/sample"
	keepertest "github.com/neutron-org/neutron/v6/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func TestLimitOrderTrancheUserQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
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
			response, err := keeper.LimitOrderTrancheUser(ctx, tc.request)
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

func TestLimitOrderTrancheUserQuerySingleWithdrawableShares(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	msgs := createNLimitOrderTrancheUser(keeper, ctx, 2)
	tranches := createNLimitOrderTranches(keeper, ctx, 2)

	tranches[0].TotalMakerDenom = math.NewInt(100)
	tranches[0].TotalTakerDenom = math.NewInt(50)
	tranches[0].ReservesTakerDenom = math.NewInt(50)

	keeper.SetLimitOrderTranche(ctx, tranches[0])

	tranches[1].TotalMakerDenom = math.NewInt(100)
	keeper.SetLimitOrderTranche(ctx, tranches[1])

	ZERO := math.ZeroInt()
	FIFTY := math.NewInt(50)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetLimitOrderTrancheUserRequest
		response *types.QueryGetLimitOrderTrancheUserResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetLimitOrderTrancheUserRequest{
				TrancheKey:             msgs[0].TrancheKey,
				Address:                msgs[0].Address,
				CalcWithdrawableShares: true,
			},
			response: &types.QueryGetLimitOrderTrancheUserResponse{LimitOrderTrancheUser: msgs[0], WithdrawableShares: &FIFTY},
		},
		{
			desc: "Second",
			request: &types.QueryGetLimitOrderTrancheUserRequest{
				TrancheKey:             msgs[1].TrancheKey,
				Address:                msgs[1].Address,
				CalcWithdrawableShares: true,
			},
			response: &types.QueryGetLimitOrderTrancheUserResponse{LimitOrderTrancheUser: msgs[1], WithdrawableShares: &ZERO},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.LimitOrderTrancheUser(ctx, tc.request)
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
			resp, err := keeper.LimitOrderTrancheUserAll(ctx, request(nil, uint64(i), uint64(step), false)) //nolint:gosec
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
			resp, err := keeper.LimitOrderTrancheUserAll(ctx, request(next, 0, uint64(step), false)) //nolint:gosec
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
		resp, err := keeper.LimitOrderTrancheUserAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
		require.ElementsMatch(t,
			msgs,
			resp.LimitOrderTrancheUser,
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LimitOrderTrancheUserAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

func TestLimitOrderTrancheUserAllByAddress(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	address := sample.AccAddress()
	msgs := createNLimitOrderTrancheUserWithAddress(keeper, ctx, address, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllLimitOrderTrancheUserByAddressRequest {
		return &types.QueryAllLimitOrderTrancheUserByAddressRequest{
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
			resp, err := keeper.LimitOrderTrancheUserAllByAddress(ctx, request(nil, uint64(i), uint64(step), false)) //nolint:gosec
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
			resp, err := keeper.LimitOrderTrancheUserAllByAddress(ctx, request(next, 0, uint64(step), false)) //nolint:gosec
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
		resp, err := keeper.LimitOrderTrancheUserAllByAddress(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
		require.ElementsMatch(t,
			msgs,
			resp.LimitOrders,
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LimitOrderTrancheUserAllByAddress(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
