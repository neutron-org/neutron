package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v10/testutil/common/sample"
	dextypes "github.com/neutron-org/neutron/v10/x/dex/types"
)

func TestMsgWithdrawalWithShares_Validate(t *testing.T) {
	tests := []struct {
		name        string
		msg         dextypes.MsgWithdrawalWithShares
		expectedErr error
	}{
		{
			"valid message",
			dextypes.MsgWithdrawalWithShares{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				SharesToRemove: sdk.Coins{sdk.NewCoin("neutron/pool/0", sdkmath.OneInt())},
			},
			nil,
		},
		{
			"invalid creator address",
			dextypes.MsgWithdrawalWithShares{
				Creator:        "invalid_address",
				Receiver:       sample.AccAddress(),
				SharesToRemove: sdk.Coins{sdk.NewCoin("neutron/pool/0", sdkmath.OneInt())},
			},
			dextypes.ErrInvalidAddress,
		},
		{
			"invalid receiver address",
			dextypes.MsgWithdrawalWithShares{
				Creator:        sample.AccAddress(),
				Receiver:       "invalid_address",
				SharesToRemove: sdk.Coins{sdk.NewCoin("neutron/pool/0", sdkmath.OneInt())},
			},
			dextypes.ErrInvalidAddress,
		},
		{
			"empty shares to remove",
			dextypes.MsgWithdrawalWithShares{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				SharesToRemove: sdk.Coins{},
			},
			dextypes.ErrZeroWithdraw,
		},
		{
			"zero share amount",
			dextypes.MsgWithdrawalWithShares{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				SharesToRemove: sdk.Coins{sdk.NewCoin("neutron/pool/0", sdkmath.ZeroInt())},
			},
			dextypes.ErrZeroWithdraw,
		},
		{
			"invalid share denom",
			dextypes.MsgWithdrawalWithShares{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				SharesToRemove: sdk.Coins{sdk.NewCoin("invalid_denom", sdkmath.OneInt())},
			},
			dextypes.ErrInvalidPoolDenom,
		},
		{
			"duplicate share denom",
			dextypes.MsgWithdrawalWithShares{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				SharesToRemove: sdk.Coins{
					sdk.NewCoin("neutron/pool/0", sdkmath.OneInt()),
					sdk.NewCoin("neutron/pool/0", sdkmath.OneInt()),
				},
			},
			dextypes.ErrDuplicatePoolWithdraw,
		},
		{
			"valid message with multiple shares",
			dextypes.MsgWithdrawalWithShares{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				SharesToRemove: sdk.Coins{
					sdk.NewCoin("neutron/pool/0", sdkmath.OneInt()),
					sdk.NewCoin("neutron/pool/1", sdkmath.OneInt()),
					sdk.NewCoin("neutron/pool/2", sdkmath.OneInt()),
				},
			},
			nil,
		},
		{
			"invalid creator and receiver addresses",
			dextypes.MsgWithdrawalWithShares{
				Creator:        "invalid_creator",
				Receiver:       "invalid_receiver",
				SharesToRemove: sdk.Coins{sdk.NewCoin("neutron/pool/0", sdkmath.OneInt())},
			},
			dextypes.ErrInvalidAddress,
		},
		{
			"zero amount with valid denom",
			dextypes.MsgWithdrawalWithShares{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				SharesToRemove: sdk.Coins{sdk.NewCoin("neutron/pool/12345", sdkmath.ZeroInt())},
			},
			dextypes.ErrZeroWithdraw,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
