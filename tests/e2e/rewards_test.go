package e2e

import (
	"github.com/neutron-org/neutron/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

type RewardsTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestRewards(t *testing.T) {
	suite.Run(t, new(RewardsTestSuite))
}

func (suite *RewardsTestSuite) TestOnRecvPacketHooks() {
	suite.ConfigureTransferChannel()
	neutron := suite.GetNeutronZoneApp(suite.ChainA)

	// so we have a connection
	suite.
		suite.TransferPath.EndpointA.ChannelID

	// we need to make a transaction with non-untrn fees

	// after that check that we transferred fees to the specified address
}
