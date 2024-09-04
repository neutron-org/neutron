package ibc_rate_limit_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v4/testutil/apptesting"
)

type GenesisTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.Setup()
}

func (suite *GenesisTestSuite) TestInitExportGenesis() {
	//testAddress := sdk.AccAddress([]byte("addr1_______________")).String()
	//suite.SetupTest()
	//k := suite.App.RateLimitingICS4Wrapper
	//
	//initialGenesis := types.GenesisState{
	//	Params: types.Params{
	//		ContractAddress: testAddress,
	//	},
	//}
	//
	//k.InitGenesis(suite.Ctx, initialGenesis)
	//
	//suite.Require().Equal(testAddress, k.GetParams(suite.Ctx).ContractAddress)
	//
	//exportedGenesis := k.ExportGenesis(suite.Ctx)
	//
	//suite.Require().Equal(initialGenesis, *exportedGenesis)
}
