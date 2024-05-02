package feeburner_test

import (
	"testing"

	"github.com/neutron-org/neutron/v4/app/config"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/testutil/common/nullify"
	"github.com/neutron-org/neutron/v4/testutil/feeburner/keeper"
	"github.com/neutron-org/neutron/v4/x/feeburner"
	"github.com/neutron-org/neutron/v4/x/feeburner/types"
)

func TestGenesis(t *testing.T) {
	_ = config.GetDefaultConfig()

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keeper.FeeburnerKeeper(t)
	feeburner.InitGenesis(ctx, *k, genesisState)
	got := feeburner.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
