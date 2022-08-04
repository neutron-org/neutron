package keeper_test

import (
	"fmt"
	"testing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/x/interchainqueries/keeper"
	iqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	// TestOwnerAddress defines a reusable bech32 address for testing purposes
	TestOwnerAddress = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"

	// TestVersion defines a reusable interChainAccounts version string for testing purposes
	TestVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ibctesting.FirstConnectionID,
		HostConnectionId:       ibctesting.FirstConnectionID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite *KeeperTestSuite) TestRegisterInterchainQuery() {
	suite.SetupTest()

	var msg iqtypes.MsgRegisterInterchainQuery

	tests := []struct {
		name        string
		malleate    func()
		expectedErr error
	}{
		{
			"invalid connection",
			func() {
				msg = iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "unknown",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          iqtypes.InterchainQueryTypeTX,
					ZoneId:             "id",
					UpdatePeriod:       1,
					Sender:             TestOwnerAddress,
				}
			},
			iqtypes.ErrInvalidConnectionID,
		},
		{
			"invalid query type",
			func() {
				msg = iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          "invalid_type",
					ZoneId:             "id",
					UpdatePeriod:       1,
					Sender:             TestOwnerAddress,
				}
			},
			iqtypes.ErrInvalidQueryType,
		},
		{
			"invalid transactions filter format",
			func() {
				msg = iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					TransactionsFilter: "&)(^Y(*&(*&(&(*",
					Keys:               nil,
					QueryType:          iqtypes.InterchainQueryTypeTX,
					ZoneId:             "id",
					UpdatePeriod:       1,
					Sender:             TestOwnerAddress,
				}
			},
			iqtypes.ErrInvalidQueryType,
		},
		{
			"valid",
			func() {
				msg = iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          iqtypes.InterchainQueryTypeTX,
					ZoneId:             "osmosis",
					UpdatePeriod:       1,
					Sender:             "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}
			},
			nil,
		},
	}

	for _, tt := range tests {
		suite.SetupTest()

		tt.malleate()

		msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

		res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &msg)

		if tt.expectedErr != nil {
			suite.Require().ErrorIs(err, tt.expectedErr)
			suite.Require().Nil(res)
		} else {
			suite.Require().NoError(err)
			suite.Require().NotNil(res)
		}
	}
}

func (suite *KeeperTestSuite) TestSubmitInterchainQueryResult() {
	suite.SetupTest()

	var msg iqtypes.MsgSubmitQueryResult

	tests := []struct {
		name          string
		malleate      func()
		expectedError error
	}{
		{
			"invalid query id",
			func() {
				// now we don't care what is really under the value, we just need to be sure that we can verify KV proofs
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
				resp := suite.ChainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestOwnerAddress,
					ClientId: suite.Path.EndpointA.ClientID,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height),
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidQueryID,
		},
		{
			"empty result",
			func() {
				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          iqtypes.InterchainQueryTypeTX,
					ZoneId:             "osmosis",
					UpdatePeriod:       1,
					Sender:             "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					ClientId: suite.Path.EndpointA.ClientID,
				}
			},
			iqtypes.ErrEmptyResult,
		},
		{
			"empty kv results and blocks",
			func() {
				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          iqtypes.InterchainQueryTypeTX,
					ZoneId:             "osmosis",
					UpdatePeriod:       1,
					Sender:             "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					Sender:   TestOwnerAddress,
					QueryId:  res.Id,
					ClientId: suite.Path.EndpointA.ClientID,
					Result: &iqtypes.QueryResult{
						KvResults: nil,
						Block:     nil,
						Height:    0,
					},
				}
			},
			iqtypes.ErrEmptyResult,
		},
		{
			"valid KV storage proof",
			func() {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: host.StoreKey, Key: clientKey},
					},
					QueryType:    iqtypes.InterchainQueryTypeKV,
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				//suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp := suite.ChainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.Path.EndpointA.ClientID,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height),
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			nil,
		},
		{
			"non-registered key in KV result",
			func() {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: host.StoreKey, Key: clientKey},
					},
					QueryType:    iqtypes.InterchainQueryTypeKV,
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				//suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp := suite.ChainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   []byte("non-registered key"),
					Prove:  true,
				})

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.Path.EndpointA.ClientID,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height),
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidSubmittedResult,
		},
		{
			"non-registered path in KV result",
			func() {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: host.StoreKey, Key: clientKey},
					},
					QueryType:    iqtypes.InterchainQueryTypeKV,
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				//suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp := suite.ChainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.Path.EndpointA.ClientID,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: "non-registered-path",
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height),
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidSubmittedResult,
		},
		{
			"non existence KV proof",
			func() {
				clientKey := []byte("non_existed_key")

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: host.StoreKey, Key: clientKey},
					},
					QueryType:    iqtypes.InterchainQueryTypeKV,
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				//suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				// now we don't care what is really under the value, we just need to be sure that we can verify KV proofs
				resp := suite.ChainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.Path.EndpointA.ClientID,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height),
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			nil,
		},
		{
			"header with invalid height",
			func() {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: host.StoreKey, Key: clientKey},
					},
					QueryType:    iqtypes.InterchainQueryTypeKV,
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp := suite.ChainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height,
					Data:   clientKey,
					Prove:  true,
				})

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.Path.EndpointA.ClientID,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height),
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			ibcclienttypes.ErrConsensusStateNotFound,
		},
		{
			"invalid KV storage value",
			func() {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: host.StoreKey, Key: clientKey},
					},
					QueryType:    iqtypes.InterchainQueryTypeKV,
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp := suite.ChainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.Path.EndpointA.ClientID,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         []byte("some evil data"),
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height),
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidProof,
		},
		{
			"query result height is too old",
			func() {

				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: host.StoreKey, Key: clientKey},
					},
					QueryType:    iqtypes.InterchainQueryTypeKV,
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				// pretend like we have a very new query result
				suite.NoError(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.UpdateLastRemoteHeight(suite.ChainA.GetContext(), res.Id, 9999))

				resp := suite.ChainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.Path.EndpointA.ClientID,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height),
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidHeight,
		},
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i, len(tests)), func() {
			suite.SetupTest()

			tt.malleate()

			msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

			res, err := msgSrv.SubmitQueryResult(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), &msg)

			if tt.expectedError != nil {
				suite.Require().ErrorIs(err, tt.expectedError)
				suite.Require().Nil(res)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			}
		})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
