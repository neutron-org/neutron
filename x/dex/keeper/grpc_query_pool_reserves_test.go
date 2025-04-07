package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v6/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func TestPoolReservesQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	msgs := createNPoolReserves(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetPoolReservesRequest
		response *types.QueryGetPoolReservesResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetPoolReservesRequest{
				PairId:    "TokenA<>TokenB",
				TickIndex: msgs[0].Key.TickIndexTakerToMaker,
				TokenIn:   "TokenA",
				Fee:       msgs[0].Key.Fee,
			},
			response: &types.QueryGetPoolReservesResponse{PoolReserves: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetPoolReservesRequest{
				PairId:    "TokenA<>TokenB",
				TickIndex: msgs[1].Key.TickIndexTakerToMaker,
				TokenIn:   "TokenA",
				Fee:       msgs[1].Key.Fee,
			},
			response: &types.QueryGetPoolReservesResponse{PoolReserves: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetPoolReservesRequest{
				PairId:    "TokenA<>TokenB",
				TickIndex: 0,
				TokenIn:   "TokenA",
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
			response, err := keeper.PoolReserves(ctx, tc.request)
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

func TestPoolReservesQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	msgs := createNPoolReserves(keeper, ctx, 2)
	// Add more data to make sure only LO tranches are returned
	createNLimitOrderTranches(keeper, ctx, 2)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllPoolReservesRequest {
		return &types.QueryAllPoolReservesRequest{
			PairId:  "TokenA<>TokenB",
			TokenIn: "TokenA",
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
			resp, err := keeper.PoolReservesAll(ctx, request(nil, uint64(i), uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PoolReserves), step)
			require.Subset(t,
				msgs,
				resp.PoolReserves,
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.PoolReservesAll(ctx, request(next, 0, uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PoolReserves), step)
			require.Subset(t,
				msgs,
				resp.PoolReserves,
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.PoolReservesAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
		require.ElementsMatch(t,
			msgs,
			resp.PoolReserves,
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.PoolReservesAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
