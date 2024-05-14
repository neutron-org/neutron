package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
)

func TestMsgWithdrawal_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  dextypes.MsgWithdrawal
		err  error
	}{
		{
			name: "invalid creator",
			msg: dextypes.MsgWithdrawal{
				Creator:         "invalid_address",
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: dextypes.ErrInvalidAddress,
		},
		{
			name: "invalid receiver",
			msg: dextypes.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        "invalid_address",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: dextypes.ErrInvalidAddress,
		},
		{
			name: "invalid fee indexes length",
			msg: dextypes.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: dextypes.ErrUnbalancedTxArray,
		},
		{
			name: "invalid tick indexes length",
			msg: dextypes.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: dextypes.ErrUnbalancedTxArray,
		},
		{
			name: "invalid shares to remove length",
			msg: dextypes.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{},
			},
			err: dextypes.ErrUnbalancedTxArray,
		},
		{
			name: "no withdraw specs",
			msg: dextypes.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				SharesToRemove:  []math.Int{},
			},
			err: dextypes.ErrZeroWithdraw,
		},
		{
			name: "no withdraw specs",
			msg: dextypes.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{math.ZeroInt()},
			},
			err: dextypes.ErrZeroWithdraw,
		},
		{
			name: "invalid tick + fee upper",
			msg: dextypes.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{3},
				TickIndexesAToB: []int64{559678},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: dextypes.ErrTickOutsideRange,
		},
		{
			name: "invalid tick + fee lower",
			msg: dextypes.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{50},
				TickIndexesAToB: []int64{-559631},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: dextypes.ErrTickOutsideRange,
		},
		{
			name: "valid msg",
			msg: dextypes.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{math.OneInt()},
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
