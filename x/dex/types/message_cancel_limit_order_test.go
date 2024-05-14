package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
)

func TestMsgCancelLimitOrder_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  dextypes.MsgCancelLimitOrder
		err  error
	}{
		{
			name: "invalid creator",
			msg: dextypes.MsgCancelLimitOrder{
				Creator:    "invalid_address",
				TrancheKey: "ORDER123",
			},
			err: dextypes.ErrInvalidAddress,
		}, {
			name: "valid msg",
			msg: dextypes.MsgCancelLimitOrder{
				Creator:    sample.AccAddress(),
				TrancheKey: "ORDER123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
