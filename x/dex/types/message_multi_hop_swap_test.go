package types_test

import (
	"testing"

	math_utils "github.com/neutron-org/neutron/v2/utils/math"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v2/testutil/common/sample"
	. "github.com/neutron-org/neutron/v2/x/dex/types"
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
				Routes: []*MultiHopRoute{
					{Hops: []string{"A", "B", "C"}},
				},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: ErrInvalidAddress,
		},
		{
			name: "invalid receiver address",
			msg: MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: "invalid_address",
				Routes: []*MultiHopRoute{
					{Hops: []string{"A", "B", "C"}},
				},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: ErrInvalidAddress,
		},
		{
			name: "missing route",
			msg: MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*MultiHopRoute{},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
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
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: ErrMultihopExitTokensMismatch,
		},
		{
			name: "invalid amountIn",
			msg: MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
				AmountIn:       math.NewInt(-1),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: ErrZeroSwap,
		},
		{
			name: "cycles in hops",
			msg: MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Routes: []*MultiHopRoute{
					{Hops: []string{"A", "B", "C"}},                // normal
					{Hops: []string{"A", "B", "D", "E", "B", "C"}}, // has cycle
				},
				AmountIn:       math.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: ErrCycleInHops,
		},
		{
			name: "zero exit limit price",
			msg: MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
				AmountIn:       math.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0"),
			},
			err: ErrZeroExitPrice,
		},
		{
			name: "negative exit limit price",
			msg: MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
				AmountIn:       math.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("-0.5"),
			},
			err: ErrZeroExitPrice,
		},
		{
			name: "valid",
			msg: MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
				AmountIn:       math.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
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
