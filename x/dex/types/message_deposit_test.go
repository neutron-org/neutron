package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/neutron-org/neutron/testutil/common/sample"
	. "github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/require"
)

func TestMsgDeposit_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeposit
		err  error
	}{
		{
			name: "invalid creator",
			msg: MsgDeposit{
				Creator:         "invalid_address",
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdk.Int{sdk.OneInt()},
				AmountsB:        []sdk.Int{sdk.OneInt()},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid receiver",
			msg: MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        "invalid address",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdk.Int{sdk.OneInt()},
				AmountsB:        []sdk.Int{sdk.OneInt()},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid fee indexes length",
			msg: MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{},
				AmountsA:        []sdk.Int{},
				AmountsB:        []sdk.Int{},
			},
			err: ErrUnbalancedTxArray,
		},
		{
			name: "invalid tick indexes length",
			msg: MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdk.Int{},
				AmountsB:        []sdk.Int{},
			},
			err: ErrUnbalancedTxArray,
		},
		{
			name: "invalid amounts A length",
			msg: MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				AmountsA:        []sdk.Int{sdk.OneInt()},
				AmountsB:        []sdk.Int{},
			},
			err: ErrUnbalancedTxArray,
		},
		{
			name: "invalid amounts B length",
			msg: MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				AmountsA:        []sdk.Int{},
				AmountsB:        []sdk.Int{sdk.OneInt()},
			},
			err: ErrUnbalancedTxArray,
		},
		{
			name: "invalid no deposit",
			msg: MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				AmountsA:        []sdk.Int{},
				AmountsB:        []sdk.Int{},
			},
			err: ErrZeroDeposit,
		},
		{
			name: "invalid no deposit",
			msg: MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdk.Int{sdk.ZeroInt()},
				AmountsB:        []sdk.Int{sdk.ZeroInt()},
			},
			err: ErrZeroDeposit,
		},
		{
			name: "valid msg",
			msg: MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdk.Int{sdk.OneInt()},
				AmountsB:        []sdk.Int{sdk.OneInt()},
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
