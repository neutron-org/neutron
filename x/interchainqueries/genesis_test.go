package interchainqueries_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/lidofinance/interchain-adapter/testutil/interchainqueries/keeper"
	"github.com/lidofinance/interchain-adapter/testutil/interchainqueries/nullify"
	"github.com/lidofinance/interchain-adapter/x/interchainqueries"
	"github.com/lidofinance/interchain-adapter/x/interchainqueries/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.InterchainadapterKeeper(t)
	interchainqueries.InitGenesis(ctx, *k, genesisState)
	got := interchainqueries.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
