package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/testutil/interchaintxs/keeper"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.InterchainTxsKeeper(t, nil, nil, nil, nil, nil)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, nil)
	require.Error(t, err)
	require.Nil(t, response)

	response, err = keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
