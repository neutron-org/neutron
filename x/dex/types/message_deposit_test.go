package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
)

func TestMsgDeposit_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  dextypes.MsgDeposit
		err  error
	}{
		{
			name: "invalid creator",
			msg: dextypes.MsgDeposit{
				Creator:         "invalid_address",
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []math.Int{math.OneInt()},
				AmountsB:        []math.Int{math.OneInt()},
				Options:         []*dextypes.DepositOptions{{false}},
			},
			err: dextypes.ErrInvalidAddress,
		},
		{
			name: "invalid receiver",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        "invalid address",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []math.Int{math.OneInt()},
				AmountsB:        []math.Int{math.OneInt()},
				Options:         []*dextypes.DepositOptions{{false}},
			},
			err: dextypes.ErrInvalidAddress,
		},
		{
			name: "invalid fee indexes length",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{},
				AmountsA:        []math.Int{},
				AmountsB:        []math.Int{},
				Options:         []*dextypes.DepositOptions{{false}},
			},
			err: dextypes.ErrUnbalancedTxArray,
		},
		{
			name: "invalid tick indexes length",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{0},
				AmountsA:        []math.Int{},
				AmountsB:        []math.Int{},
				Options:         []*dextypes.DepositOptions{{true}},
			},
			err: dextypes.ErrUnbalancedTxArray,
		},
		{
			name: "invalid amounts A length",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				AmountsA:        []math.Int{math.OneInt()},
				AmountsB:        []math.Int{},
				Options:         []*dextypes.DepositOptions{{true}},
			},
			err: dextypes.ErrUnbalancedTxArray,
		},
		{
			name: "invalid amounts B length",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				AmountsA:        []math.Int{},
				AmountsB:        []math.Int{math.OneInt()},
				Options:         []*dextypes.DepositOptions{{true}},
			},
			err: dextypes.ErrUnbalancedTxArray,
		},
		{
			name: "invalid options length",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{1},
				TickIndexesAToB: []int64{1},
				AmountsA:        []math.Int{math.OneInt()},
				AmountsB:        []math.Int{math.OneInt()},
				Options:         []*dextypes.DepositOptions{},
			},
			err: dextypes.ErrUnbalancedTxArray,
		},
		{
			name: "invalid no deposit",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				AmountsA:        []math.Int{},
				AmountsB:        []math.Int{},
				Options:         []*dextypes.DepositOptions{},
			},
			err: dextypes.ErrZeroDeposit,
		},
		{
			name: "invalid duplicate deposit",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{1, 2, 1},
				TickIndexesAToB: []int64{0, 0, 0},
				AmountsA:        []math.Int{math.OneInt(), math.OneInt(), math.OneInt()},
				AmountsB:        []math.Int{math.OneInt(), math.OneInt(), math.OneInt()},
				Options:         []*dextypes.DepositOptions{{false}, {false}, {false}},
			},
			err: dextypes.ErrDuplicatePoolDeposit,
		},
		{
			name: "invalid no deposit",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []math.Int{math.ZeroInt()},
				AmountsB:        []math.Int{math.ZeroInt()},
				Options:         []*dextypes.DepositOptions{{false}},
			},
			err: dextypes.ErrZeroDeposit,
		},
		{
			name: "invalid tick + fee upper",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{3},
				TickIndexesAToB: []int64{559678},
				AmountsA:        []math.Int{math.OneInt()},
				AmountsB:        []math.Int{math.OneInt()},
				Options:         []*dextypes.DepositOptions{{false}},
			},
			err: dextypes.ErrTickOutsideRange,
		},
		{
			name: "invalid tick + fee lower",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{50},
				TickIndexesAToB: []int64{-559631},
				AmountsA:        []math.Int{math.OneInt()},
				AmountsB:        []math.Int{math.OneInt()},
				Options:         []*dextypes.DepositOptions{{false}},
			},
			err: dextypes.ErrTickOutsideRange,
		},
		{
			name: "valid msg",
			msg: dextypes.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []math.Int{math.OneInt()},
				AmountsB:        []math.Int{math.OneInt()},
				Options:         []*dextypes.DepositOptions{{false}},
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
