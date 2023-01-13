package keeper_test

import (
	"testing"

	"github.com/neutron-org/neutron/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/neutron-org/neutron/testutil/feeburner/keeper"
	"github.com/neutron-org/neutron/x/feeburner/types"
	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	_ = app.GetDefaultConfig()

	keeper, ctx := testkeeper.FeeburnerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
