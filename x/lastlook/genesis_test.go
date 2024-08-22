package lastlook_test

import (
	"testing"

	"github.com/neutron-org/neutron/v4/x/lastlook"

	"github.com/neutron-org/neutron/v4/app/config"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/testutil/common/nullify"
	"github.com/neutron-org/neutron/v4/testutil/lastlook/keeper"
	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

func TestGenesis(t *testing.T) {
	_ = config.GetDefaultConfig()

	params := types.DefaultParams()

	genesisState := types.GenesisState{
		Params: params,
	}

	k, ctx := keeper.LastLookKeeper(t)
	lastlook.InitGenesis(ctx, *k, genesisState)
	got := lastlook.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, genesisState, *got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
