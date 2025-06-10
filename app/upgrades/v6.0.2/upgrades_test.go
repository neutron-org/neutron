package v602_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v601 "github.com/neutron-org/neutron/v6/app/upgrades/v6.0.1"
	"github.com/neutron-org/neutron/v6/testutil"
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

func (suite *UpgradeTestSuite) TestUpgradeRemoveOrphanedLimitOrders() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	upgrade := upgradetypes.Plan{
		Name:   v601.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}

	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))
}
