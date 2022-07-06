package keeper_test

import (
	"fmt"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/lidofinance/gaia-wasm-zone/app"
	"github.com/lidofinance/gaia-wasm-zone/testutil"
	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/keeper"
	iqtypes "github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
	itypes "github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
	ictxstypes "github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"
)

var (
	// TestOwnerAddress defines a reusable bech32 address for testing purposes
	TestOwnerAddress = "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"

	// TestVersion defines a resuable interchainaccounts version string for testing purposes
	TestVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ibctesting.FirstConnectionID,
		HostConnectionId:       ibctesting.FirstConnectionID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

func init() {
	ibctesting.DefaultTestingAppInit = testutil.SetupTestingApp
}

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	path *ibctesting.Path
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	suite.path = NewICAPath(suite.chainA, suite.chainB)

	suite.coordinator.SetupConnections(suite.path)

	suite.NoError(SetupICAPath(suite.path, TestOwnerAddress))
}

func NewICAPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.PortID
	path.EndpointB.ChannelConfig.PortID = icatypes.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = TestVersion
	path.EndpointB.ChannelConfig.Version = TestVersion

	return path
}

// SetupICAPath invokes the InterchainAccounts entrypoint and subsequent channel handshake handlers
func SetupICAPath(path *ibctesting.Path, owner string) error {
	if err := RegisterInterchainAccount(path.EndpointA, owner); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenTry(); err != nil {
		return err
	}

	if err := path.EndpointA.ChanOpenAck(); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenConfirm(); err != nil {
		return err
	}

	return nil
}

// RegisterInterchainAccount is a helper function for starting the channel handshake
func RegisterInterchainAccount(endpoint *ibctesting.Endpoint, owner string) error {
	icaOwner, _ := ictxstypes.NewICAOwner(TestOwnerAddress, "owner_id")
	portID, err := icatypes.NewControllerPortID(icaOwner.String())
	if err != nil {
		return err
	}

	channelSequence := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(endpoint.Chain.GetContext())

	a, ok := endpoint.Chain.App.(*app.App)
	if !ok {
		return fmt.Errorf("not GaiaWasmZoneApp")
	}

	if err := a.ICAControllerKeeper.RegisterInterchainAccount(endpoint.Chain.GetContext(), endpoint.ConnectionID, icaOwner.String()); err != nil {
		return err
	}

	// commit state changes for proof verification
	endpoint.Chain.App.Commit()
	endpoint.Chain.NextBlock()

	// update port/channel ids
	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID

	return nil
}

func (s *KeeperTestSuite) GetGaiaWasmZoneApp(chain *ibctesting.TestChain) *app.App {
	testApp, ok := chain.App.(*app.App)
	if !ok {
		panic("not GaiaWasmZone app")
	}

	return testApp
}

