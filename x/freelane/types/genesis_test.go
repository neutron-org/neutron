package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v5/x/freelane/types"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.Params{
					BlockSpace:     0.1,
					SequenceNumber: 2,
				},
			},
			valid: true,
		},
		{
			desc: "invalid genesis state - params are invalid",
			genState: &types.GenesisState{
				Params: types.Params{
					BlockSpace:     -0.1,
					SequenceNumber: 2,
				},
			},
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
