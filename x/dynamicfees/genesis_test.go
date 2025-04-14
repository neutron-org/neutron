package dynamicfees_test

import (
	"testing"

	"cosmossdk.io/math"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dynamicfees"

	"github.com/neutron-org/neutron/v6/app/config"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"
	"github.com/neutron-org/neutron/v6/testutil/dynamicfees/keeper"
	"github.com/neutron-org/neutron/v6/x/dynamicfees/types"
)

func TestGenesis(t *testing.T) {
	_ = config.GetDefaultConfig()

	params := types.DefaultParams()
	params.NtrnPrices = append(params.NtrnPrices, cosmostypes.DecCoin{Denom: "uatom", Amount: math.LegacyMustNewDecFromStr("10")})

	genesisState := types.GenesisState{
		Params: params,
	}

	k, ctx := keeper.DynamicFeesKeeper(t)
	dynamicfees.InitGenesis(ctx, *k, genesisState)
	got := dynamicfees.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, genesisState, *got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
