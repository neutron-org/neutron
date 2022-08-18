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
				_, err := keeper.Keeper.LastRemoteHeight(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper, sdk.WrapSDKContext(ctx), &iqtypes.QueryLastRemoteHeight{ConnectionId: "connection-0"})
				suite.Require().NoError(err)
			},
		},
		// {
		// 	"valid request",
		// 	func() {
		// 		ctx := suite.ChainA.GetContext()
		// 		_, err := keeper.Keeper.LastRemoteHeight(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper, ctx.Context(), &iqtypes.QueryLastRemoteHeight{ConnectionId: "t"})
		// 		// suite.Require().Greater(uint64(resp.Height), 0)
		// 		suite.Require().NoError(err)
		// 	},
		// },
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i, len(tests)), func() {
			suite.SetupTest()
			tc.run()
		})
	}
}
