package keeper_test

import (
	gocontext "context"

	"github.com/neutron-org/neutron/x/epochs/types"
)

func (suite *KeeperTestSuite) TestQueryEpochInfos() {
	suite.SetupTest()
	queryClient := suite.queryClient

	// Check that querying epoch infos on default genesis returns the default genesis epoch infos
	epochInfosResponse, err := queryClient.EpochInfos(
		gocontext.Background(),
		&types.QueryEpochsInfoRequest{},
	)
	suite.Require().NoError(err)
	suite.Require().Len(epochInfosResponse.Epochs, 3)

	expectedEpochs := types.DefaultGenesis().Epochs
	for id := range expectedEpochs {
		expectedEpochs[id].StartTime = suite.Ctx.BlockTime()
		expectedEpochs[id].CurrentEpochStartHeight = suite.Ctx.BlockHeight()
		expectedEpochs[id].CurrentEpoch = 1
		expectedEpochs[id].EpochCountingStarted = true
	}

	suite.Require().Equal(expectedEpochs, epochInfosResponse.Epochs)
}
