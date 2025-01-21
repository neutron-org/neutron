package keeper_test

import (
	"context"
	"github.com/neutron-org/neutron/v5/testutil"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	"github.com/neutron-org/neutron/v5/x/harpoon/keeper"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"

	"github.com/golang/mock/gomock"

	testutil_keeper "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/harpoon/types"
)

const (
	ContractAddress1 = "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a"
	ContractAddress2 = "neutron1u9dulasrfe6laxwwjx83njhm5as466uz43arfheq79k8eqahqhasf3tj94"
	ContractAddress3 = "neutron1k5996rnrjcu4dwxtjq46zuh4dryl6pdrlyja7fw2qm7yarkvk97s9uzmuz"
)

func setupMsgServer(t testing.TB) (keeper.Keeper, types.MsgServer, context.Context) {
	k, ctx := keepertest.HarpoonKeeper(t, nil, nil)
	return *k, keeper.NewMsgServerImpl(*k), ctx
}

func TestMsgServer(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
	require.NotEmpty(t, k)
}

func TestManageHookSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountKeeper := mock_types.NewMockAccountKeeper(ctrl)

	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmKeeper, accountKeeper)

	msgServer := keeper.NewMsgServerImpl(*k)

	tests := []struct {
		name                      string
		manageHookSubscriptionMsg types.MsgManageHookSubscription
		expectedErr               string
	}{
		{
			"empty authority",
			types.MsgManageHookSubscription{
				Authority: "",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{},
				},
			},
			"authority is invalid: empty address string is not allowed",
		},
		{
			"non unique hooks",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{types.HookType_AfterValidatorBonded, types.HookType_AfterDelegationModified, types.HookType_AfterValidatorBonded},
				},
			},
			"subscription hooks are not unique",
		},
		{
			"non existing hook type",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{types.HookType(100)},
				},
			},
			"non-existing hook type",
		},
		{
			"good case - empty hooks",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{},
				},
			},
			"",
		},
		{
			"good case - some hooks present",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{types.HookType_AfterValidatorCreated, types.HookType_AfterDelegationModified},
				},
			},
			"",
		},
	}

	for _, tt := range tests {
		res, err := msgServer.ManageHookSubscription(ctx, &tt.manageHookSubscriptionMsg)

		if tt.expectedErr == "" {
			require.NoError(t, err, tt.expectedErr)
			require.Equal(t, res, &types.MsgManageHookSubscriptionResponse{})
		} else {
			require.ErrorContains(t, err, tt.expectedErr)
			require.Empty(t, res)
		}
	}
}

func TestUpdateHookSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)

	accountKeeper := mock_types.NewMockAccountKeeper(ctrl)
	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmKeeper, accountKeeper)

	// empty update on empty subscription should work
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{},
	})

	// add hook to ContractAddress1
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1})

	// add same hook to ContractAddress2
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress2})

	// add hooks to ContractAddress3
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress3,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated, types.HookType_AfterValidatorBonded},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress2, ContractAddress3})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorBonded), []string{ContractAddress3})

	// remove hook from ContractAddress2
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress3})

	// add more hooks for ContractAddress2
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated, types.HookType_AfterValidatorBonded, types.HookType_AfterUnbondingInitiated, types.HookType_BeforeValidatorModified},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress3, ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorBonded), []string{ContractAddress3, ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterUnbondingInitiated), []string{ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeValidatorModified), []string{ContractAddress2})

	// update hooks for ContractAddress3 deleting some hooks, adding new hook
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		// note: deleted HookType_AfterValidatorBonded, added HookType_BeforeDelegationRemoved
		Hooks: []types.HookType{types.HookType_AfterValidatorCreated, types.HookType_AfterUnbondingInitiated, types.HookType_BeforeValidatorModified, types.HookType_BeforeDelegationRemoved},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated), []string{ContractAddress1, ContractAddress3, ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorBonded), []string{ContractAddress3})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterUnbondingInitiated), []string{ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeValidatorModified), []string{ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeDelegationRemoved), []string{ContractAddress2})
}
