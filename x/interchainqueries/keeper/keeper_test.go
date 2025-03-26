package keeper_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/app/params"

	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/neutron-org/neutron/v6/testutil"
	"github.com/neutron-org/neutron/v6/x/interchainqueries/keeper"
	iqtypes "github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

var reflectContractPath = "../../../wasmbinding/testdata/reflect.wasm"

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite *KeeperTestSuite) TestRegisterInterchainQuery() {
	var msg iqtypes.MsgRegisterInterchainQuery

	tests := []struct {
		name         string
		topupBalance bool
		malleate     func(sender string)
		expectedErr  error
	}{
		{
			"invalid connection",
			true,
			func(sender string) {
				msg = iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "unknown",
					TransactionsFilter: "[]",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             sender,
				}
			},
			iqtypes.ErrInvalidConnectionID,
		},
		{
			"insufficient funds for deposit",
			false,
			func(sender string) {
				msg = iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					TransactionsFilter: "[]",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             sender,
				}
			},
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"not a contract address",
			false,
			func(_ string) {
				msg = iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					TransactionsFilter: "[]",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             wasmKeeper.RandomAccountAddress(suite.T()).String(),
				}
			},
			iqtypes.ErrNotContract,
		},
		{
			"invalid bech32 sender address",
			false,
			func(_ string) {
				msg = iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					TransactionsFilter: "[]",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             "notbech32",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"valid",
			true,
			func(sender string) {
				msg = iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					TransactionsFilter: "[]",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             sender,
				}
			},
			nil,
		},
	}

	for _, tt := range tests {
		suite.SetupTest()

		var (
			ctx           = suite.ChainA.GetContext()
			contractOwner = wasmKeeper.RandomAccountAddress(suite.T())
		)

		// Store code and instantiate reflect contract.
		codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
		contractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
		suite.Require().NotEmpty(contractAddress)

		err := testutil.SetupICAPath(suite.Path, contractAddress.String())
		suite.Require().NoError(err)

		tt.malleate(contractAddress.String())

		if tt.topupBalance {
			// Top up contract address with native coins for deposit
			senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
			suite.TopUpWallet(ctx, senderAddress, contractAddress)
		}

		msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

		res, err := msgSrv.RegisterInterchainQuery(ctx, &msg)

		if tt.expectedErr != nil {
			suite.Require().ErrorIs(err, tt.expectedErr)
			suite.Require().Nil(res)
		} else {
			query, _ := keeper.Keeper.RegisteredQuery(
				suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper, ctx,
				&iqtypes.QueryRegisteredQueryRequest{QueryId: 1})

			suite.Require().Equal(iqtypes.DefaultQueryDeposit, query.RegisteredQuery.Deposit)
			suite.Require().Equal(iqtypes.DefaultQuerySubmitTimeout, query.RegisteredQuery.SubmitTimeout)
			suite.Require().NoError(err)
			suite.Require().NotNil(res)
		}
	}
}

