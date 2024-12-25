package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

    keepertest "github.com/neutron-org/neutron/v5/testutil/keeper"
    "github.com/neutron-org/neutron/v5/x/harpoon/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := keepertest.HarpoonKeeper(t)
	params := types.DefaultParams()
	require.NoError(t, keeper.SetParams(ctx, params))

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
