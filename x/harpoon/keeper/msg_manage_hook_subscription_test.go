package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v5/testutil"
	testutil_keeper "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/harpoon/types"
	"github.com/neutron-org/neutron/v5/x/harpoon/keeper"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

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