func (suite *KeeperTestSuite) TestUpdateInterchainQuery() {
	var msg iqtypes.MsgUpdateInterchainQueryRequest
	originalKVQuery := iqtypes.MsgRegisterInterchainQuery{
		QueryType: string(iqtypes.InterchainQueryTypeKV),
		Keys: []*iqtypes.KVKey{
			{
				Path: "somepath",
				Key:  []byte("somedata"),
			},
		},
		TransactionsFilter: "",
		ConnectionId:       suite.Path.EndpointA.ConnectionID,
		UpdatePeriod:       1,
		Sender:             "",
	}

	originalTXQuery := iqtypes.MsgRegisterInterchainQuery{
		QueryType:          string(iqtypes.InterchainQueryTypeTX),
		Keys:               nil,
		TransactionsFilter: "[]",
		ConnectionId:       suite.Path.EndpointA.ConnectionID,
		UpdatePeriod:       1,
		Sender:             "",
	}

	tests := []struct {
		name                  string
		malleate              func(sender string)
		expectedErr           error
		expectedPeriod        uint64
		expectedQueryKeys     []*iqtypes.KVKey
		expectedQueryTXFilter string
		query                 iqtypes.MsgRegisterInterchainQuery
	}{
		{
			"valid update period for kv",
			func(sender string) {
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:         1,
					NewKeys:         nil,
					NewUpdatePeriod: 2,
					Sender:          sender,
				}
			},
			nil,
			2,
			originalKVQuery.Keys,
			"",
			originalKVQuery,
		},
		{
			"valid update period for tx",
			func(sender string) {
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:         1,
					NewKeys:         nil,
					NewUpdatePeriod: 2,
					Sender:          sender,
				}
			},
			nil,
			2,
			nil,
			originalTXQuery.TransactionsFilter,
			originalTXQuery,
		},
		{
			"valid kv query data",
			func(sender string) {
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId: 1,
					NewKeys: []*iqtypes.KVKey{
						{
							Path: "newpath",
							Key:  []byte("newdata"),
						},
					},
					NewUpdatePeriod: 0,
					Sender:          sender,
				}
			},
			nil,
			originalKVQuery.UpdatePeriod,
			[]*iqtypes.KVKey{
				{
					Path: "newpath",
					Key:  []byte("newdata"),
				},
			},
			"",
			originalKVQuery,
		},
		{
			"valid tx filter",
			func(sender string) {
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:               1,
					NewUpdatePeriod:       0,
					NewTransactionsFilter: "[]",
					Sender:                sender,
				}
			},
			nil,
			originalTXQuery.UpdatePeriod,
			nil,
			"[]",
			originalTXQuery,
		},
		{
			"valid kv query both query keys and update period and ignore tx filter",
			func(sender string) {
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId: 1,
					NewKeys: []*iqtypes.KVKey{
						{
							Path: "newpath",
							Key:  []byte("newdata"),
						},
					},
					NewUpdatePeriod: 2,
					Sender:          sender,
				}
			},
			nil,
			2,
			[]*iqtypes.KVKey{
				{
					Path: "newpath",
					Key:  []byte("newdata"),
				},
			},
			"",
			originalKVQuery,
		},
		{
			"valid tx query both tx filter and update period and ignore query keys",
			func(sender string) {
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:               1,
					NewUpdatePeriod:       2,
					NewTransactionsFilter: "[]",
					Sender:                sender,
				}
			},
			nil,
			2,
			nil,
			"[]",
			originalTXQuery,
		},
		{
			"must fail on update filter for a kv query",
			func(sender string) {
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:               1,
					NewUpdatePeriod:       2,
					NewTransactionsFilter: "[]",
					Sender:                sender,
				}
			},
			sdkerrors.ErrInvalidRequest,
			originalKVQuery.UpdatePeriod,
			originalKVQuery.Keys,
			originalKVQuery.TransactionsFilter,
			originalKVQuery,
		},
		{
			"must fail on update keys for a tx query",
			func(sender string) {
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId: 1,
					NewKeys: []*iqtypes.KVKey{
						{
							Path: "newpath",
							Key:  []byte("newdata"),
						},
					},
					NewUpdatePeriod: 2,
					Sender:          sender,
				}
			},
			sdkerrors.ErrInvalidRequest,
			originalTXQuery.UpdatePeriod,
			originalTXQuery.Keys,
			originalTXQuery.TransactionsFilter,
			originalTXQuery,
		},
		{
			"invalid query id",
			func(sender string) {
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId: 2,
					NewKeys: []*iqtypes.KVKey{
						{
							Path: "newpath",
							Key:  []byte("newdata"),
						},
					},
					NewUpdatePeriod: 2,
					Sender:          sender,
				}
			},
			iqtypes.ErrInvalidQueryID,
			originalKVQuery.UpdatePeriod,
			originalKVQuery.Keys,
			"",
			originalKVQuery,
		},
		{
			"failed due to auth error",
			func(_ string) {
				var (
					ctx           = suite.ChainA.GetContext()
					contractOwner = wasmKeeper.RandomAccountAddress(suite.T())
				)
				codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
				newContractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
				suite.Require().NotEmpty(newContractAddress)
				msg = iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:         1,
					NewKeys:         nil,
					NewUpdatePeriod: 2,
					Sender:          newContractAddress.String(),
				}
			},
			sdkerrors.ErrUnauthorized,
			originalKVQuery.UpdatePeriod,
			originalKVQuery.Keys,
			"",
			originalKVQuery,
		},
	}

	for i, tt := range tests {
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i+1, len(tests)), func() {
			suite.SetupTest()

			var (
				ctx           = suite.ChainA.GetContext()
				contractOwner = wasmKeeper.RandomAccountAddress(suite.T())
			)

			// Store code and instantiate reflect contract.
			codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
			contractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
			suite.Require().NotEmpty(contractAddress)

			err := testutil.SetupICAPath(suite.Path, contractAddress.String())
			suite.Require().NoError(err)

			// Top up contract address with native coins for deposit
			senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
			suite.TopUpWallet(ctx, senderAddress, contractAddress)

			tt.malleate(contractAddress.String())

			iqkeeper := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper

			msgSrv := keeper.NewMsgServerImpl(iqkeeper)

			tt.query.Sender = contractAddress.String()
			resRegister, err := msgSrv.RegisterInterchainQuery(ctx, &tt.query)
			suite.Require().NoError(err)
			suite.Require().NotNil(resRegister)

			resUpdate, err := msgSrv.UpdateInterchainQuery(ctx, &msg)

			if tt.expectedErr != nil {
				suite.Require().ErrorIs(err, tt.expectedErr)
				suite.Require().Nil(resUpdate)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resUpdate)
			}
			query, err := iqkeeper.GetQueryByID(ctx, 1)
			suite.Require().NoError(err)
			suite.Require().Equal(tt.expectedQueryKeys, query.GetKeys())
			suite.Require().Equal(tt.expectedQueryTXFilter, query.GetTransactionsFilter())
			suite.Require().Equal(tt.expectedPeriod, query.GetUpdatePeriod())
		})
	}
}

