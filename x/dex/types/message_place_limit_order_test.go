package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v3/testutil/common/sample"
	dextypes "github.com/neutron-org/neutron/v3/x/dex/types"
)

func TestMsgPlaceLimitOrder_ValidateBasic(t *testing.T) {
	ZEROINT := math.ZeroInt()
	ONEINT := math.OneInt()
	tests := []struct {
		name string
		msg  dextypes.MsgPlaceLimitOrder
		err  error
	}{
		{
			name: "invalid creator",
			msg: dextypes.MsgPlaceLimitOrder{
				Creator:          "invalid_address",
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.OneInt(),
			},
			err: dextypes.ErrInvalidAddress,
		},
		{
			name: "invalid receiver",
			msg: dextypes.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         "invalid_address",
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.OneInt(),
			},
			err: dextypes.ErrInvalidAddress,
		},
		{
			name: "invalid zero limit order",
			msg: dextypes.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.ZeroInt(),
			},
			err: dextypes.ErrZeroLimitOrder,
		},
		{
			name: "zero maxOut",
			msg: dextypes.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.OneInt(),
				MaxAmountOut:     &ZEROINT,
				OrderType:        dextypes.LimitOrderType_FILL_OR_KILL,
			},
			err: dextypes.ErrZeroMaxAmountOut,
		},
		{
			name: "max out with maker order",
			msg: dextypes.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.OneInt(),
				MaxAmountOut:     &ONEINT,
				OrderType:        dextypes.LimitOrderType_GOOD_TIL_CANCELLED,
			},
			err: dextypes.ErrInvalidMaxAmountOutForMaker,
		},
		{
			name: "tick outside range upper",
			msg: dextypes.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 700_000,
				AmountIn:         math.OneInt(),
				OrderType:        dextypes.LimitOrderType_GOOD_TIL_CANCELLED,
			},
			err: dextypes.ErrTickOutsideRange,
		},
		{
			name: "tick outside range lower",
			msg: dextypes.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: -600_000,
				AmountIn:         math.OneInt(),
				OrderType:        dextypes.LimitOrderType_GOOD_TIL_CANCELLED,
			},
			err: dextypes.ErrTickOutsideRange,
		},
		{
			name: "valid msg",
			msg: dextypes.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.OneInt(),
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
