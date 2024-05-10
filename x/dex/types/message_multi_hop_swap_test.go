package types_test

import (
	"testing"

	math_utils "github.com/neutron-org/neutron/v4/utils/math"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
)

func TestMsgMultiHopSwap_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  dextypes.MsgMultiHopSwap
		err  error
	}{
		{
			name: "invalid creator address",
			msg: dextypes.MsgMultiHopSwap{
				Creator:  "invalid_address",
				Receiver: sample.AccAddress(),
				Routes: []*dextypes.MultiHopRoute{
					{Hops: []string{"A", "B", "C"}},
				},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: dextypes.ErrInvalidAddress,
		},
		{
			name: "invalid receiver address",
			msg: dextypes.MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: "invalid_address",
				Routes: []*dextypes.MultiHopRoute{
					{Hops: []string{"A", "B", "C"}},
				},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: dextypes.ErrInvalidAddress,
		},
		{
			name: "missing route",
			msg: dextypes.MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*dextypes.MultiHopRoute{},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: dextypes.ErrMissingMultihopRoute,
		},
		{
			name: "invalid exit tokens",
			msg: dextypes.MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Routes: []*dextypes.MultiHopRoute{
					{Hops: []string{"A", "B", "C"}},
					{Hops: []string{"A", "B", "Z"}},
				},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: dextypes.ErrMultihopExitTokensMismatch,
		},
		{
			name: "invalid amountIn",
			msg: dextypes.MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*dextypes.MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
				AmountIn:       math.NewInt(-1),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: dextypes.ErrZeroSwap,
		},
		{
			name: "cycles in hops",
			msg: dextypes.MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Routes: []*dextypes.MultiHopRoute{
					{Hops: []string{"A", "B", "C"}},                // normal
					{Hops: []string{"A", "B", "D", "E", "B", "C"}}, // has cycle
				},
				AmountIn:       math.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			err: dextypes.ErrCycleInHops,
		},
		{
			name: "zero exit limit price",
			msg: dextypes.MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*dextypes.MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
				AmountIn:       math.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0"),
			},
			err: dextypes.ErrZeroExitPrice,
		},
		{
			name: "negative exit limit price",
			msg: dextypes.MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*dextypes.MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
				AmountIn:       math.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("-0.5"),
			},
			err: dextypes.ErrZeroExitPrice,
		},
		{
			name: "valid",
			msg: dextypes.MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*dextypes.MultiHopRoute{{Hops: []string{"A", "B", "C"}}},
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
