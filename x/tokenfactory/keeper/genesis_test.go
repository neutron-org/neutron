package keeper_test

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/bitcoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
				},
				HookContractAddress: "",
			},
			{
				Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/diff-admin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
				},
			},
			{
				Denom: "factory/neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2/litecoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
				},
				HookContractAddress: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
			},
		},
	}
	app := suite.GetNeutronZoneApp(suite.ChainA)
	context := suite.ChainA.GetContext()
	// Test both with bank denom metadata set, and not set.
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, sets bank metadata to exist if i != 0, to cover both cases.
		if i != 0 {
			app.BankKeeper.SetDenomMetaData(context, banktypes.Metadata{Base: denom.GetDenom()})
		}
	}

	err := app.TokenFactoryKeeper.SetParams(context, types.Params{})
	suite.Require().NoError(err)
	app.TokenFactoryKeeper.InitGenesis(context, genesisState)
	exportedGenesis := app.TokenFactoryKeeper.ExportGenesis(context)
	suite.Require().NotNil(exportedGenesis)
	suite.Require().Equal(genesisState, *exportedGenesis)
}
