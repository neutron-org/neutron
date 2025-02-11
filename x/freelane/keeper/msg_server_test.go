package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v5/testutil"
	"github.com/neutron-org/neutron/v5/testutil/freelane/keeper"
	"github.com/neutron-org/neutron/v5/x/freelane/types"
)

func TestMsgUpdateParamsValidate(t *testing.T) {
	k, ctx := keeper.FreeLaneKeeper(t)

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
			"invalid block space",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					BlockSpace: -0.1,
				},
			},
			"block_space is invalid",
		},
		{
			"invalid block space",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					BlockSpace: 1.1,
				},
			},
			"block_space is invalid",
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
