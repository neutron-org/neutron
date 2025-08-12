package v800rc0_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/neutron-org/neutron/v8/app/upgrades/v8.0.0-rc0"
	"github.com/neutron-org/neutron/v8/testutil"
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

func (suite *UpgradeTestSuite) TestUpgrade() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext()

	upgrade := upgradetypes.Plan{
		Name:   v800rc0.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}

	// Apply upgrade
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))
}
