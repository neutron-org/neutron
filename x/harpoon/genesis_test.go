package harpoon_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/errors"
	types2 "github.com/cosmos/cosmos-sdk/types"
	errors2 "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/golang/mock/gomock"

	"github.com/neutron-org/neutron/v6/app/config"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/harpoon/types"
	"github.com/neutron-org/neutron/v6/x/harpoon"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v6/testutil/harpoon/keeper"
	"github.com/neutron-org/neutron/v6/x/harpoon/types"
)

const (
	ContractAddress1 = "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a"
	ContractAddress2 = "neutron1u9dulasrfe6laxwwjx83njhm5as466uz43arfheq79k8eqahqhasf3tj94"
)

func TestGenesis(t *testing.T) {
	config.GetDefaultConfig()

	ctrl := gomock.NewController(t)

	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)

	// nil state genesis works
	genesisState := types.GenesisState{
		HookSubscriptions: nil,
	}

	k, ctx := keepertest.HarpoonKeeper(t, wasmKeeper)
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
	k, ctx = keepertest.HarpoonKeeper(t, wasmKeeper)
	wasmKeeper.EXPECT().HasContractInfo(ctx, types2.MustAccAddressFromBech32(ContractAddress1)).Times(2).Return(true)
	wasmKeeper.EXPECT().HasContractInfo(ctx, types2.MustAccAddressFromBech32(ContractAddress2)).Times(1).Return(true)
	harpoon.InitGenesis(ctx, k, genesisState2)
	got2 := harpoon.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, &genesisState2, got2)

	// non nil genesis -  hook subscriptions, contract does not exist
	genesisState3 := types.GenesisState{
		HookSubscriptions: []types.HookSubscriptions{
			{
				HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_BONDED,
				ContractAddresses: []string{ContractAddress1},
			},
		},
	}
	k, ctx = keepertest.HarpoonKeeper(t, wasmKeeper)
	wasmKeeper.EXPECT().HasContractInfo(ctx, types2.MustAccAddressFromBech32(ContractAddress1)).Times(1).Return(false)
	require.PanicsWithError(t, errors.Wrap(errors2.ErrInvalidAddress, fmt.Sprintf("contract address not found: %s", ContractAddress1)).Error(), func() { harpoon.InitGenesis(ctx, k, genesisState3) })
}
