package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

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

func (suite *KeeperTestSuite) TestRegisteredQueries() {
	tests := []struct {
		name                  string
		registeredQueries     []iqtypes.RegisteredQuery
		req                   *iqtypes.QueryRegisteredQueriesRequest
		expectedQueryResponse []iqtypes.RegisteredQuery
	}{
		{
			name: "all",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:           1,
					Owner:        "owner1",
					ConnectionId: "connection-0",
				},
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner2",
					ConnectionId: "connection-1",
				},
			},
			req: &iqtypes.QueryRegisteredQueriesRequest{
				Owners:       nil,
				ConnectionId: "",
				Pagination:   nil,
			},
			expectedQueryResponse: []iqtypes.RegisteredQuery{
				{
					Id:           1,
					Owner:        "owner1",
					ConnectionId: "connection-0",
				},
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner2",
					ConnectionId: "connection-1",
				},
			},
		},
		{
			name: "limit 2 offset 1",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:           1,
					Owner:        "owner1",
					ConnectionId: "connection-0",
				},
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner2",
					ConnectionId: "connection-1",
				},
				{
					Id:           4,
					Owner:        "owner2",
					ConnectionId: "connection-2",
				},
			},
			req: &iqtypes.QueryRegisteredQueriesRequest{
				Owners:       nil,
				ConnectionId: "",
				Pagination: &query.PageRequest{
					Key:        nil,
					Offset:     1,
					Limit:      2,
					CountTotal: false,
					Reverse:    false,
				},
			},
			expectedQueryResponse: []iqtypes.RegisteredQuery{
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner2",
					ConnectionId: "connection-1",
				},
			},
		},
		{
			name: "limit 2 with key instead of offset 1",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:           1,
					Owner:        "owner1",
					ConnectionId: "connection-0",
				},
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner2",
					ConnectionId: "connection-1",
				},
				{
					Id:           4,
					Owner:        "owner2",
					ConnectionId: "connection-2",
				},
			},
			req: &iqtypes.QueryRegisteredQueriesRequest{
				Owners:       nil,
				ConnectionId: "",
				Pagination: &query.PageRequest{
					Key:        iqtypes.GetRegisteredQueryByIDKey(2)[1:], // cut out the store key cause the key is for substore
					Offset:     0,
					Limit:      2,
					CountTotal: false,
					Reverse:    false,
				},
			},
			expectedQueryResponse: []iqtypes.RegisteredQuery{
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner2",
					ConnectionId: "connection-1",
				},
			},
		},
		{
			name: "filter by owner1",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:           1,
					Owner:        "owner1",
					ConnectionId: "connection-0",
				},
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
				{
					Id:           4,
					Owner:        "owner2",
					ConnectionId: "connection-2",
				},
				{
					Id:           5,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
			},
			req: &iqtypes.QueryRegisteredQueriesRequest{
				Owners:       []string{"owner1"},
				ConnectionId: "",
				Pagination: &query.PageRequest{
					Key:        nil,
					Offset:     0,
					Limit:      0,
					CountTotal: false,
					Reverse:    false,
				},
			},
			expectedQueryResponse: []iqtypes.RegisteredQuery{
				{
					Id:           1,
					Owner:        "owner1",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
				{
					Id:           5,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
			},
		},
		{
			name: "filter by owner1 offset 2",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:           1,
					Owner:        "owner1",
					ConnectionId: "connection-0",
				},
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
				{
					Id:           4,
					Owner:        "owner2",
					ConnectionId: "connection-2",
				},
				{
					Id:           5,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
			},
			req: &iqtypes.QueryRegisteredQueriesRequest{
				Owners:       []string{"owner1"},
				ConnectionId: "",
				Pagination: &query.PageRequest{
					Key:        nil,
					Offset:     2,
					Limit:      0,
					CountTotal: false,
					Reverse:    false,
				},
			},
			expectedQueryResponse: []iqtypes.RegisteredQuery{
				{
					Id:           5,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
			},
		},
		{
			name: "filter by connection-1 offset 2",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:           1,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
				{
					Id:           4,
					Owner:        "owner2",
					ConnectionId: "connection-2",
				},
				{
					Id:           5,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
			},
			req: &iqtypes.QueryRegisteredQueriesRequest{
				Owners:       nil,
				ConnectionId: "connection-1",
				Pagination: &query.PageRequest{
					Key:        nil,
					Offset:     2,
					Limit:      0,
					CountTotal: false,
					Reverse:    false,
				},
			},
			expectedQueryResponse: []iqtypes.RegisteredQuery{
				{
					Id:           5,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
			},
		},
		{
			name: "filter by connection-1, owner2 and offset 1",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:           1,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
				{
					Id:           2,
					Owner:        "owner2",
					ConnectionId: "connection-0",
				},
				{
					Id:           3,
					Owner:        "owner2",
					ConnectionId: "connection-1",
				},
				{
					Id:           4,
					Owner:        "owner2",
					ConnectionId: "connection-1",
				},
				{
					Id:           5,
					Owner:        "owner1",
					ConnectionId: "connection-1",
				},
			},
			req: &iqtypes.QueryRegisteredQueriesRequest{
				Owners:       []string{"owner2"},
				ConnectionId: "connection-1",
				Pagination: &query.PageRequest{
					Key:        nil,
					Offset:     1,
					Limit:      0,
					CountTotal: false,
					Reverse:    false,
				},
			},
			expectedQueryResponse: []iqtypes.RegisteredQuery{
				{
					Id:           4,
					Owner:        "owner2",
					ConnectionId: "connection-1",
				},
			},
		},
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i, len(tests)), func() {
			suite.SetupTest()

			for _, q := range tt.registeredQueries {
				suite.NoError(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.SaveQuery(suite.ChainA.GetContext(), q))
			}

			resp, err := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.RegisteredQueries(sdk.WrapSDKContext(suite.ChainA.GetContext()), tt.req)
			suite.NoError(err)

			suite.Equal(tt.expectedQueryResponse, resp.RegisteredQueries)
		})
	}
}
