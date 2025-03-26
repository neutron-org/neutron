package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil"

	"github.com/neutron-org/neutron/v6/x/harpoon/types"
)

func TestMsgManageHookSubscriptionValidate(t *testing.T) {
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
					Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_BONDED, types.HOOK_TYPE_AFTER_DELEGATION_MODIFIED, types.HOOK_TYPE_AFTER_VALIDATOR_BONDED},
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
			"unspecified hook type",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{types.HOOK_TYPE_UNSPECIFIED},
				},
			},
			"non-existing hook type: unspecified hooks are not allowed",
		},
		{
			"all good",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{types.HOOK_TYPE_BEFORE_DELEGATION_REMOVED},
				},
			},
			"",
		},
	}

	for _, tt := range tests {
		err := tt.manageHookSubscriptionMsg.Validate()

		if tt.expectedErr == "" {
			require.NoError(t, err, tt.expectedErr)
		} else {
			require.ErrorContains(t, err, tt.expectedErr)
		}
	}
}
