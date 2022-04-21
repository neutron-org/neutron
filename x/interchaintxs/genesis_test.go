package interchaintxs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/lidofinance/interchain-adapter/testutil/interchaintxs/keeper"
	"github.com/lidofinance/interchain-adapter/testutil/interchaintxs/nullify"
	"github.com/lidofinance/interchain-adapter/x/interchaintxs"
	"github.com/lidofinance/interchain-adapter/x/interchaintxs/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.InterchainadapterKeeper(t)
	interchaintxs.InitGenesis(ctx, *k, genesisState)
	got := interchaintxs.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
