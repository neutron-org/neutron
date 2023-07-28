package keeper_test

import (
	"testing"

	testkeeper "github.com/neutron-org/neutron/testutil/cron/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/cron/types"
	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.CronKeeper(t, nil, nil)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)
	
	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
