package test

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/wasmbinding/bindings"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"io/ioutil"
	"testing"
)

type CustomQuerierTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite *CustomQuerierTestSuite) TestInterchainQueryResult() {
	neutron := suite.GetNeutronZoneApp(suite.ChainA)

	ctx := neutron.NewContext(true, suite.ChainA.CurrentHeader)

	// Store code and instantiate reflect contract
	owner := keeper.RandomAccountAddress(suite.T())
	codeId := suite.storeReflectCode(ctx, owner)
	contractAddress := suite.instantiateReflectContract(ctx, owner, codeId)
	suite.Require().NotEmpty(contractAddress)

	// Register and submit query result
	lastID := neutron.InterchainQueriesKeeper.GetLastRegisteredQueryKey(ctx) + 1
	neutron.InterchainQueriesKeeper.SetLastRegisteredQueryKey(ctx, lastID)
	registeredQuery := icqtypes.RegisteredQuery{
		Id:                lastID,
		QueryData:         `{"delegator": "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"}`,
		QueryType:         "x/staking/DelegatorDelegations",
		ZoneId:            "osmosis",
		UpdatePeriod:      1,
		ConnectionId:      suite.Path.EndpointA.ConnectionID,
		LastEmittedHeight: uint64(ctx.BlockHeight()),
	}
	neutron.InterchainQueriesKeeper.SetLastRegisteredQueryKey(ctx, lastID)
	err := neutron.InterchainQueriesKeeper.SaveQuery(ctx, registeredQuery)
	suite.Require().NoError(err)

	clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
	chainBResp := suite.ChainB.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
		Height: suite.ChainB.LastHeader.Header.Height - 1,
		Data:   clientKey,
		Prove:  true,
	})

	expectedQueryResult := &icqtypes.QueryResult{
		KvResults: []*icqtypes.StorageValue{{
			Key:           chainBResp.Key,
			Proof:         chainBResp.ProofOps,
			Value:         chainBResp.Value,
			StoragePrefix: host.StoreKey,
		}},
		// we don't have tests to test transactions proofs verification since it's a tendermint layer, and we don't have access to it here
		Block:    nil,
		Height:   uint64(chainBResp.Height),
		Revision: suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
	}
	err = neutron.InterchainQueriesKeeper.SaveKVQueryResult(ctx, lastID, expectedQueryResult)
	suite.Require().NoError(err)

	// Query interchain query result
	query := bindings.NeutronQuery{
		InterchainQueryResult: &bindings.QueryRegisteredQueryResultRequest{
			QueryId: lastID,
		},
	}
	resp := icqtypes.QueryRegisteredQueryResultResponse{}
	err = suite.queryCustom(ctx, contractAddress, query, &resp)
	suite.Require().NoError(err)

	suite.Require().Equal(uint64(chainBResp.Height), resp.Result.Height)
	suite.Require().Equal(suite.ChainA.LastHeader.GetHeight().GetRevisionNumber(), resp.Result.Revision)
	suite.Require().Empty(resp.Result.Block)
	suite.Require().NotEmpty(resp.Result.KvResults)
	suite.Require().Equal([]*icqtypes.StorageValue{{
		Key:           chainBResp.Key,
		Proof:         nil,
		Value:         chainBResp.Value,
		StoragePrefix: host.StoreKey,
	}}, resp.Result.KvResults)
}

func (suite *CustomQuerierTestSuite) TestInterchainQueryResultNotFound() {
	neutron := suite.GetNeutronZoneApp(suite.ChainA)

	ctx := neutron.NewContext(true, suite.ChainA.CurrentHeader)

	// Store code and instantiate reflect contract
	owner := keeper.RandomAccountAddress(suite.T())
	codeId := suite.storeReflectCode(ctx, owner)
	contractAddress := suite.instantiateReflectContract(ctx, owner, codeId)
	suite.Require().NotEmpty(contractAddress)

	// Query interchain query result
	query := bindings.NeutronQuery{
		InterchainQueryResult: &bindings.QueryRegisteredQueryResultRequest{
			QueryId: 1,
		},
	}
	resp := icqtypes.QueryRegisteredQueryResultResponse{}
	err := suite.queryCustom(ctx, contractAddress, query, &resp)
	expectedErrMag := "Generic error: Querier contract error: codespace: interchainqueries, code: 1115: query wasm contract failed"
	suite.Require().Errorf(err, expectedErrMag)
}

