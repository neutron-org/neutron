package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/wasmbinding/bindings"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

type CustomQuerierTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite *CustomQuerierTestSuite) TestInterchainQueryResult() {
	var (
		neutron = suite.GetNeutronZoneApp(suite.ChainA)
		ctx     = suite.ChainA.GetContext()
		owner   = keeper.RandomAccountAddress(suite.T()) // We don't care what this address is
	)

	// Store code and instantiate reflect contract
	codeId := suite.StoreReflectCode(ctx, owner, "../testdata/reflect.wasm")
	contractAddress := suite.InstantiateReflectContract(ctx, owner, codeId)
	suite.Require().NotEmpty(contractAddress)

	// Register and submit query result
	clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
	lastID := neutron.InterchainQueriesKeeper.GetLastRegisteredQueryKey(ctx) + 1
	neutron.InterchainQueriesKeeper.SetLastRegisteredQueryKey(ctx, lastID)
	registeredQuery := icqtypes.RegisteredQuery{
		Id: lastID,
		Keys: []*icqtypes.KVKey{
			{Path: host.StoreKey, Key: clientKey},
		},
		QueryType:    string(icqtypes.InterchainQueryTypeKV),
		UpdatePeriod: 1,
		ConnectionId: suite.Path.EndpointA.ConnectionID,
	}
	neutron.InterchainQueriesKeeper.SetLastRegisteredQueryKey(ctx, lastID)
	err := neutron.InterchainQueriesKeeper.SaveQuery(ctx, registeredQuery)
	suite.Require().NoError(err)

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
			QueryID: lastID,
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
	var (
		ctx   = suite.ChainA.GetContext()
		owner = keeper.RandomAccountAddress(suite.T()) // We don't care what this address is
	)

	// Store code and instantiate reflect contract
	codeId := suite.StoreReflectCode(ctx, owner, "../testdata/reflect.wasm")
	contractAddress := suite.InstantiateReflectContract(ctx, owner, codeId)
	suite.Require().NotEmpty(contractAddress)

	// Query interchain query result
	query := bindings.NeutronQuery{
		InterchainQueryResult: &bindings.QueryRegisteredQueryResultRequest{
			QueryID: 1,
		},
	}
	resp := icqtypes.QueryRegisteredQueryResultResponse{}
	err := suite.queryCustom(ctx, contractAddress, query, &resp)
	expectedErrMsg := fmt.Sprintf("Generic error: Querier contract error: codespace: interchainqueries, code: %d: query wasm contract failed", icqtypes.ErrNoQueryResult.ABCICode())
	suite.Require().ErrorContains(err, expectedErrMsg)
}

func (suite *CustomQuerierTestSuite) TestInterchainAccountAddress() {
	var (
		ctx   = suite.ChainA.GetContext()
		owner = keeper.RandomAccountAddress(suite.T()) // We don't care what this address is
	)

	// Store code and instantiate reflect contract
	codeId := suite.StoreReflectCode(ctx, owner, "../testdata/reflect.wasm")
	contractAddress := suite.InstantiateReflectContract(ctx, owner, codeId)
	suite.Require().NotEmpty(contractAddress)

	err := testutil.SetupICAPath(suite.Path, contractAddress.String())
	suite.Require().NoError(err)

	query := bindings.NeutronQuery{
		InterchainAccountAddress: &bindings.QueryInterchainAccountAddressRequest{
			OwnerAddress:        contractAddress.String(),
			InterchainAccountID: testutil.TestInterchainID,
			ConnectionID:        suite.Path.EndpointA.ConnectionID,
		},
	}
	resp := ictxtypes.QueryInterchainAccountAddressResponse{}
	err = suite.queryCustom(ctx, contractAddress, query, &resp)
	suite.Require().NoError(err)

	hostNeutronApp, ok := suite.ChainB.App.(*app.App)
	suite.Require().True(ok)

	expected := hostNeutronApp.ICAHostKeeper.GetAllInterchainAccounts(suite.ChainB.GetContext())[0].AccountAddress // we expect only one registered ICA
	suite.Require().Equal(expected, resp.InterchainAccountAddress)
}

func (suite *CustomQuerierTestSuite) TestUnknownInterchainAcc() {
	var (
		ctx   = suite.ChainA.GetContext()
		owner = keeper.RandomAccountAddress(suite.T()) // We don't care what this address is
	)

	// Store code and instantiate reflect contract
	codeId := suite.StoreReflectCode(ctx, owner, "../testdata/reflect.wasm")
	contractAddress := suite.InstantiateReflectContract(ctx, owner, codeId)
	suite.Require().NotEmpty(contractAddress)

	err := testutil.SetupICAPath(suite.Path, contractAddress.String())
	suite.Require().NoError(err)

	query := bindings.NeutronQuery{
		InterchainAccountAddress: &bindings.QueryInterchainAccountAddressRequest{
			OwnerAddress:        testutil.TestOwnerAddress,
			InterchainAccountID: "wrong_account_id",
			ConnectionID:        suite.Path.EndpointA.ConnectionID,
		},
	}
	resp := ictxtypes.QueryInterchainAccountAddressResponse{}
	expectedErrorMsg := "Generic error: Querier contract error: codespace: interchaintxs, code: 1102: query wasm contract failed"

	err = suite.queryCustom(ctx, contractAddress, query, &resp)
	suite.Require().ErrorContains(err, expectedErrorMsg)
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(CustomQuerierTestSuite))
}
