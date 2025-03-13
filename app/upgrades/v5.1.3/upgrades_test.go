package v513_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v513 "github.com/neutron-org/neutron/v5/app/upgrades/v5.1.3"
	"github.com/neutron-org/neutron/v5/testutil"
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
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	upgrade := upgradetypes.Plan{
		Name:   v513.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))
}
