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

func TestPoolMetadataQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	msgs := createNPoolMetadata(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetPoolMetadataRequest
		response *types.QueryGetPoolMetadataResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetPoolMetadataRequest{Id: msgs[0].Id},
			response: &types.QueryGetPoolMetadataResponse{PoolMetadata: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetPoolMetadataRequest{Id: msgs[1].Id},
			response: &types.QueryGetPoolMetadataResponse{PoolMetadata: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetPoolMetadataRequest{Id: uint64(len(msgs))},
			err:     status.Error(codes.NotFound, "PoolMetadata not found for key"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.PoolMetadata(ctx, tc.request)
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

func TestPoolMetadataQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	msgs := createNPoolMetadata(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllPoolMetadataRequest {
		return &types.QueryAllPoolMetadataRequest{
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
			resp, err := keeper.PoolMetadataAll(ctx, request(nil, uint64(i), uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PoolMetadata), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.PoolMetadata),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.PoolMetadataAll(ctx, request(next, 0, uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PoolMetadata), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.PoolMetadata),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.PoolMetadataAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.PoolMetadata),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.PoolMetadataAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
