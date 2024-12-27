package harpoon_test

import (
	"testing"

	"github.com/neutron-org/neutron/v5/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	harpoon "github.com/neutron-org/neutron/v5/x/harpoon/module"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.HarpoonKeeper(t, nil, nil)
	harpoon.InitGenesis(ctx, *k, genesisState)
	got := harpoon.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
