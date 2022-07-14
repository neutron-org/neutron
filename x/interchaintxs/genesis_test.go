package interchaintxs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/testutil/interchaintxs/keeper"
	"github.com/neutron-org/neutron/testutil/interchaintxs/nullify"
	"github.com/neutron-org/neutron/x/interchaintxs"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.InterchainTxsKeeper(t)
	interchaintxs.InitGenesis(ctx, *k, genesisState)
	got := interchaintxs.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
