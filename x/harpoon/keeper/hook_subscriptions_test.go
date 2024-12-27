package keeper_test

import (
	"github.com/golang/mock/gomock"
	testutil_keeper "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/harpoon/types"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	ContractAddress1 = "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a"
	ContractAddress2 = "neutron1u9dulasrfe6laxwwjx83njhm5as466uz43arfheq79k8eqahqhasf3tj94"
	ContractAddress3 = "neutron1k5996rnrjcu4dwxtjq46zuh4dryl6pdrlyja7fw2qm7yarkvk97s9uzmuz"
)

func TestUpdateHookSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountKeeper := mock_types.NewMockAccountKeeper(ctrl)
	wasmMsgServer := mock_types.NewMockWasmMsgServer(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmMsgServer, accountKeeper)

	// empty update on empty subscription should work
	err := k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{},
	})
	require.NoError(t, err)

	// add hook to ContractAddress1
	err = k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated},
	})
	require.NoError(t, err)
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1})

	// add same hook to ContractAddress2
	err = k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated},
	})
	require.NoError(t, err)
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress2})

	// add hooks to ContractAddress3
	err = k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress3,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated, types.HookType_AfterValidatorBonded},
	})
	require.NoError(t, err)
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress2, ContractAddress3})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorBonded), []string{ContractAddress3})

	// remove hook from ContractAddress2
	err = k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{},
	})
	require.NoError(t, err)
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress3})

	// add more hooks for ContractAddress2
	err = k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated, types.HookType_AfterValidatorBonded, types.HookType_AfterUnbondingInitiated, types.HookType_BeforeValidatorModified},
	})
	require.NoError(t, err)
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress3, ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorBonded), []string{ContractAddress3, ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterUnbondingInitiated), []string{ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeValidatorModified), []string{ContractAddress2})

	// update hooks for ContractAddress3 deleting some hooks, adding new hook
	err = k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		// note: deleted HookType_AfterValidatorBonded, added HookType_BeforeDelegationRemoved
		Hooks: []types.HookType{types.HookType_AfterValidatorCreated, types.HookType_AfterUnbondingInitiated, types.HookType_BeforeValidatorModified, types.HookType_BeforeDelegationRemoved},
	})
	require.NoError(t, err)
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress3, ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorBonded), []string{ContractAddress3})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterUnbondingInitiated), []string{ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeValidatorModified), []string{ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeDelegationRemoved), []string{ContractAddress2})
}
