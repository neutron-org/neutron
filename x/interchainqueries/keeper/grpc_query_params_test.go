package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testkeeper "github.com/neutron-org/neutron/testutil/interchainqueries/keeper"
	"github.com/neutron-org/neutron/x/interchainqueries/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
