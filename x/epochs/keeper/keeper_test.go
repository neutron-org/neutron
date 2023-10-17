package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/testutil/apptesting"
	"github.com/neutron-org/neutron/x/epochs/types"
)

type EpochsTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (suite *EpochsTestSuite) SetupTest() {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func TestEpochsTestSuite(t *testing.T) {
	suite.Run(t, new(EpochsTestSuite))
}
