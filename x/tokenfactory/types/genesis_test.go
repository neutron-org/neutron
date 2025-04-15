package types_test

import (
	"testing"

	"github.com/neutron-org/neutron/v6/app/config"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

func TestGenesisState_Validate(t *testing.T) {
	config.GetDefaultConfig()

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
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "different admin from creator",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "empty admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "no admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
					},
				},
			},
			valid: true,
		},
		{
			desc: "invalid admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "moose",
						},
					},
				},
			},
			valid: false,
		},
		{
			desc: "multiple denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/litecoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicate denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: false,
		},
		{
			desc: "empty hook address",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
						},
						HookContractAddress: "",
					},
				},
			},
			valid: true,
		},
		{
			desc: "invalid hook address",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
						},
						HookContractAddress: "sfsdfsdfsdfs",
					},
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
