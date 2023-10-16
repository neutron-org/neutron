package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/testutil/common/sample"
	. "github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/require"
)

func TestMsgMultiHopSwap_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMultiHopSwap
		err  error
	}{
		{
			name: "invalid creator address",
			msg: MsgMultiHopSwap{
				Creator:  "invalid_address",
				Receiver: sample.AccAddress(),
			},
			err: ErrInvalidAddress,
		},
		{
			name: "invalid receiver address",
			msg: MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: "invalid_address",
			},
			err: ErrInvalidAddress,
		},
		{
			name: "missing route",
			msg: MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
			},
			err: ErrMissingMultihopRoute,
		},
		{
			name: "invalid exit tokens",
			msg: MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Routes: []*MultiHopRoute{
					{Hops: []string{"A", "B", "C"}},
					{Hops: []string{"A", "B", "Z"}},
				},
			},
			err: ErrMultihopExitTokensMismatch,
		},
		{
			name: "invalid amountIn",
			msg: MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Routes:   []*MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
				AmountIn: math.NewInt(-1),
			},
			err: ErrZeroSwap,
		},
		{
			name: "valid",
			msg: MsgMultiHopSwap{
				Routes:   []*MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				AmountIn: math.OneInt(),
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
