package keeper_test

import (
	"fmt"

	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	"github.com/cosmos/cosmos-sdk/types/query"

	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/neutron-org/neutron/v6/x/interchainqueries/keeper"
	iqtypes "github.com/neutron-org/neutron/v6/x/interchainqueries/types"
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
				_, err := keeper.Keeper.LastRemoteHeight(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper, ctx, &iqtypes.QueryLastRemoteHeight{ConnectionId: "test"})
				suite.Require().Error(err)
			},
		},
		{
			"valid request",
			func() {
				ctx := suite.ChainA.GetContext()

				oldHeight, err := keeper.Keeper.LastRemoteHeight(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper, ctx, &iqtypes.QueryLastRemoteHeight{ConnectionId: suite.Path.EndpointA.ConnectionID})
				suite.Require().NoError(err)
				suite.Require().Greater(oldHeight.Height, uint64(0))

				// update client N times
				N := uint64(100)
				for i := uint64(0); i < N; i++ {
					suite.NoError(suite.Path.EndpointA.UpdateClient())
				}

				updatedHeight, err := keeper.Keeper.LastRemoteHeight(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper, ctx, &iqtypes.QueryLastRemoteHeight{ConnectionId: suite.Path.EndpointA.ConnectionID})
				suite.Require().Equal(updatedHeight.Height, oldHeight.Height+N) // check that last remote height really equals oldHeight+N
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i+1, len(tests)), func() {
			suite.SetupTest()
			tc.run()
		})
	}
}

func (suite *KeeperTestSuite) TestRegisteredQueries() {
	height := ibcclienttypes.NewHeight(1, 1)
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
					Id:                              1,
					Owner:                           "owner1",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner2",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
			},
			req: &iqtypes.QueryRegisteredQueriesRequest{
				Owners:       nil,
				ConnectionId: "",
				Pagination:   nil,
			},
			expectedQueryResponse: []iqtypes.RegisteredQuery{
				{
					Id:                              1,
					Owner:                           "owner1",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner2",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
			},
		},
		{
			name: "limit 2 offset 1",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:                              1,
					Owner:                           "owner1",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner2",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              4,
					Owner:                           "owner2",
					ConnectionId:                    "connection-2",
					LastSubmittedResultRemoteHeight: &height,
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
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner2",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
			},
		},
		{
			name: "limit 2 with key instead of offset 1",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:                              1,
					Owner:                           "owner1",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner2",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              4,
					Owner:                           "owner2",
					ConnectionId:                    "connection-2",
					LastSubmittedResultRemoteHeight: &height,
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
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner2",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
			},
		},
		{
			name: "filter by owner1",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:                              1,
					Owner:                           "owner1",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              4,
					Owner:                           "owner2",
					ConnectionId:                    "connection-2",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              5,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
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
					Id:                              1,
					Owner:                           "owner1",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              5,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
			},
		},
		{
			name: "filter by owner1 offset 2",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:                              1,
					Owner:                           "owner1",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              4,
					Owner:                           "owner2",
					ConnectionId:                    "connection-2",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              5,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
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
					Id:                              5,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
			},
		},
		{
			name: "filter by connection-1 offset 2",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:                              1,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              4,
					Owner:                           "owner2",
					ConnectionId:                    "connection-2",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              5,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
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
					Id:                              5,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
			},
		},
		{
			name: "filter by connection-1, owner2 and offset 1",
			registeredQueries: []iqtypes.RegisteredQuery{
				{
					Id:                              1,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              2,
					Owner:                           "owner2",
					ConnectionId:                    "connection-0",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              3,
					Owner:                           "owner2",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              4,
					Owner:                           "owner2",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
				{
					Id:                              5,
					Owner:                           "owner1",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
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
					Id:                              4,
					Owner:                           "owner2",
					ConnectionId:                    "connection-1",
					LastSubmittedResultRemoteHeight: &height,
				},
			},
		},
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i+1, len(tests)), func() {
			suite.SetupTest()

			for _, q := range tt.registeredQueries {
				suite.NoError(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.SaveQuery(suite.ChainA.GetContext(), &q))
			}

			resp, err := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.RegisteredQueries(suite.ChainA.GetContext(), tt.req)
			suite.NoError(err)

			suite.Equal(tt.expectedQueryResponse, resp.RegisteredQueries)
		})
	}
}

func (suite *KeeperTestSuite) TestQueryResult() {
	clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
	ctx := suite.ChainA.GetContext()
	contractOwner := wasmKeeper.RandomAccountAddress(suite.T())
	codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
	contractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
	registerMsg := iqtypes.MsgRegisterInterchainQuery{
		ConnectionId: suite.Path.EndpointA.ConnectionID,
		Keys: []*iqtypes.KVKey{
			{Path: ibchost.StoreKey, Key: clientKey},
		},
		QueryType:    string(iqtypes.InterchainQueryTypeKV),
		UpdatePeriod: 1,
		Sender:       contractAddress.String(),
	}
	// Top up contract address with native coins for deposit
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	suite.TopUpWallet(ctx, senderAddress, contractAddress)

	msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)
	regQuery1, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
	suite.Require().NoError(err)

	// Top up contract address with native coins for deposit
	suite.TopUpWallet(ctx, senderAddress, contractAddress)
	regQuery2, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
	suite.Require().NoError(err)

	resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
		Height: suite.ChainB.LastHeader.Header.Height - 1,
		Data:   clientKey,
		Prove:  true,
	})
	suite.Require().NoError(err)

	msg := iqtypes.MsgSubmitQueryResult{
		QueryId: regQuery1.Id,
		Sender:  contractAddress.String(),
		Result: &iqtypes.QueryResult{
			KvResults: []*iqtypes.StorageValue{{
				Key:           resp.Key,
				Proof:         resp.ProofOps,
				Value:         resp.Value,
				StoragePrefix: ibchost.StoreKey,
			}},
			// we don't have tests to test transactions proofs verification since it's a tendermint layer,
			// and we don't have access to it here
			Block:    nil,
			Height:   uint64(resp.Height), //nolint:gosec
			Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
		},
	}

	_, err = msgSrv.SubmitQueryResult(ctx, &msg)
	suite.NoError(err)

	queryResultResponse, err := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.QueryResult(ctx, &iqtypes.QueryRegisteredQueryResultRequest{
		QueryId: regQuery1.Id,
	})
	suite.NoError(err)
	// suite.Equal(msg.Result, queryResultResponse)
	// KvResults - is a list of pointers, we check it explicitly. Result should be equal, but we do not store the proofs
	expectKvResults := msg.Result.KvResults
	queryKvResult := queryResultResponse.GetResult().KvResults
	msg.Result = nil
	queryResultResponse = nil
	suite.EqualValues(msg.Result, queryResultResponse.GetResult())
	for i, kv := range expectKvResults {
		kv.Proof = nil
		suite.Equal(*kv, *queryKvResult[i])
	}
	suite.Equal(len(expectKvResults), len(queryKvResult))

	_, err = suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.QueryResult(ctx, &iqtypes.QueryRegisteredQueryResultRequest{
		QueryId: regQuery2.Id,
	})
	suite.ErrorContains(err, "no query result")

	_, err = suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.QueryResult(ctx, &iqtypes.QueryRegisteredQueryResultRequest{
		QueryId: regQuery2.Id + 1,
	})
	suite.ErrorContains(err, "invalid query id")
}
