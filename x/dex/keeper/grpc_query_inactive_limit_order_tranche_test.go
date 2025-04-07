package keeper_test

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v6/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestInactiveLimitOrderTrancheQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	msgs := createNInactiveLimitOrderTranche(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetInactiveLimitOrderTrancheRequest
		response *types.QueryGetInactiveLimitOrderTrancheResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetInactiveLimitOrderTrancheRequest{
				PairId:     msgs[0].Key.TradePairId.MustPairID().CanonicalString(),
				TokenIn:    msgs[0].Key.TradePairId.MakerDenom,
				TickIndex:  msgs[0].Key.TickIndexTakerToMaker,
				TrancheKey: msgs[0].Key.TrancheKey,
			},
			response: &types.QueryGetInactiveLimitOrderTrancheResponse{InactiveLimitOrderTranche: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetInactiveLimitOrderTrancheRequest{
				PairId:     msgs[1].Key.TradePairId.MustPairID().CanonicalString(),
				TokenIn:    msgs[1].Key.TradePairId.MakerDenom,
				TickIndex:  msgs[1].Key.TickIndexTakerToMaker,
				TrancheKey: msgs[1].Key.TrancheKey,
			},
			response: &types.QueryGetInactiveLimitOrderTrancheResponse{InactiveLimitOrderTranche: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetInactiveLimitOrderTrancheRequest{
				PairId:     "TokenZ<>TokenQ",
				TokenIn:    strconv.Itoa(100000),
				TickIndex:  100000,
				TrancheKey: "100000",
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.InactiveLimitOrderTranche(ctx, tc.request)
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

func TestInactiveLimitOrderTrancheQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	msgs := createNInactiveLimitOrderTranche(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllInactiveLimitOrderTrancheRequest {
		return &types.QueryAllInactiveLimitOrderTrancheRequest{
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
			resp, err := keeper.InactiveLimitOrderTrancheAll(ctx, request(nil, uint64(i), uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.InactiveLimitOrderTranche), step)
			require.Subset(t,
				msgs,
				resp.InactiveLimitOrderTranche,
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.InactiveLimitOrderTrancheAll(ctx, request(next, 0, uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.InactiveLimitOrderTranche), step)
			require.Subset(t,
				msgs,
				resp.InactiveLimitOrderTranche,
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.InactiveLimitOrderTrancheAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
		require.ElementsMatch(t,
			msgs,
			resp.InactiveLimitOrderTranche,
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.InactiveLimitOrderTrancheAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
