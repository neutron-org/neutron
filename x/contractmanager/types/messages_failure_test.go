package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/neutron-org/neutron/testutil/contractmanager/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateFailure_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateFailure
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateFailure{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateFailure{
				Creator: sample.AccAddress(),
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

func TestMsgUpdateFailure_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateFailure
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUpdateFailure{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgUpdateFailure{
				Creator: sample.AccAddress(),
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

func TestMsgDeleteFailure_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeleteFailure
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDeleteFailure{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgDeleteFailure{
				Creator: sample.AccAddress(),
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
