package interchaintxs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v5/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v5/testutil/interchaintxs/keeper"
	"github.com/neutron-org/neutron/v5/x/interchaintxs"
	"github.com/neutron-org/neutron/v5/x/interchaintxs/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.InterchainTxsKeeper(t, nil, nil, nil, nil, nil, nil, nil)
	interchaintxs.InitGenesis(ctx, *k, genesisState)
	got := interchaintxs.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
