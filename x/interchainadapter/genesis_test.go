package interchainadapter_test

import (
	"testing"

	keepertest "github.com/lidofinance/interchain-adapter/testutil/keeper"
	"github.com/lidofinance/interchain-adapter/testutil/nullify"
	"github.com/lidofinance/interchain-adapter/x/interchainadapter"
	"github.com/lidofinance/interchain-adapter/x/interchainadapter/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.InterchainadapterKeeper(t)
	interchainadapter.InitGenesis(ctx, *k, genesisState)
	got := interchainadapter.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
