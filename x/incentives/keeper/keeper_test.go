package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/testutil/apptesting"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/keeper"
	"github.com/neutron-org/neutron/x/incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	QueryServer keeper.QueryServer
	MsgServer   types.MsgServer
	LPDenom0    string
	LPDenom1    string
}

// SetupTest sets incentives parameters from the suite's context
func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.QueryServer = keeper.NewQueryServer(suite.App.IncentivesKeeper)
	suite.MsgServer = keeper.NewMsgServerImpl(suite.App.IncentivesKeeper)
	pool0, _ := suite.App.DexKeeper.GetOrInitPool(suite.Ctx,
		&dextypes.PairID{
			Token0: "TokenA",
			Token1: "TokenB",
		},
		0,
		1,
	)
	suite.LPDenom0 = pool0.GetPoolDenom()
	pool1, _ := suite.App.DexKeeper.GetOrInitPool(suite.Ctx,
		&dextypes.PairID{
			Token0: "TokenA",
			Token1: "TokenB",
		},
		1,
		1,
	)
	suite.LPDenom1 = pool1.GetPoolDenom()

	suite.SetEpochStartTime()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
