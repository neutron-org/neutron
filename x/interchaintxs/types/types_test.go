package types_test

import (
	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestICAOwner(t *testing.T) {
	for _, tc := range []struct {
		desc            string
		contractAddress string
		owner           string
		expectedErr     error
	}{
		{
			desc:            "valid",
			contractAddress: "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs",
			expectedErr:     nil,
		},
		{
			desc:            "invalid",
			contractAddress: "invalid_contract",
			expectedErr:     types.ErrInvalidAccountAddress,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			icaOwner, err := types.NewICAOwner(tc.contractAddress, tc.owner)
			if tc.expectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, icaOwner.GetContract().String(), tc.contractAddress)
			} else {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedErr)
			}
		})
	}
}
