package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/x/interchainqueries/keeper"
	iqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
)

func (suite *KeeperTestSuite) TestRemoteLastHeight() {
	tests := []struct {
		name string
		run  func()
	}{
		{
			"wrong connection id",
			func() {
				ctx := suite.ChainA.GetContext()
				_, err := keeper.Keeper.LastRemoteHeight(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper, sdk.WrapSDKContext(ctx), &iqtypes.QueryLastRemoteHeight{ConnectionId: "test"})
				suite.Require().Error(err)
			},
		},
		{
			"valid request",
			func() {
				ctx := suite.ChainA.GetContext()

				oldHeight, err := keeper.Keeper.LastRemoteHeight(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper, sdk.WrapSDKContext(ctx), &iqtypes.QueryLastRemoteHeight{ConnectionId: suite.Path.EndpointA.ConnectionID})
				suite.Require().NoError(err)
				suite.Require().Greater(oldHeight.Height, uint64(0))

				// update client N times
				N := uint64(100)
				for i := uint64(0); i < N; i++ {
					suite.NoError(suite.Path.EndpointA.UpdateClient())
				}

				updatedHeight, err := keeper.Keeper.LastRemoteHeight(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper, sdk.WrapSDKContext(ctx), &iqtypes.QueryLastRemoteHeight{ConnectionId: suite.Path.EndpointA.ConnectionID})
				suite.Require().Equal(updatedHeight.Height, oldHeight.Height+N) // check that last remote height really equals oldHeight+N
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i, len(tests)), func() {
			suite.SetupTest()
			tc.run()
		})
	}
}
