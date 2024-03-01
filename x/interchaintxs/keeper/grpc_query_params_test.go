package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v3/testutil/interchaintxs/keeper"
	"github.com/neutron-org/neutron/v3/x/interchaintxs/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.InterchainTxsKeeper(t, nil, nil, nil, nil, nil, nil)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := keeper.Params(wctx, nil)
	require.Error(t, err)
	require.Nil(t, response)

	response, err = keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
