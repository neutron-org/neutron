package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/neutron-org/neutron/testutil/contractmanager/keeper"
	"github.com/neutron-org/neutron/testutil/contractmanager/nullify"
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestFailureQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.ContractmanagerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNFailure(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetFailureRequest
		response *types.QueryGetFailureResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetFailureRequest{
				Address: msgs[0].Address,
			},
			response: &types.QueryGetFailureResponse{Failures: []types.Failure{msgs[0]}},
		},
		{
			desc: "Second",
			request: &types.QueryGetFailureRequest{
				Address: msgs[1].Address,
			},
			response: &types.QueryGetFailureResponse{Failures: []types.Failure{msgs[1]}},
		},
		{
			desc: "KeyIsAbsent",
			request: &types.QueryGetFailureRequest{
				Address: strconv.Itoa(100000),
			},
			response: &types.QueryGetFailureResponse{Failures: []types.Failure{}},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Failure(wctx, tc.request)
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

func TestFailureQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.ContractmanagerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNFailure(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllFailureRequest {
		return &types.QueryAllFailureRequest{
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
			resp, err := keeper.AllFailures(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Failure), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Failure),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.AllFailures(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Failure), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Failure),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.AllFailures(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Failure),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.AllFailures(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
