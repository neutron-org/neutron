package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/x/harpoon/types"
)

const ContractAddress1 = "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a"

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
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
				HookSubscriptions: []types.HookSubscriptions{
					{
						HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_CREATED,
						ContractAddresses: []string{ContractAddress1},
					},
				},
			},
			valid: true,
		},
		{
			desc: "empty address",
			genState: &types.GenesisState{
				HookSubscriptions: []types.HookSubscriptions{
					{
						HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_CREATED,
						ContractAddresses: []string{""},
					},
				},
			},
			valid: false,
		},
		{
			desc: "unspecified hook type",
			genState: &types.GenesisState{
				HookSubscriptions: []types.HookSubscriptions{
					{
						HookType:          types.HOOK_TYPE_UNSPECIFIED,
						ContractAddresses: []string{ContractAddress1},
					},
				},
			},
			valid: false,
		},
		{
			desc: "invalid address",
			genState: &types.GenesisState{
				HookSubscriptions: []types.HookSubscriptions{
					{
						HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_CREATED,
						ContractAddresses: []string{"whatever"},
					},
				},
			},
			valid: false,
		},
		{
			desc: "invalid hook",
			genState: &types.GenesisState{
				HookSubscriptions: []types.HookSubscriptions{
					{
						HookType:          types.HookType(-200),
						ContractAddresses: []string{ContractAddress1},
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicate hook",
			genState: &types.GenesisState{
				HookSubscriptions: []types.HookSubscriptions{
					{
						HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_CREATED,
						ContractAddresses: []string{ContractAddress1},
					},
					{
						HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_CREATED,
						ContractAddresses: []string{ContractAddress1},
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicate contract address",
			genState: &types.GenesisState{
				HookSubscriptions: []types.HookSubscriptions{
					{
						HookType:          types.HOOK_TYPE_AFTER_VALIDATOR_CREATED,
						ContractAddresses: []string{ContractAddress1, ContractAddress1},
					},
				},
			},
			valid: false,
		},
	}
	for _, tc := range tests {
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