func (suite *CustomQuerierTestSuite) TestInterchainAccountAddress() {
	neutron := suite.GetNeutronZoneApp(suite.ChainA)

	ctx := neutron.NewContext(true, suite.ChainA.CurrentHeader)

	// Store code and instantiate reflect contract
	owner := keeper.RandomAccountAddress(suite.T())
	codeId := suite.storeReflectCode(ctx, owner)
	contractAddress := suite.instantiateReflectContract(ctx, owner, codeId)
	suite.Require().NotEmpty(contractAddress)

	query := bindings.NeutronQuery{
		InterchainAccountAddress: &bindings.QueryInterchainAccountAddressRequest{
			OwnerAddress:        testutil.TestOwnerAddress,
			InterchainAccountId: testutil.TestInterchainId,
			ConnectionId:        suite.Path.EndpointA.ConnectionID,
		},
	}
	resp := ictxtypes.QueryInterchainAccountAddressResponse{}
	err := suite.queryCustom(ctx, contractAddress, query, &resp)
	suite.Require().NoError(err)

	expected := "neutron128vd3flgem54995jslqpr9rq4zj5n0eu0rlqj9rr9a24qjf9wc9qyuvj84"
	suite.Require().Equal(expected, resp.InterchainAccountAddress)
}

func (suite *CustomQuerierTestSuite) TestUnknownInterchainAcc() {
	neutron := suite.GetNeutronZoneApp(suite.ChainA)

	ctx := neutron.NewContext(true, suite.ChainA.CurrentHeader)

	// Store code and instantiate reflect contract
	owner := keeper.RandomAccountAddress(suite.T())
	codeId := suite.storeReflectCode(ctx, owner)
	contractAddress := suite.instantiateReflectContract(ctx, owner, codeId)
	suite.Require().NotEmpty(contractAddress)

	query := bindings.NeutronQuery{
		InterchainAccountAddress: &bindings.QueryInterchainAccountAddressRequest{
			OwnerAddress:        testutil.TestOwnerAddress,
			InterchainAccountId: "wrong_account_id",
			ConnectionId:        suite.Path.EndpointA.ConnectionID,
		},
	}
	resp := ictxtypes.QueryInterchainAccountAddressResponse{}
	expectedErrorMsg := "Generic error: Querier contract error: codespace: interchaintxs, code: 1102: query wasm contract failed"
	err := suite.queryCustom(ctx, contractAddress, query, &resp)
	suite.Require().Errorf(err, expectedErrorMsg)
}

type ReflectQuery struct {
	Chain *ChainRequest `json:"chain,omitempty"`
}

type ChainRequest struct {
	Reflect wasmvmtypes.QueryRequest `json:"reflect"`
}

type ChainResponse struct {
	Data []byte `json:"data"`
}

func (suite *CustomQuerierTestSuite) queryCustom(ctx sdk.Context, contract sdk.AccAddress, request interface{}, response interface{}) error {
	msgBz, err := json.Marshal(request)
	suite.Require().NoError(err)

	query := ChainRequest{
		Reflect: wasmvmtypes.QueryRequest{Custom: msgBz},
	}

	queryBz, err := json.Marshal(query)
	if err != nil {
		return err
	}

	resBz, err := suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper.QuerySmart(ctx, contract, queryBz)
	if err != nil {
		return err
	}

	var resp ChainResponse
	err = json.Unmarshal(resBz, &resp)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Data, response)
}

func (suite *CustomQuerierTestSuite) storeReflectCode(ctx sdk.Context, addr sdk.AccAddress) uint64 {
	/// wasm file build with https://github.com/neutron-org/neutron-contracts/tree/feat/reflect-contract
	wasmCode, err := ioutil.ReadFile("../testdata/reflect.wasm")
	suite.Require().NoError(err)

	codeID, err := keeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper).Create(ctx, addr, wasmCode, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody, Address: ""})
	suite.Require().NoError(err)

	return codeID
}

func (suite *CustomQuerierTestSuite) instantiateReflectContract(ctx sdk.Context, funder sdk.AccAddress, codeID uint64) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	suite.Require().NoError(err)

	return addr
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(CustomQuerierTestSuite))
}
