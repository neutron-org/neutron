package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v3/testutil/contractmanager/keeper"
	"github.com/neutron-org/neutron/v3/x/contractmanager/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.ContractManagerKeeper(t, nil)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
