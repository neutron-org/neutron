package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v2/testutil/common/sample"
	. "github.com/neutron-org/neutron/v2/x/dex/types"
)

func TestMsgCancelLimitOrder_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCancelLimitOrder
		err  error
	}{
		{
			name: "invalid creator",
			msg: MsgCancelLimitOrder{
				Creator:    "invalid_address",
				TrancheKey: "ORDER123",
			},
			err: ErrInvalidAddress,
		}, {
			name: "valid msg",
			msg: MsgCancelLimitOrder{
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
