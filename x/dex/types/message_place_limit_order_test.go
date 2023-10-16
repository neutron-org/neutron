package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/testutil/common/sample"
	. "github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/require"
)

func TestMsgPlaceLimitOrder_ValidateBasic(t *testing.T) {
	ZEROINT := math.ZeroInt()
	ONEINT := math.OneInt()
	tests := []struct {
		name string
		msg  MsgPlaceLimitOrder
		err  error
	}{
		{
			name: "invalid creator",
			msg: MsgPlaceLimitOrder{
				Creator:          "invalid_address",
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.OneInt(),
			},
			err: ErrInvalidAddress,
		},
		{
			name: "invalid receiver",
			msg: MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         "invalid_address",
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.OneInt(),
			},
			err: ErrInvalidAddress,
		},
		{
			name: "invalid zero limit order",
			msg: MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.ZeroInt(),
			},
			err: ErrZeroLimitOrder,
		},
		{
			name: "zero maxOut",
			msg: MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.OneInt(),
				MaxAmountOut:     &ZEROINT,
				OrderType:        LimitOrderType_FILL_OR_KILL,
			},
			err: ErrZeroMaxAmountOut,
		},
		{
			name: "max out with maker order",
			msg: MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         math.OneInt(),
				MaxAmountOut:     &ONEINT,
				OrderType:        LimitOrderType_GOOD_TIL_CANCELLED,
			},
			err: ErrInvalidMaxAmountOutForMaker,
		},
		{
			name: "valid msg",
			msg: MsgPlaceLimitOrder{
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
