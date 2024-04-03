package v300_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v3/testutil"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	//suite.IBCConnectionTestSuite.SetupTest()
	//suite.Require().NoError(
	//suite.GetNeutronZoneApp(suite.ChainA).FeeBurnerKeeper.SetParams(
	//	suite.ChainA.GetContext(), feeburnertypes.NewParams(feeburnertypes.DefaultNeutronDenom),
	//))
}

func (suite *UpgradeTestSuite) TestOracleUpgrade() {
	// TODO
}
