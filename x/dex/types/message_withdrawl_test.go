package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/testutil/common/sample"
	. "github.com/neutron-org/neutron/x/dex/types"
)

func TestMsgWithdrawal_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgWithdrawal
		err  error
	}{
		{
			name: "invalid creator",
			msg: MsgWithdrawal{
				Creator:         "invalid_address",
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: ErrInvalidAddress,
		},
		{
			name: "invalid receiver",
			msg: MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        "invalid_address",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: ErrInvalidAddress,
		},
		{
			name: "invalid fee indexes length",
			msg: MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: ErrUnbalancedTxArray,
		},
		{
			name: "invalid tick indexes length",
			msg: MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: ErrUnbalancedTxArray,
		},
		{
			name: "invalid shares to remove length",
			msg: MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{},
			},
			err: ErrUnbalancedTxArray,
		},
		{
			name: "no withdraw specs",
			msg: MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				SharesToRemove:  []math.Int{},
			},
			err: ErrZeroWithdraw,
		},
		{
			name: "no withdraw specs",
			msg: MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []math.Int{math.ZeroInt()},
			},
			err: ErrZeroWithdraw,
		},
		{
			name: "invalid tick + fee upper",
			msg: MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{3},
				TickIndexesAToB: []int64{559678},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: ErrTickOutsideRange,
		},
		{
			name: "invalid tick + fee lower",
			msg: MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{50},
				TickIndexesAToB: []int64{-559631},
				SharesToRemove:  []math.Int{math.OneInt()},
			},
			err: ErrTickOutsideRange,
		},
		{
			name: "valid msg",
			msg: MsgWithdrawal{
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
