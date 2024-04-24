package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v3/testutil/globalfee/keeper"
	"github.com/neutron-org/neutron/v3/x/globalfee/keeper"
	"github.com/neutron-org/neutron/v3/x/globalfee/types"
)

func TestMsgUpdateParamsValidate(t *testing.T) {
	k, ctx := testkeeper.GlobalFeeKeeper(t)
	msgServer := keeper.NewMsgServerImpl(*k)

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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.UpdateParams(ctx, &tt.msg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}
