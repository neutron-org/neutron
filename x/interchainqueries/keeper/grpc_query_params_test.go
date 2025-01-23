package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v5/testutil/interchainqueries/keeper"
	"github.com/neutron-org/neutron/v5/x/interchainqueries/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