func (suite *KeeperTestSuite) TestRemoveInterchainQuery() {
	suite.SetupTest()

	var msg iqtypes.MsgRemoveInterchainQueryRequest
	var query iqtypes.MsgRegisterInterchainQuery
	var txQueryHashes [][]byte

	tests := []struct {
		name        string
		malleate    func(sender string)
		expectedErr error
	}{
		{
			"valid TX remove",
			func(sender string) {
				msg = iqtypes.MsgRemoveInterchainQueryRequest{
					QueryId: 1,
					Sender:  sender,
				}
				query = iqtypes.MsgRegisterInterchainQuery{
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					Keys:               nil,
					TransactionsFilter: "[]",
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					UpdatePeriod:       1,
					Sender:             "",
				}
				txQueryHashes = [][]byte{
					[]byte("txhash_1"),
					[]byte("txhash_2"),
				}
			},
			nil,
		},
		{
			"valid large TX remove",
			func(sender string) {
				msg = iqtypes.MsgRemoveInterchainQueryRequest{
					QueryId: 1,
					Sender:  sender,
				}
				query = iqtypes.MsgRegisterInterchainQuery{
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					Keys:               nil,
					TransactionsFilter: "[]",
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					UpdatePeriod:       1,
					Sender:             "",
				}
				// types.DefaultTxQueryRemovalLimit is used here for it is both big and can be
				// removed in a single tx hashes cleanup iteration
				hashesCount := int(iqtypes.DefaultTxQueryRemovalLimit) //nolint:gosec
				txQueryHashes = make([][]byte, 0, hashesCount)
				for i := 1; i <= hashesCount; i++ {
					txQueryHashes = append(txQueryHashes, []byte(fmt.Sprintf("txhash_%d", i)))
				}
			},
			nil,
		},
		{
			"valid KV remove",
			func(sender string) {
				msg = iqtypes.MsgRemoveInterchainQueryRequest{
					QueryId: 1,
					Sender:  sender,
				}
				query = iqtypes.MsgRegisterInterchainQuery{
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					Keys:               []*iqtypes.KVKey{{Key: []byte("key1"), Path: "path1"}},
					TransactionsFilter: "",
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					UpdatePeriod:       1,
					Sender:             "",
				}
			},
			nil,
		},
		{
			"invalid query id",
			func(sender string) {
				msg = iqtypes.MsgRemoveInterchainQueryRequest{
					QueryId: 2,
					Sender:  sender,
				}
				query = iqtypes.MsgRegisterInterchainQuery{
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					Keys:               []*iqtypes.KVKey{{Key: []byte("key1"), Path: "path1"}},
					TransactionsFilter: "",
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					UpdatePeriod:       1,
					Sender:             "",
				}
			},
			iqtypes.ErrInvalidQueryID,
		},
		{
			"failed due to auth error",
			func(_ string) {
				var (
					ctx           = suite.ChainA.GetContext()
					contractOwner = wasmKeeper.RandomAccountAddress(suite.T())
				)
				codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
				newContractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
				suite.Require().NotEmpty(newContractAddress)
				msg = iqtypes.MsgRemoveInterchainQueryRequest{
					QueryId: 1,
					Sender:  newContractAddress.String(),
				}
				query = iqtypes.MsgRegisterInterchainQuery{
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					Keys:               []*iqtypes.KVKey{{Key: []byte("key1"), Path: "path1"}},
					TransactionsFilter: "",
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					UpdatePeriod:       1,
					Sender:             "",
				}
			},
			sdkerrors.ErrUnauthorized,
		},
	}

	for i, tt := range tests {
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i+1, len(tests)), func() {
			suite.SetupTest()

			var (
				ctx           = suite.ChainA.GetContext()
				contractOwner = wasmKeeper.RandomAccountAddress(suite.T())
			)

			// Store code and instantiate reflect contract.
			codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
			contractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
			suite.Require().NotEmpty(contractAddress)

			err := testutil.SetupICAPath(suite.Path, contractAddress.String())
			suite.Require().NoError(err)

			// Top up contract address with native coins for deposit
			bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
			senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
			suite.TopUpWallet(ctx, senderAddress, contractAddress)

			tt.malleate(contractAddress.String())
			iqkeeper := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper

			msgSrv := keeper.NewMsgServerImpl(iqkeeper)
			query.Sender = contractAddress.String()

			resRegister, err := msgSrv.RegisterInterchainQuery(ctx, &query)
			suite.Require().NoError(err)
			suite.Require().NotNil(resRegister)

			balance, balanceErr := bankKeeper.Balance(
				ctx,
				&banktypes.QueryBalanceRequest{
					Address: contractAddress.String(),
					Denom:   params.DefaultDenom,
				},
			)
			expectedCoin := sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(0)))

			suite.Require().NoError(balanceErr)
			suite.Require().NotNil(balance)
			suite.Require().Equal(&expectedCoin, balance.Balance)

			clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
			resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
				Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
				Height: suite.ChainB.LastHeader.Header.Height - 1,
				Data:   clientKey,
				Prove:  true,
			})
			suite.Require().NoError(err)

			queryType := iqtypes.InterchainQueryType(query.GetQueryType())
			switch {
			case queryType.IsKV():
				err = iqkeeper.SaveKVQueryResult(ctx, 1, &iqtypes.QueryResult{
					KvResults: []*iqtypes.StorageValue{{
						Key:           resp.Key,
						Proof:         resp.ProofOps,
						Value:         resp.Value,
						StoragePrefix: ibchost.StoreKey,
					}},
					Block:    nil,
					Height:   1,
					Revision: 1,
				})
				suite.Require().NoError(err)
			case queryType.IsTX():
				for _, txQueryHash := range txQueryHashes {
					iqkeeper.SaveTransactionAsProcessed(ctx, 1, txQueryHash)
					suite.Require().True(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, 1, txQueryHash))
				}
			}

			respRm, err := msgSrv.RemoveInterchainQuery(ctx, &msg)
			// TxQueriesCleanup is supposed to be called in the app's EndBlock, but suite.ChainA.NextBlock()
			// passes an incorrect context to the EndBlock and thus Keeper's store is empty. So we
			// have to call it here manually and directly pass the right context into it.
			iqkeeper.TxQueriesCleanup(ctx)
			if tt.expectedErr != nil {
				suite.Require().ErrorIs(err, tt.expectedErr)
				suite.Require().Nil(respRm)
				originalQuery, queryErr := iqkeeper.GetQueryByID(ctx, 1)
				suite.Require().NoError(queryErr)
				suite.Require().NotNil(originalQuery)

				switch {
				case queryType.IsKV():
					qr, qrerr := iqkeeper.GetQueryResultByID(ctx, 1)
					suite.Require().NoError(qrerr)
					suite.Require().NotNil(qr)
				case queryType.IsTX():
					for _, txQueryHash := range txQueryHashes {
						suite.Require().True(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, 1, txQueryHash))
					}
				}
			} else {
				balance, balanceErr := bankKeeper.Balance(
					ctx,
					&banktypes.QueryBalanceRequest{
						Address: contractAddress.String(),
						Denom:   params.DefaultDenom,
					},
				)
				expectedCoin := sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(1_000_000)))

				suite.Require().NoError(balanceErr)
				suite.Require().NotNil(balance)
				suite.Require().Equal(&expectedCoin, balance.Balance)

				suite.Require().NoError(err)
				suite.Require().NotNil(respRm)
				query, queryErr := iqkeeper.GetQueryByID(ctx, 1)
				suite.Require().Error(queryErr, iqtypes.ErrInvalidQueryID)
				suite.Require().Nil(query)

				switch {
				case queryType.IsKV():
					qr, qrerr := iqkeeper.GetQueryResultByID(ctx, 1)
					suite.Require().Error(qrerr, iqtypes.ErrNoQueryResult)
					suite.Require().Nil(qr)
				case queryType.IsTX():
					for _, txQueryHash := range txQueryHashes {
						suite.Require().False(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, 1, txQueryHash))
					}
				}
			}
		})
	}
}

