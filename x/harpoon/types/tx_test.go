package types_test

import (
	"github.com/neutron-org/neutron/v5/testutil"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"
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
			"all good",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{types.HookType_BeforeDelegationRemoved},
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
