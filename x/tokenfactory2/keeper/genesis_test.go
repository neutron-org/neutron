package keeper_test

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/neutron-org/neutron/v8/x/tokenfactory2/types"
)

func getDenom(owner, denom string) string {
	if fullDenom, err := types.GetTokenDenom(owner, denom); err != nil {
		panic(err.Error())
	} else {
		return fullDenom
	}
}

func (suite *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: getDenom("neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2", "bitcoin"),
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
				},
				HookContractAddress: "",
			},
			{
				Denom: getDenom("neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2", "diff-admin"),
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2",
				},
			},
			{
				Denom: getDenom("neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2", "litecoin"),
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

	err := app.TokenFactory2Keeper.SetParams(context, types.Params{})
	suite.Require().NoError(err)
	app.TokenFactory2Keeper.InitGenesis(context, genesisState)

	exportedGenesis := app.TokenFactory2Keeper.ExportGenesis(context)
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
