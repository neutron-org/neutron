package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil"
	"github.com/neutron-org/neutron/v6/testutil/feeburner/keeper"
	"github.com/neutron-org/neutron/v6/x/feeburner/types"
)

func TestMsgUpdateParamsValidate(t *testing.T) {
	k, ctx := keeper.FeeburnerKeeper(t)

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
			"empty treasury_address",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					TreasuryAddress: "",
				},
			},
			"treasury_address is invalid",
		},
		{
			"invalid treasury_address",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					TreasuryAddress: "invalid treasury_address",
				},
			},
			"treasury_address is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := k.UpdateParams(ctx, &tt.msg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}
