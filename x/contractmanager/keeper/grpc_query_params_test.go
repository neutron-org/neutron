package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testkeeper "github.com/neutron-org/neutron/testutil/contractmanager/keeper"
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.ContractManagerKeeper(t, nil)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
