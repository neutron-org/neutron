package types_test

import (
	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestICAOwner(t *testing.T) {
	var (
		contractAddress     string
		interchainAccountID string
	)

	for _, tc := range []struct {
		desc                         string
		malleate                     func() (types.ICAOwner, error)
		expectedStringRepresentation string
		expectedErr                  error
	}{
		{
			desc:                         "valid",
			expectedErr:                  nil,
			expectedStringRepresentation: "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs" + types.Delimiter + "id_1",
			malleate: func() (types.ICAOwner, error) {
				contractAddress = "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"
				interchainAccountID = "id_1"
				return types.NewICAOwner(contractAddress, interchainAccountID)
			},
		},
		{
			desc:        "Delimiter in the middle of the interchain account id",
			expectedErr: nil,
			expectedStringRepresentation: "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs" + types.Delimiter +
				("id_1" + types.Delimiter + "another_data"),
			malleate: func() (types.ICAOwner, error) {
				contractAddress = "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"
				interchainAccountID = "id_1" + types.Delimiter + "another_data"

				portID := contractAddress + types.Delimiter + interchainAccountID

				return types.ICAOwnerFromPort(portID)
			},
		},
		{
			desc:        "invalid",
			expectedErr: types.ErrInvalidAccountAddress,
			malleate: func() (types.ICAOwner, error) {
				contractAddress = "invalid_account_address"
				interchainAccountID = "id_1"
				return types.NewICAOwner(contractAddress, interchainAccountID)
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			icaOwner, err := tc.malleate()
			if tc.expectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, contractAddress, icaOwner.GetContract().String())
				require.Equal(t, interchainAccountID, icaOwner.GetInterchainAccountID())
				require.Equal(t, tc.expectedStringRepresentation, icaOwner.String())
			} else {
				require.Error(t, err)
				require.ErrorIs(t, tc.expectedErr, err)
			}
		})
	}
}