func (suite *KeeperTestSuite) TestRegisterInterchainQuery() {
	suite.SetupTest()

	var msg itypes.MsgRegisterInterchainQuery

	tests := []struct {
		name        string
		malleate    func()
		expectedErr error
	}{
		{
			"invalid connection",
			func() {
				msg = itypes.MsgRegisterInterchainQuery{
					ConnectionId: "unknown",
					QueryData:    "kek",
					QueryType:    "type",
					ZoneId:       "id",
					UpdatePeriod: 1,
					Sender:       TestOwnerAddress,
				}
			},
			iqtypes.ErrInvalidConnectionID,
		},
		{
			"valid",
			func() {
				msg = itypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.path.EndpointA.ConnectionID,
					QueryData:    `{"delegator": "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"}`,
					QueryType:    "x/staking/DelegatorDelegations",
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs",
				}
			},
			nil,
		},
	}

	for _, tt := range tests {
		suite.SetupTest()

		tt.malleate()

		msgSrv := keeper.NewMsgServerImpl(suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper)

		res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.chainA.GetContext()), &msg)

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

	var msg itypes.MsgSubmitQueryResult

	tests := []struct {
		name          string
		malleate      func()
		expectedError error
	}{
		{
			"invalid query id",
			func() {
				// now we don't care what is really under the value, we just need to be sure that we can verify KV proofs
				clientKey := host.FullClientStateKey(suite.path.EndpointB.ClientID)
				resp := suite.chainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.chainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = itypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestOwnerAddress,
					ClientId: suite.path.EndpointA.ClientID,
					Result: &itypes.QueryResult{
						KvResults: []*itypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Blocks:   nil,
						Height:   uint64(resp.Height),
						Revision: suite.chainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidQueryID,
		},
		{
			"empty result",
			func() {
				registerMsg := itypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.path.EndpointA.ConnectionID,
					QueryData:    `{"delegator": "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"}`,
					QueryType:    "x/staking/DelegatorDelegations",
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.chainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				msg = itypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					ClientId: suite.path.EndpointA.ClientID,
				}
			},
			iqtypes.ErrEmptyResult,
		},
		{
			"empty kv results and blocks",
			func() {
				registerMsg := itypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.path.EndpointA.ConnectionID,
					QueryData:    `{"delegator": "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"}`,
					QueryType:    "x/staking/DelegatorDelegations",
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.chainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				msg = itypes.MsgSubmitQueryResult{
					Sender:   TestOwnerAddress,
					QueryId:  res.Id,
					ClientId: suite.path.EndpointA.ClientID,
					Result: &itypes.QueryResult{
						KvResults: nil,
						Blocks:    nil,
						Height:    0,
					},
				}
			},
			iqtypes.ErrEmptyResult,
		},
		{
			"valid KV storage proof",
			func() {
				registerMsg := itypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.path.EndpointA.ConnectionID,
					QueryData:    `{"delegator": "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"}`,
					QueryType:    "x/staking/DelegatorDelegations",
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.chainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				//suite.NoError(suite.path.EndpointB.UpdateClient())
				suite.NoError(suite.path.EndpointA.UpdateClient())

				// now we don't care what is really under the value, we just need to be sure that we can verify KV proofs
				clientKey := host.FullClientStateKey(suite.path.EndpointB.ClientID)
				resp := suite.chainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.chainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = itypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.path.EndpointA.ClientID,
					Result: &itypes.QueryResult{
						KvResults: []*itypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Blocks:   nil,
						Height:   uint64(resp.Height),
						Revision: suite.chainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			nil,
		},
		{
			"header with invalid height",
			func() {
				registerMsg := itypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.path.EndpointA.ConnectionID,
					QueryData:    `{"delegator": "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"}`,
					QueryType:    "x/staking/DelegatorDelegations",
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.chainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.path.EndpointB.UpdateClient())
				suite.NoError(suite.path.EndpointA.UpdateClient())

				clientKey := host.FullClientStateKey(suite.path.EndpointB.ClientID)
				resp := suite.chainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.chainB.LastHeader.Header.Height,
					Data:   clientKey,
					Prove:  true,
				})

				msg = itypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.path.EndpointA.ClientID,
					Result: &itypes.QueryResult{
						KvResults: []*itypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Blocks:   nil,
						Height:   uint64(resp.Height),
						Revision: suite.chainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			ibcclienttypes.ErrConsensusStateNotFound,
		},
		{
			"invalid KV storage value",
			func() {
				registerMsg := itypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.path.EndpointA.ConnectionID,
					QueryData:    `{"delegator": "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"}`,
					QueryType:    "x/staking/DelegatorDelegations",
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.chainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.path.EndpointB.UpdateClient())
				suite.NoError(suite.path.EndpointA.UpdateClient())

				clientKey := host.FullClientStateKey(suite.path.EndpointB.ClientID)
				resp := suite.chainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.chainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = itypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.path.EndpointA.ClientID,
					Result: &itypes.QueryResult{
						KvResults: []*itypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         []byte("some evil data"),
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Blocks:   nil,
						Height:   uint64(resp.Height),
						Revision: suite.chainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidProof,
		},
		{
			"query result height is too old",
			func() {

				registerMsg := itypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.path.EndpointA.ConnectionID,
					QueryData:    `{"delegator": "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"}`,
					QueryType:    "x/staking/DelegatorDelegations",
					ZoneId:       "osmosis",
					UpdatePeriod: 1,
					Sender:       "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs",
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(sdktypes.WrapSDKContext(suite.chainA.GetContext()), &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.path.EndpointB.UpdateClient())
				suite.NoError(suite.path.EndpointA.UpdateClient())

				// pretend like we have a very new query result
				suite.NoError(suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper.UpdateLastRemoteHeight(suite.chainA.GetContext(), res.Id, 9999))

				// now we don't care what is really under the value, we just need to be sure that we can verify KV proofs
				clientKey := host.FullClientStateKey(suite.path.EndpointB.ClientID)
				resp := suite.chainB.App.Query(abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
					Height: suite.chainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})

				msg = itypes.MsgSubmitQueryResult{
					QueryId:  res.Id,
					Sender:   TestOwnerAddress,
					ClientId: suite.path.EndpointA.ClientID,
					Result: &itypes.QueryResult{
						KvResults: []*itypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: host.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Blocks:   nil,
						Height:   uint64(resp.Height),
						Revision: suite.chainA.LastHeader.GetHeight().GetRevisionNumber(),
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

			msgSrv := keeper.NewMsgServerImpl(suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper)

			res, err := msgSrv.SubmitQueryResult(sdktypes.WrapSDKContext(suite.chainA.GetContext()), &msg)

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

// test that GetSubmittedTransactions with start/end works properly
func (suite *KeeperTestSuite) TestQueryTransactions() {
	queryID := uint64(1)

	ctx := suite.chainA.GetContext()
	queriesKeeper := suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper

	lastID := queriesKeeper.GetLastSubmittedTransactionIDForQuery(ctx, queryID)

	submittedTransactions := make([]*itypes.Transaction, 0)
	for i := 0; i < 10; i++ {

		tx := itypes.Transaction{
			Id:     lastID,
			Height: 0,
			Data:   append([]byte("data"), byte(i)),
		}

		suite.NoError(queriesKeeper.SaveSubmittedTransaction(ctx, queryID, lastID, tx.Height, tx.Data))
		lastID += 1

		submittedTransactions = append(submittedTransactions, &tx)
	}
	queriesKeeper.SetLastSubmittedTransactionIDForQuery(ctx, queryID, lastID)

	start, end := 4, 9

	txs, err := queriesKeeper.GetSubmittedTransactions(ctx, queryID, uint64(start), uint64(end))
	suite.NoError(err)

	suite.Equal(txs, submittedTransactions[start:end])

	// check the same but with multiple query IDS, they should not conflict with each other
	queryID = uint64(2)

	lastID = queriesKeeper.GetLastSubmittedTransactionIDForQuery(ctx, queryID)

	submittedTransactions = make([]*itypes.Transaction, 0)
	for i := 0; i < 20; i++ {

		tx := itypes.Transaction{
			Id:     lastID,
			Height: 1,
			Data:   append([]byte("another data"), byte(i)),
		}

		suite.NoError(queriesKeeper.SaveSubmittedTransaction(ctx, queryID, lastID, tx.Height, tx.Data))
		lastID += 1

		submittedTransactions = append(submittedTransactions, &tx)
	}
	queriesKeeper.SetLastSubmittedTransactionIDForQuery(ctx, queryID, lastID)

	start, end = 3, 8

	txs, err = queriesKeeper.GetSubmittedTransactions(ctx, queryID, uint64(start), uint64(end))
	suite.NoError(err)

	suite.Equal(txs, submittedTransactions[start:end])
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
