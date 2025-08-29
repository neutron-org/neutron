package keeper_test

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/neutron-org/neutron/v8/x/tokenfactory/types"
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
			app.BankKeeper.SetDenomMetaData(context, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}

	err := app.TokenFactoryKeeper.SetParams(context, types.Params{})
	suite.Require().NoError(err)
	app.TokenFactoryKeeper.InitGenesis(context, genesisState)

	exportedGenesis := app.TokenFactoryKeeper.ExportGenesis(context)
	suite.Require().NotNil(exportedGenesis)
	suite.Require().Equal(genesisState, *exportedGenesis)

	// verify that the exported bank genesis is valid
	err = app.BankKeeper.SetParams(context, banktypes.DefaultParams())
	suite.Require().NoError(err)
	exportedBankGenesis := app.BankKeeper.ExportGenesis(context)
	suite.Require().NoError(exportedBankGenesis.Validate())

	app.BankKeeper.InitGenesis(context, exportedBankGenesis)
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, check whether bank metadata is not replaced if i != 0, to cover both cases.
		if i != 0 {
			metadata, found := app.BankKeeper.GetDenomMetaData(context, denom.GetDenom())
			suite.Require().True(found)
			suite.Require().EqualValues(metadata, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}
}
