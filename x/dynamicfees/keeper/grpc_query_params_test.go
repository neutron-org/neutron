package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v6/testutil/dynamicfees/keeper"
	"github.com/neutron-org/neutron/v6/x/dynamicfees/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.DynamicFeesKeeper(t)
	params := types.DefaultParams()
	params.NtrnPrices = append(params.NtrnPrices, cosmostypes.DecCoin{Denom: "uatom", Amount: math.LegacyMustNewDecFromStr("10")})
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
