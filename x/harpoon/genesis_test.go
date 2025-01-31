package harpoon_test

import (
	"testing"

	"github.com/neutron-org/neutron/v5/x/harpoon"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v5/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

const (
	ContractAddress1 = "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a"
	ContractAddress2 = "neutron1u9dulasrfe6laxwwjx83njhm5as466uz43arfheq79k8eqahqhasf3tj94"
)

func TestGenesis(t *testing.T) {
	// nil state genesis works
	genesisState := types.GenesisState{
		HookSubscriptions: nil,
	}

	k, ctx := keepertest.HarpoonKeeper(t, nil)
	harpoon.InitGenesis(ctx, k, genesisState)
	got := harpoon.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, &genesisState, got)

	// non nil genesis -  hook subscriptions get set properly
	genesisState2 := types.GenesisState{
		HookSubscriptions: []types.HookSubscriptions{
			{
				HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_BONDED,
				ContractAddresses: []string{ContractAddress1},
			},
			{
				HookType:          types.HOOK_TYPE_BEFORE_DELEGATION_REMOVED,
				ContractAddresses: []string{ContractAddress1, ContractAddress2},
			},
		},
	}
	k, ctx = keepertest.HarpoonKeeper(t, nil)
	harpoon.InitGenesis(ctx, k, genesisState2)
	got2 := harpoon.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, &genesisState2, got2)
}
