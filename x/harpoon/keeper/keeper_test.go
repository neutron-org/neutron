package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"

	types2 "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/x/harpoon/types"

	"github.com/golang/mock/gomock"

	testutil_keeper "github.com/neutron-org/neutron/v6/testutil/harpoon/keeper"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/harpoon/types"
)

func TestUpdateHookSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)

	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmKeeper)

	// empty update on empty subscription should work
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{},
	})

	// add hook to ContractAddress1
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED), []string{ContractAddress1})

	// add same hook to ContractAddress2
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED), []string{ContractAddress1, ContractAddress2})

	// add hooks to ContractAddress3
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress3,
		Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED, types.HOOK_TYPE_AFTER_VALIDATOR_BONDED},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED), []string{ContractAddress1, ContractAddress2, ContractAddress3})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_BONDED), []string{ContractAddress3})

	// remove hook from ContractAddress2
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED), []string{ContractAddress1, ContractAddress3})

	// add more hooks for ContractAddress2
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED, types.HOOK_TYPE_AFTER_VALIDATOR_BONDED, types.HOOK_TYPE_AFTER_UNBONDING_INITIATED, types.HOOK_TYPE_BEFORE_VALIDATOR_MODIFIED},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED), []string{ContractAddress1, ContractAddress3, ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_BONDED), []string{ContractAddress3, ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_UNBONDING_INITIATED), []string{ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_BEFORE_VALIDATOR_MODIFIED), []string{ContractAddress2})

	// update hooks for ContractAddress3 deleting some hooks, adding new hook
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		// note: deleted HOOK_TYPE_AFTER_VALIDATOR_BONDED, added HOOK_TYPE_BEFORE_DELEGATION_REMOVED
		Hooks: []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED, types.HOOK_TYPE_AFTER_UNBONDING_INITIATED, types.HOOK_TYPE_BEFORE_VALIDATOR_MODIFIED, types.HOOK_TYPE_BEFORE_DELEGATION_REMOVED},
	})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED), []string{ContractAddress1, ContractAddress3, ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_BONDED), []string{ContractAddress3})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_AFTER_UNBONDING_INITIATED), []string{ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_BEFORE_VALIDATOR_MODIFIED), []string{ContractAddress2})
	require.EqualValues(t, k.GetSubscribedAddressesForHookType(ctx, types.HOOK_TYPE_BEFORE_DELEGATION_REMOVED), []string{ContractAddress2})
}

func TestCallSudoForSubscriptionType(t *testing.T) {
	ctrl := gomock.NewController(t)

	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmKeeper)

	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED},
	})

	msg := types.AfterValidatorCreatedSudoMsg{
		AfterValidatorCreated: types.AfterValidatorCreatedMsg{
			ValAddr: "neutronvaloper18hl5c9xn5dze2g50uaw0l2mr02ew57zk5tccmr",
		},
	}
	msgBz, err := json.Marshal(msg)
	require.NoError(t, err)

	// Returning no error and no calls when not subscribed to hook
	err = k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_DELEGATION_MODIFIED, msg)
	require.NoError(t, err)

	// Returning no error and call when subscribed to hook
	wasmKeeper.EXPECT().Sudo(ctx, types2.MustAccAddressFromBech32(ContractAddress1), msgBz).Times(1).Return(nil, nil)
	err = k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED, msg)
	require.NoError(t, err)

	// Returning error when the only one subscribed hook is erroring
	returnedError := fmt.Errorf("error")
	wasmKeeper.EXPECT().Sudo(ctx, types2.MustAccAddressFromBech32(ContractAddress1), msgBz).Times(1).Return(nil, returnedError)
	err = k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED, msg)
	require.ErrorIs(t, err, returnedError)

	// Returning errors when one of the subscribed hooks is erroring
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED},
	})
	wasmKeeper.EXPECT().Sudo(ctx, types2.MustAccAddressFromBech32(ContractAddress1), msgBz).Times(1).Return(nil, nil)
	wasmKeeper.EXPECT().Sudo(ctx, types2.MustAccAddressFromBech32(ContractAddress2), msgBz).Times(1).Return(nil, returnedError)
	err = k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED, msg)
	require.ErrorIs(t, err, returnedError)
}

func TestSetHookSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)

	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmKeeper)

	k.SetHookSubscription(ctx, types.HookSubscriptions{
		HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_BONDED,
		ContractAddresses: []string{ContractAddress1},
	})

	res := k.GetAllSubscriptions(ctx)
	require.EqualValues(t, []types.HookSubscriptions{
		{
			HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_BONDED,
			ContractAddresses: []string{ContractAddress1},
		},
	}, res)
}

func TestGetAllSubscriptions(t *testing.T) {
	ctrl := gomock.NewController(t)

	wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmKeeper)

	res := k.GetAllSubscriptions(ctx)
	require.Empty(t, res)

	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_BONDED, types.HOOK_TYPE_BEFORE_DELEGATION_REMOVED},
	})

	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress2,
		Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_BONDED, types.HOOK_TYPE_AFTER_VALIDATOR_BEGIN_UNBONDING},
	})
	res = k.GetAllSubscriptions(ctx)

	require.EqualValues(t, []types.HookSubscriptions{
		{
			HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_BONDED,
			ContractAddresses: []string{ContractAddress1, ContractAddress2},
		},
		{
			HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_BEGIN_UNBONDING,
			ContractAddresses: []string{ContractAddress2},
		},
		{
			HookType:          types.HOOK_TYPE_BEFORE_DELEGATION_REMOVED,
			ContractAddresses: []string{ContractAddress1},
		},
	}, res)
}
