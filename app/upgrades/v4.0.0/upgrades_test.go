package v400_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/suite"

	v400 "github.com/neutron-org/neutron/v4/app/upgrades/v4.0.0"
	"github.com/neutron-org/neutron/v4/testutil"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
}

func (suite *UpgradeTestSuite) TestOracleUpgrade() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext()
	t := suite.T()

	upgrade := upgradetypes.Plan{
		Name:   v400.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	params, err := app.MarketMapKeeper.GetParams(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(params.MarketAuthorities[0], "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z")
	suite.Require().Equal(params.Admin, "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z")
}
