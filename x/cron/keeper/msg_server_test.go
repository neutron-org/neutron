package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/testutil"
	"github.com/neutron-org/neutron/v4/testutil/cron/keeper"
	"github.com/neutron-org/neutron/v4/x/cron/types"
)

func TestMsgUpdateParamsValidate(t *testing.T) {
	k, ctx := keeper.CronKeeper(t, nil, nil)

	tests := []struct {
		name        string
		msg         types.MsgUpdateParams
		expectedErr string
	}{
		{
			"empty authority",
			types.MsgUpdateParams{
				Authority: "",
			},
			"authority is invalid",
		},
		{
			"invalid authority",
			types.MsgUpdateParams{
				Authority: "invalid authority",
			},
			"authority is invalid",
		},
		{
			"empty security_address",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					SecurityAddress: "",
				},
			},
			"security_address is invalid",
		},
		{
			"invalid security_address",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					SecurityAddress: "invalid security_address",
				},
			},
			"security_address is invalid",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			resp, err := k.UpdateParams(ctx, &tt.msg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}
