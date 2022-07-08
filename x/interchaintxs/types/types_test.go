package types_test

import (
	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestICAOwner(t *testing.T) {
	for _, tc := range []struct {
		desc                         string
		contractAddress              string
		interchainAccountID          string
		expectedErr                  error
		expectedStringRepresentation string
	}{
		{
			desc:                         "valid",
			contractAddress:              "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs",
			interchainAccountID:          "id_1",
			expectedErr:                  nil,
			expectedStringRepresentation: "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs.id_1",
		},
		{
			desc:                "invalid",
			contractAddress:     "invalid_contract",
			interchainAccountID: "id_1",
			expectedErr:         types.ErrInvalidAccountAddress,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			icaOwner, err := types.NewICAOwner(tc.contractAddress, tc.interchainAccountID)
			if tc.expectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, tc.contractAddress, icaOwner.GetContract().String())
				require.Equal(t, tc.interchainAccountID, icaOwner.GetInterchainAccountID())
				require.Equal(t, tc.expectedStringRepresentation, icaOwner.String())
			} else {
				require.Error(t, err)
				require.ErrorIs(t, tc.expectedErr, err)
			}
		})
	}
}
