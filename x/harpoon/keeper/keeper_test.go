package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"

	types2 "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"

	"github.com/golang/mock/gomock"

	testutil_keeper "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/harpoon/types"
)

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

func TestCallSudoForSubscriptionType(t *testing.T) {
	ctrl := gomock.NewController(t)

	accountKeeper := mock_types.NewMockAccountKeeper(ctrl)
	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmKeeper, accountKeeper)

	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated},
	})

	msg := types.SudoAfterValidatorCreated{
		ValAddr: []byte("test"),
	}
	msgBz, err := json.Marshal(msg)
	require.NoError(t, err)

	// Returning no error and no calls when not subscribed to hook
	err = k.CallSudoForSubscriptionType(ctx, types.HookType_AfterDelegationModified, msg)
	require.NoError(t, err)

	// Returning no error and call when subscribed to hook
	wasmKeeper.EXPECT().Sudo(ctx, types2.MustAccAddressFromBech32(ContractAddress1), msgBz).Times(1).Return(nil, nil)
	err = k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorCreated, msg)
	require.NoError(t, err)

	// Returning error when the only one subscribed hook is erroring
	returnedError := fmt.Errorf("error")
	wasmKeeper.EXPECT().Sudo(ctx, types2.MustAccAddressFromBech32(ContractAddress1), msgBz).Times(1).Return(nil, returnedError)
	err = k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorCreated, msg)
	require.ErrorIs(t, err, returnedError)

	// Returning errors when one of the subscribed hooks is erroring
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated},
	})
	wasmKeeper.EXPECT().Sudo(ctx, types2.MustAccAddressFromBech32(ContractAddress1), msgBz).Times(1).Return(nil, nil)
	wasmKeeper.EXPECT().Sudo(ctx, types2.MustAccAddressFromBech32(ContractAddress2), msgBz).Times(1).Return(nil, returnedError)
	err = k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorCreated, msg)
	require.ErrorIs(t, err, returnedError)
}

func TestSetHookSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)

	accountKeeper := mock_types.NewMockAccountKeeper(ctrl)
	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmKeeper, accountKeeper)

	k.SetHookSubscription(ctx, types.HookSubscriptions{
		HookType:          types.HookType_AfterValidatorBonded,
		ContractAddresses: []string{ContractAddress1},
	})

	res := k.GetAllSubscriptions(ctx)
	require.EqualValues(t, []types.HookSubscriptions{
		{
			HookType:          types.HookType_AfterValidatorBonded,
			ContractAddresses: []string{ContractAddress1},
		},
	}, res)
}

func TestGetAllSubscriptions(t *testing.T) {
	ctrl := gomock.NewController(t)

	accountKeeper := mock_types.NewMockAccountKeeper(ctrl)
	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmKeeper, accountKeeper)

	res := k.GetAllSubscriptions(ctx)
	require.Empty(t, res)

	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{types.HookType_AfterValidatorBonded, types.HookType_BeforeDelegationRemoved},
	})

	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HookType_AfterValidatorBonded, types.HookType_AfterValidatorBeginUnbonding},
	})
	res = k.GetAllSubscriptions(ctx)

	require.EqualValues(t, []types.HookSubscriptions{
		{
			HookType:          types.HookType_AfterValidatorBonded,
			ContractAddresses: []string{ContractAddress1, ContractAddress2},
		},
		{
			HookType:          types.HookType_AfterValidatorBeginUnbonding,
			ContractAddresses: []string{ContractAddress2},
		},
		{
			HookType:          types.HookType_BeforeDelegationRemoved,
			ContractAddresses: []string{ContractAddress1},
		},
	}, res)
}
