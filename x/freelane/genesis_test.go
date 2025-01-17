package freelane_test

import (
	"testing"

	"github.com/neutron-org/neutron/v5/x/freelane"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v5/testutil/common/nullify"
	"github.com/neutron-org/neutron/v5/testutil/freelane/keeper"
	"github.com/neutron-org/neutron/v5/x/freelane/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keeper.FreeLaneKeeper(t)
	freelane.InitGenesis(ctx, *k, genesisState)
	got := freelane.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, genesisState, *got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
