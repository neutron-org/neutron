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
	keeper, ctx := keepertest.ContractManagerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNFailure(keeper, ctx, 2, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetFailuresByAddressRequest
		response *types.QueryGetFailuresByAddressResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetFailuresByAddressRequest{
				Address: msgs[0][0].Address,
			},
			response: &types.QueryGetFailuresByAddressResponse{Failures: msgs[0], Pagination: &query.PageResponse{Total: 2}},
		},
		{
			desc: "Second",
			request: &types.QueryGetFailuresByAddressRequest{
				Address: msgs[1][0].Address,
			},
			response: &types.QueryGetFailuresByAddressResponse{Failures: msgs[1], Pagination: &query.PageResponse{Total: 2}},
		},
		{
			desc: "KeyIsAbsent",
			request: &types.QueryGetFailuresByAddressRequest{
				Address: "cosmos132juzk0gdmwuxvx4phug7m3ymyatxlh9m9paea",
			},
			response: &types.QueryGetFailuresByAddressResponse{Failures: []types.Failure{}, Pagination: &query.PageResponse{Total: 0}},
		},
		{
			desc: "InvalidAddress",
			request: &types.QueryGetFailuresByAddressRequest{
				Address: "wrong_address",
			},
			err: status.Error(codes.InvalidArgument, "failed to parse address: wrong_address"),
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
	keeper, ctx := keepertest.ContractManagerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNFailure(keeper, ctx, 5, 3)
	flattenItems := flattenFailures(msgs)

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
		for i := 0; i < len(flattenItems); i += step {
			resp, err := keeper.AllFailures(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Failures), step)
			require.Subset(t,
				nullify.Fill(flattenItems),
				nullify.Fill(resp.Failures),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(flattenItems); i += step {
			resp, err := keeper.AllFailures(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Failures), step)
			require.Subset(t,
				nullify.Fill(flattenItems),
				nullify.Fill(resp.Failures),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.AllFailures(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(flattenItems), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(flattenItems),
			nullify.Fill(resp.Failures),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.AllFailures(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
