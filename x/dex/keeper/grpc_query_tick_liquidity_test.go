package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/neutron-org/neutron/v6/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func TestTickLiquidityQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	msgs := CreateNTickLiquidity(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllTickLiquidityRequest {
		return &types.QueryAllTickLiquidityRequest{
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
			resp, err := keeper.TickLiquidityAll(ctx, request(nil, uint64(i), uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.TickLiquidity), step)
			require.Subset(t,
				msgs,
				resp.TickLiquidity,
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.TickLiquidityAll(ctx, request(next, 0, uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.TickLiquidity), step)
			require.Subset(t,
				msgs,
				resp.TickLiquidity,
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.TickLiquidityAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
		require.ElementsMatch(t,
			msgs,
			resp.TickLiquidity,
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.TickLiquidityAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