// Test get all registered queries
func (suite *KeeperTestSuite) TestGetAllRegisteredQueries() {
	suite.SetupTest()

	tests := []struct {
		name    string
		queries []*iqtypes.RegisteredQuery
	}{
		{
			"all registered queries",
			[]*iqtypes.RegisteredQuery{
				&(iqtypes.RegisteredQuery{
					Id:        1,
					QueryType: string(iqtypes.InterchainQueryTypeKV),
				}),
				&(iqtypes.RegisteredQuery{
					Id:        2,
					QueryType: string(iqtypes.InterchainQueryTypeKV),
				}),
			},
		},
		{
			"no registered queries",
			nil,
		},
	}

	for i, tt := range tests {
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i+1, len(tests)), func() {
			suite.SetupTest()

			ctx := suite.ChainA.GetContext()

			iqkeeper := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper
			for _, query := range tt.queries {
				err := iqkeeper.SaveQuery(ctx, query)
				suite.Require().NoError(err)
			}

			allQueries := iqkeeper.GetAllRegisteredQueries(ctx)

			suite.Require().Equal(tt.queries, allQueries)
		})
	}
}

func (suite *KeeperTestSuite) TestSubmitInterchainQueryResult() {
	var msg iqtypes.MsgSubmitQueryResult

	tests := []struct {
		name          string
		malleate      func(sender string, ctx sdk.Context)
		expectedError error
	}{
		{
			"invalid query id",
			func(sender string, ctx sdk.Context) {
				// now we don't care what is really under the value, we just need to be sure that we can verify KV proofs
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: 1,
					Sender:  sender,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: ibchost.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height), //nolint:gosec
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidQueryID,
		},
		{
			"valid KV storage proof",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				// suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
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
			},
			nil,
		},
		{
			"invalid number of KvResults",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				// suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: ibchost.StoreKey,
						}, {
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
			},
			iqtypes.ErrInvalidSubmittedResult,
		},
		{
			"invalid query type",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       suite.Path.EndpointA.ConnectionID,
					Keys:               nil,
					TransactionsFilter: "[]",
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				// suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
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
			},
			iqtypes.ErrInvalidType,
		},
		{
			"nil proof",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				// suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         nil,
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
			},
			iqtypes.ErrInvalidType,
		},
		{
			"non-registered key in KV result",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   []byte("non-registered key"),
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: ibchost.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height), //nolint:gosec
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidSubmittedResult,
		},
		{
			"non-registered path in KV result",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: "non-registered-path",
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer,
						// and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height), //nolint:gosec
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidSubmittedResult,
		},
		{
			"non existence KV proof",
			func(sender string, ctx sdk.Context) {
				clientKey := []byte("non_existed_key")

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				// suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				// now we don't care what is really under the value, we just need to be sure that we can verify KV proofs
				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender, // A bit weird that query owner submits the results, but it doesn't really matter
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
			},
			nil,
		},
		{
			"header with invalid height",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: ibchost.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height), //nolint:gosec
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			ibcclienttypes.ErrConsensusStateNotFound,
		},
		{
			"invalid KV storage value",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         []byte("some evil data"),
							StoragePrefix: ibchost.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height), //nolint:gosec
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidProof,
		},
		{
			"query result height is too old",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				// pretend like we have a very new query result
				suite.NoError(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.UpdateLastRemoteHeight(ctx, res.Id, ibcclienttypes.NewHeight(suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(), 9999)))

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: ibchost.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:    nil,
						Height:   uint64(resp.Height), //nolint:gosec
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidHeight,
		},
		{
			"query result revision number check",
			func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: clientKey},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				// pretend like we have a very new query result
				suite.NoError(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.UpdateLastRemoteHeight(ctx, res.Id, ibcclienttypes.NewHeight(suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(), 9999)))

				// pretend like we have a very new query result with updated revision height
				suite.NoError(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper.UpdateLastRemoteHeight(ctx, res.Id, ibcclienttypes.NewHeight(suite.ChainA.LastHeader.GetHeight().GetRevisionNumber()+1, 1)))

				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key:           resp.Key,
							Proof:         resp.ProofOps,
							Value:         resp.Value,
							StoragePrefix: ibchost.StoreKey,
						}},
						// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
						Block:  nil,
						Height: uint64(resp.Height), //nolint:gosec
						// we forecefully "updated" revision height
						Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
					},
				}
			},
			iqtypes.ErrInvalidHeight,
		},
		// in this test we check that storageValue.Key with special bytes (characters) can be properly verified
		{
			"non existence KV proof with special bytes in key",
			func(sender string, ctx sdk.Context) {
				keyWithSpecialBytes, err := hex.DecodeString("0220c746274d3fe20c2c9d06c017e15f8e03f92598fca39d7540aab02244073efe26756a756e6f78")
				suite.Require().NoError(err)

				registerMsg := iqtypes.MsgRegisterInterchainQuery{
					ConnectionId: suite.Path.EndpointA.ConnectionID,
					Keys: []*iqtypes.KVKey{
						{Path: ibchost.StoreKey, Key: keyWithSpecialBytes},
					},
					QueryType:    string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod: 1,
					Sender:       sender,
				}

				msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

				res, err := msgSrv.RegisterInterchainQuery(ctx, &registerMsg)
				suite.Require().NoError(err)

				// suite.NoError(suite.Path.EndpointB.UpdateClient())
				suite.NoError(suite.Path.EndpointA.UpdateClient())

				// now we don't care what is really under the value, we just need to be sure that we can verify KV proofs
				resp, err := suite.ChainB.App.Query(ctx, &abci.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainB.LastHeader.Header.Height - 1,
					Data:   keyWithSpecialBytes,
					Prove:  true,
				})
				suite.Require().NoError(err)

				msg = iqtypes.MsgSubmitQueryResult{
					QueryId: res.Id,
					Sender:  sender, // A bit weird that query owner submits the results, but it doesn't really matter
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
			},
			nil,
		},
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i+1, len(tests)), func() {
			suite.SetupTest()

			var (
				ctx           = suite.ChainA.GetContext()
				contractOwner = wasmKeeper.RandomAccountAddress(suite.T())
			)

			// Store code and instantiate reflect contract.
			codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
			contractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
			suite.Require().NotEmpty(contractAddress)

			err := testutil.SetupICAPath(suite.Path, contractAddress.String())
			suite.Require().NoError(err)

			// Top up contract address with native coins for deposit
			senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
			suite.TopUpWallet(ctx, senderAddress, contractAddress)

			tt.malleate(contractAddress.String(), ctx)

			msgSrv := keeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper)

			res, err := msgSrv.SubmitQueryResult(ctx, &msg)

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

