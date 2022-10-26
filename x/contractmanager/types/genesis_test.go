package types_test

import (
	"testing"

	"github.com/neutron-org/neutron/x/contractmanager/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{

				FailureList: []types.Failure{
					{
						Address: "address1",
						Offset:  1,
					},
					{
						Address: "address1",
						Offset:  2,
					},
					{
						Address: "address2",
						Offset:  1,
					},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated failure",
			genState: &types.GenesisState{
				FailureList: []types.Failure{
					{
						Address: "address1",
						Offset:  1,
					},
					{
						Address: "address1",
						Offset:  1,
					},
					{
						Address: "address2",
						Offset:  1,
					},
				},
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
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
