package v800_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	v800 "github.com/neutron-org/neutron/v7/app/upgrades/v8.0.0"
	"github.com/neutron-org/neutron/v7/testutil"
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

func (suite *UpgradeTestSuite) TestUpgradeDenomsMetadata() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext()

	// Setup before state
	params := app.FeeKeeper.GetParams(ctx)
	params.FeeEnabled = false
	suite.NoError(app.FeeKeeper.SetParams(ctx, params))

	upgrade := upgradetypes.Plan{
		Name:   v800.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}

	// Before upgrade
	params = app.FeeKeeper.GetParams(ctx)
	suite.Require().False(params.FeeEnabled)

	// Apply upgrade
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	// After upgrade
	params = app.FeeKeeper.GetParams(ctx)
	suite.Require().True(params.FeeEnabled)
}