func (suite *KeeperTestSuite) TestTxQueriesCleanup() {
	suite.Run("SingleIterSingleQuery", func() {
		suite.SetupTest()
		iqkeeper := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper
		ctx := suite.ChainA.GetContext()

		// create a query and add results for it
		var queryID uint64 = 1
		query := iqtypes.RegisteredQuery{Id: queryID, QueryType: string(iqtypes.InterchainQueryTypeTX)}
		err := iqkeeper.SaveQuery(ctx, &query)
		suite.Require().NoError(err)
		_, err = iqkeeper.GetQueryByID(ctx, queryID)
		suite.Require().Nil(err)
		txHashes := suite.buildTxHashes(50)
		for _, hash := range txHashes {
			iqkeeper.SaveTransactionAsProcessed(ctx, queryID, hash)
			suite.Require().True(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID, hash))
		}

		// remove query and call cleanup
		iqkeeper.RemoveQuery(ctx, &query)
		iqkeeper.TxQueriesCleanup(ctx)

		// make sure removal and cleanup worked as expected
		for _, hash := range txHashes {
			suite.Require().Falsef(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID, hash), "%s expected not to be in the store", hash)
		}
		suite.Require().Nilf(iqkeeper.GetTxQueriesToRemove(ctx, 0), "expected not to have any TX queries to remove after cleanup")
		_, err = iqkeeper.GetQueryByID(ctx, queryID)
		suite.Require().ErrorIs(err, iqtypes.ErrInvalidQueryID)
	})

	suite.Run("SingleIterMultipeQueries", func() {
		suite.SetupTest()
		iqkeeper := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper
		ctx := suite.ChainA.GetContext()

		txHashes := suite.buildTxHashes(100)
		// create a query and add results for it
		var queryID1 uint64 = 1
		query1 := iqtypes.RegisteredQuery{Id: queryID1, QueryType: string(iqtypes.InterchainQueryTypeTX)}
		err := iqkeeper.SaveQuery(ctx, &query1)
		suite.Require().NoError(err)
		_, err = iqkeeper.GetQueryByID(ctx, queryID1)
		suite.Require().NoError(err)
		txHashesQ1 := txHashes[:len(txHashes)/2] // first half of the build hashes come to the first query
		for _, hash := range txHashesQ1 {
			iqkeeper.SaveTransactionAsProcessed(ctx, queryID1, hash)
			suite.Require().True(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID1, hash))
		}
		// create another query and add results for it
		var queryID2 uint64 = 2
		query2 := iqtypes.RegisteredQuery{Id: queryID2, QueryType: string(iqtypes.InterchainQueryTypeTX)}
		err = iqkeeper.SaveQuery(ctx, &query2)
		suite.Require().NoError(err)
		_, err = iqkeeper.GetQueryByID(ctx, queryID2)
		suite.Require().NoError(err)
		txHashesQ2 := txHashes[len(txHashes)/2:] // second half of the build hashes come to the second query
		for _, hash := range txHashesQ2 {
			iqkeeper.SaveTransactionAsProcessed(ctx, queryID2, hash)
			suite.Require().True(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID2, hash))
		}

		// remove queries and call cleanup
		iqkeeper.RemoveQuery(ctx, &query1)
		iqkeeper.RemoveQuery(ctx, &query2)
		iqkeeper.TxQueriesCleanup(ctx)

		// make sure removal and cleanup worked as expected
		for _, hash := range txHashesQ1 {
			suite.Require().Falsef(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID1, hash), "%s expected not to be in the store", hash)
		}
		for _, hash := range txHashesQ2 {
			suite.Require().Falsef(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID2, hash), "%s expected not to be in the store", hash)
		}
		suite.Require().Nilf(iqkeeper.GetTxQueriesToRemove(ctx, 0), "expected not to have any TX queries to remove after cleanup")
		_, err = iqkeeper.GetQueryByID(ctx, queryID1)
		suite.Require().ErrorIs(err, iqtypes.ErrInvalidQueryID)
		_, err = iqkeeper.GetQueryByID(ctx, queryID2)
		suite.Require().ErrorIs(err, iqtypes.ErrInvalidQueryID)
	})

	suite.Run("MultipleIterSingleQuery", func() {
		suite.SetupTest()
		iqkeeper := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper
		ctx := suite.ChainA.GetContext()

		// set TxQueryRemovalLimit to a low value
		limit := 50
		params := iqkeeper.GetParams(ctx)
		params.TxQueryRemovalLimit = uint64(limit) //nolint:gosec
		err := iqkeeper.SetParams(ctx, params)
		suite.Require().NoError(err)

		// create a query and add results for it
		var queryID uint64 = 1
		query := iqtypes.RegisteredQuery{Id: queryID, QueryType: string(iqtypes.InterchainQueryTypeTX)}
		err = iqkeeper.SaveQuery(ctx, &query)
		suite.Require().NoError(err)
		_, err = iqkeeper.GetQueryByID(ctx, queryID)
		suite.Require().NoError(err)
		limitOverflow := 10
		txHashes := suite.buildTxHashes(limit + limitOverflow) // create a bit more hashes than the limit
		for _, hash := range txHashes {
			iqkeeper.SaveTransactionAsProcessed(ctx, queryID, hash)
			suite.Require().True(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID, hash))
		}

		// remove query and call cleanup
		iqkeeper.RemoveQuery(ctx, &query)
		iqkeeper.TxQueriesCleanup(ctx)

		// make sure removal and cleanup worked as expected
		removed, left := suite.txHashesRemovalProgress(ctx, iqkeeper, queryID, txHashes)
		suite.Require().Equalf(limit, removed, "first cleanup removed hashes count should be as many as limit")
		suite.Require().Equalf(limitOverflow, left, "first cleanup left hashes count should be as many as limitOverflow")
		suite.Require().Equalf([]uint64{queryID}, iqkeeper.GetTxQueriesToRemove(ctx, 0), "expected to have a TX query to remove after partial cleanup")
		_, err = iqkeeper.GetQueryByID(ctx, queryID)
		suite.Require().ErrorIs(err, iqtypes.ErrInvalidQueryID)

		// call cleanup one more time and make sure it worked as expected
		iqkeeper.TxQueriesCleanup(ctx)
		for _, hash := range txHashes { // by this point all hashes should be removed
			suite.Require().Falsef(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID, hash), "%s expected not to be in the store", hash)
		}
		suite.Require().Nilf(iqkeeper.GetTxQueriesToRemove(ctx, 0), "expected not to have any TX queries to remove after cleanup")
		_, err = iqkeeper.GetQueryByID(ctx, queryID)
		suite.Require().ErrorIs(err, iqtypes.ErrInvalidQueryID)
	})

	suite.Run("MultipleIterMultipeQueries", func() {
		suite.SetupTest()
		iqkeeper := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper
		ctx := suite.ChainA.GetContext()

		// set TxQueryRemovalLimit to a low value
		limit := 50
		params := iqkeeper.GetParams(ctx)
		params.TxQueryRemovalLimit = uint64(limit) //nolint:gosec
		err := iqkeeper.SetParams(ctx, params)
		suite.Require().NoError(err)

		limitOverflow := 10
		txHashes := suite.buildTxHashes(limit + limitOverflow)
		txHashesQ1 := txHashes[:len(txHashes)/2] // first half of the build hashes come to the first query
		txHashesQ2 := txHashes[len(txHashes)/2:] // second half of the build hashes come to the second query
		// create a query and add results for it
		var queryID1 uint64 = 1
		query1 := iqtypes.RegisteredQuery{Id: queryID1, QueryType: string(iqtypes.InterchainQueryTypeTX)}
		err = iqkeeper.SaveQuery(ctx, &query1)
		suite.Require().NoError(err)
		_, err = iqkeeper.GetQueryByID(ctx, queryID1)
		suite.Require().NoError(err)
		for _, hash := range txHashesQ1 {
			iqkeeper.SaveTransactionAsProcessed(ctx, queryID1, hash)
			suite.Require().True(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID1, hash))
		}
		// create another query and add results for it
		var queryID2 uint64 = 2
		query2 := iqtypes.RegisteredQuery{Id: queryID2, QueryType: string(iqtypes.InterchainQueryTypeTX)}
		err = iqkeeper.SaveQuery(ctx, &query2)
		suite.Require().NoError(err)
		_, err = iqkeeper.GetQueryByID(ctx, queryID2)
		suite.Require().NoError(err)
		for _, hash := range txHashesQ2 {
			iqkeeper.SaveTransactionAsProcessed(ctx, queryID2, hash)
			suite.Require().True(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID2, hash))
		}

		// remove queries and call cleanup
		iqkeeper.RemoveQuery(ctx, &query1)
		iqkeeper.RemoveQuery(ctx, &query2)
		iqkeeper.TxQueriesCleanup(ctx)

		// make sure removal and cleanup worked as expected
		removedQ1, leftQ1 := suite.txHashesRemovalProgress(ctx, iqkeeper, queryID1, txHashesQ1)
		removedQ2, leftQ2 := suite.txHashesRemovalProgress(ctx, iqkeeper, queryID2, txHashesQ2)
		suite.Require().Equalf(limit, removedQ1+removedQ2, "first cleanup removed hashes count should be as many as limit")
		suite.Require().Equalf(limitOverflow, leftQ1+leftQ2, "first cleanup remaining hashes count should be as many as limitOverflow")
		suite.Require().Equalf([]uint64{queryID2}, iqkeeper.GetTxQueriesToRemove(ctx, 0), "expected to have one TX query to remove after partial cleanup")

		// call cleanup one more time and make sure it worked as expected
		iqkeeper.TxQueriesCleanup(ctx)
		removedQ1, leftQ1 = suite.txHashesRemovalProgress(ctx, iqkeeper, queryID1, txHashesQ1)
		removedQ2, leftQ2 = suite.txHashesRemovalProgress(ctx, iqkeeper, queryID2, txHashesQ2)
		suite.Require().Equalf(limit+limitOverflow, removedQ1+removedQ2, "all hashes should be removed after the second cleanup")
		suite.Require().Equalf(0, leftQ1+leftQ2, "no hashes should left after the second cleanup")
		suite.Require().Nilf(iqkeeper.GetTxQueriesToRemove(ctx, 0), "expected not to have any TX queries to remove after cleanup")
		_, err = iqkeeper.GetQueryByID(ctx, queryID1)
		suite.Require().ErrorIs(err, iqtypes.ErrInvalidQueryID)
		_, err = iqkeeper.GetQueryByID(ctx, queryID2)
		suite.Require().ErrorIs(err, iqtypes.ErrInvalidQueryID)
	})

	suite.Run("Unlimited", func() {
		suite.SetupTest()
		iqkeeper := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper
		ctx := suite.ChainA.GetContext()

		// set TxQueryRemovalLimit to a low value
		params := iqkeeper.GetParams(ctx)
		params.TxQueryRemovalLimit = 0
		err := iqkeeper.SetParams(ctx, params)
		suite.Require().NoError(err)
		suite.Require().Equal(uint64(0), iqkeeper.GetParams(ctx).TxQueryRemovalLimit)

		// create a query and add results for it
		var queryID uint64 = 1
		query := iqtypes.RegisteredQuery{Id: queryID, QueryType: string(iqtypes.InterchainQueryTypeTX)}
		err = iqkeeper.SaveQuery(ctx, &query)
		suite.Require().NoError(err)
		_, err = iqkeeper.GetQueryByID(ctx, queryID)
		suite.Require().NoError(err)
		txHashes := suite.buildTxHashes(int(iqtypes.DefaultTxQueryRemovalLimit) * 2) //nolint:gosec
		for _, hash := range txHashes {
			iqkeeper.SaveTransactionAsProcessed(ctx, queryID, hash)
			suite.Require().True(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID, hash))
		}

		// remove query and call cleanup
		iqkeeper.RemoveQuery(ctx, &query)
		iqkeeper.TxQueriesCleanup(ctx)

		// make sure removal and cleanup worked as expected
		for _, hash := range txHashes {
			suite.Require().Falsef(iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID, hash), "%s expected not to be in the store", hash)
		}
		suite.Require().Nilf(iqkeeper.GetTxQueriesToRemove(ctx, 0), "expected not to have any TX queries to remove after cleanup")
		_, err = iqkeeper.GetQueryByID(ctx, queryID)
		suite.Require().ErrorIs(err, iqtypes.ErrInvalidQueryID)
	})
}

