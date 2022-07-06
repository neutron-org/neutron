package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestICAOwner(t *testing.T) {
	validAddr, _ := sdk.AccAddressFromBech32("cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs")
	for _, tc := range []struct {
		desc        string
		icaOwner    types.ICAOwner
		contract    sdk.AccAddress
		expectedErr error
	}{
		{
			desc:        "valid",
			icaOwner:    types.NewICAOwner("cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs", "owner"),
			contract:    validAddr,
			expectedErr: nil,
		},
		{
			desc:        "invalid",
			icaOwner:    types.ICAOwner("invalid_owner_format"),
			contract:    nil,
			expectedErr: types.ErrInvalidICAOwner,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			addr, err := tc.icaOwner.GetContract()
			if tc.expectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, validAddr, addr)
			} else {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedErr)
			}
		})
	}
}
