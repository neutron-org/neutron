package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v4/testutil/dynamicfees/keeper"
	"github.com/neutron-org/neutron/v4/x/dynamicfees/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.DynamicFeesKeeper(t)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