// TestRemoveFreshlyCreatedICQ mostly makes sure the query's RegisteredAtHeight field works.
func (suite *KeeperTestSuite) TestRemoveFreshlyCreatedICQ() {
	suite.SetupTest()
	var (
		ctx           = suite.ChainA.GetContext()
		contractOwner = wasmKeeper.RandomAccountAddress(suite.T())
	)

	// Store code and instantiate reflect contract.
	codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
	contractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
	suite.Require().NotEmpty(contractAddress)

	// Top up contract address with native coins for deposit
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	suite.TopUpWallet(ctx, senderAddress, contractAddress)

	iqkeeper := suite.GetNeutronZoneApp(suite.ChainA).InterchainQueriesKeeper
	params := iqkeeper.GetParams(ctx)
	params.QuerySubmitTimeout = 5
	err := iqkeeper.SetParams(ctx, params)
	suite.Require().NoError(err)
	msgSrv := keeper.NewMsgServerImpl(iqkeeper)

	resRegister, err := msgSrv.RegisterInterchainQuery(ctx, &iqtypes.MsgRegisterInterchainQuery{
		QueryType:          string(iqtypes.InterchainQueryTypeKV),
		Keys:               []*iqtypes.KVKey{{Key: []byte("key1"), Path: "path1"}},
		TransactionsFilter: "",
		ConnectionId:       suite.Path.EndpointA.ConnectionID,
		UpdatePeriod:       1,
		Sender:             contractAddress.String(),
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(resRegister)

	registeredQuery, err := iqkeeper.GetQueryByID(ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(ctx.BlockHeight()), registeredQuery.RegisteredAtHeight) //nolint:gosec
	suite.Require().Equal(uint64(0), registeredQuery.LastSubmittedResultLocalHeight)
	suite.Require().Equal(params.QuerySubmitTimeout, registeredQuery.SubmitTimeout)
	suite.Require().Greater(uint64(ctx.BlockHeight()), registeredQuery.LastSubmittedResultLocalHeight+registeredQuery.SubmitTimeout) //nolint:gosec

	newContractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
	suite.Require().NotEmpty(newContractAddress)
	resp, err := msgSrv.RemoveInterchainQuery(ctx, &iqtypes.MsgRemoveInterchainQueryRequest{
		QueryId: 1,
		Sender:  newContractAddress.String(),
	})
	suite.Nil(resp)
	suite.ErrorContains(err, "only owner can remove a query within its service period")
}

func (suite *KeeperTestSuite) TopUpWallet(ctx sdk.Context, sender, contractAddress sdk.AccAddress) {
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(1_000_000))))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
	suite.Require().NoError(err)
}

// buildTxHashes generates the given amount of fake tx hashes.
func (*KeeperTestSuite) buildTxHashes(amount int) [][]byte {
	txHashes := make([][]byte, 0, amount)
	for i := 1; i <= amount; i++ {
		txHashes = append(txHashes, []byte(fmt.Sprintf("tx_hash_%d", i)))
	}
	return txHashes
}

// txHashesRemovalProgress calculates how many hashes have been removed and how many are left in the
// keeper's store for the given query.
func (*KeeperTestSuite) txHashesRemovalProgress(ctx sdk.Context, iqkeeper keeper.Keeper, queryID uint64, initHashes [][]byte) (removed, left int) {
	for _, hash := range initHashes {
		if iqkeeper.CheckTransactionIsAlreadyProcessed(ctx, queryID, hash) {
			left++
		} else {
			removed++
		}
	}
	return removed, left
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
