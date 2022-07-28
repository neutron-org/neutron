package test

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/wasmbinding/bindings"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"io/ioutil"
	"testing"
)

func init() {
	config := app.GetDefaultConfig()
	config.Seal()
}

func TestInterchainQueryResult(t *testing.T) {
	// Setup IBC chains and create connection between them
	ibcStruct := testutil.SetupIBCConnection(t)
	neutron, ok := ibcStruct.ChainA.App.(*app.App)
	require.True(t, ok)

	ctx := neutron.NewContext(true, ibcStruct.ChainA.CurrentHeader)

	// Store code and instantiate reflect contract
	owner := keeper.RandomAccountAddress(t)
	codeId := storeReflectCode(t, ctx, neutron, owner)
	contractAddress := instantiateReflectContract(t, ctx, neutron, owner, codeId)
	require.NotEmpty(t, contractAddress)

	// Register and submit query result
	lastID := neutron.InterchainQueriesKeeper.GetLastRegisteredQueryKey(ctx) + 1
	neutron.InterchainQueriesKeeper.SetLastRegisteredQueryKey(ctx, lastID)
	registeredQuery := icqtypes.RegisteredQuery{
		Id:                lastID,
		QueryData:         `{"delegator": "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"}`,
		QueryType:         "x/staking/DelegatorDelegations",
		ZoneId:            "osmosis",
		UpdatePeriod:      1,
		ConnectionId:      ibcStruct.Path.EndpointA.ConnectionID,
		LastEmittedHeight: uint64(ctx.BlockHeight()),
	}
	neutron.InterchainQueriesKeeper.SetLastRegisteredQueryKey(ctx, lastID)
	err := neutron.InterchainQueriesKeeper.SaveQuery(ctx, registeredQuery)
	require.NoError(t, err)

	clientKey := host.FullClientStateKey(ibcStruct.Path.EndpointB.ClientID)
	chainBResp := ibcStruct.ChainB.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
		Height: ibcStruct.ChainB.LastHeader.Header.Height - 1,
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
		Blocks:   nil,
		Height:   uint64(chainBResp.Height),
		Revision: ibcStruct.ChainA.LastHeader.GetHeight().GetRevisionNumber(),
	}
	err = neutron.InterchainQueriesKeeper.SaveQueryResult(ctx, lastID, expectedQueryResult)
	require.NoError(t, err)

	// Query interchain query result
	query := bindings.NeutronQuery{
		InterchainQueryResult: &icqtypes.QueryRegisteredQueryResultRequest{
			QueryId: lastID,
		},
	}
	resp := icqtypes.QueryRegisteredQueryResultResponse{}
	queryCustom(t, ctx, neutron, contractAddress, query, &resp)

	require.EqualValues(t, uint64(chainBResp.Height), resp.Result.Height)
	require.EqualValues(t, ibcStruct.ChainA.LastHeader.GetHeight().GetRevisionNumber(), resp.Result.Revision)
	require.Empty(t, resp.Result.Blocks)
	require.NotEmpty(t, resp.Result.KvResults)
	require.EqualValues(t, []*icqtypes.StorageValue{{
		Key:           chainBResp.Key,
		Proof:         nil,
		Value:         chainBResp.Value,
		StoragePrefix: host.StoreKey,
	}}, resp.Result.KvResults)
}

func TestInterchainAccountAddress(t *testing.T) {
	// Setup IBC chains and create connection between them
	ibcStruct := testutil.SetupIBCConnection(t)
	neutron, ok := ibcStruct.ChainA.App.(*app.App)
	require.True(t, ok)

	ctx := neutron.NewContext(true, ibcStruct.ChainA.CurrentHeader)

	// Store code and instantiate reflect contract
	owner := keeper.RandomAccountAddress(t)
	codeId := storeReflectCode(t, ctx, neutron, owner)
	contractAddress := instantiateReflectContract(t, ctx, neutron, owner, codeId)
	require.NotEmpty(t, contractAddress)

	// Query real account address
	icaOwner, err := ictxtypes.NewICAOwner(testutil.TestOwnerAddress, testutil.TestInterchainId)
	require.NoError(t, err)

	query := bindings.NeutronQuery{
		InterchainAccountAddress: &ictxtypes.QueryInterchainAccountAddressRequest{
			OwnerAddress: icaOwner.String(),
			ConnectionId: ibcStruct.Path.EndpointA.ConnectionID,
		},
	}
	resp := ictxtypes.QueryInterchainAccountAddressResponse{}
	queryCustom(t, ctx, neutron, contractAddress, query, &resp)

	expected := "neutron128vd3flgem54995jslqpr9rq4zj5n0eu0rlqj9rr9a24qjf9wc9qyuvj84"
	require.EqualValues(t, expected, resp.InterchainAccountAddress)
}

type ReflectQuery struct {
	Chain *ChainRequest `json:"chain,omitempty"`
}

type ChainRequest struct {
	Request wasmvmtypes.QueryRequest `json:"request"`
}

type ChainResponse struct {
	Data []byte `json:"data"`
}

func queryCustom(t *testing.T, ctx sdk.Context, neutron *app.App, contract sdk.AccAddress, request bindings.NeutronQuery, response interface{}) {
	msgBz, err := json.Marshal(request)
	require.NoError(t, err)

	query := ReflectQuery{
		Chain: &ChainRequest{
			Request: wasmvmtypes.QueryRequest{Custom: msgBz},
		},
	}
	queryBz, err := json.Marshal(query)
	require.NoError(t, err)

	resBz, err := neutron.WasmKeeper.QuerySmart(ctx, contract, queryBz)
	require.NoError(t, err)
	var resp ChainResponse
	err = json.Unmarshal(resBz, &resp)
	require.NoError(t, err)
	err = json.Unmarshal(resp.Data, response)
	require.NoError(t, err)
}

func storeReflectCode(t *testing.T, ctx sdk.Context, neutron *app.App, addr sdk.AccAddress) uint64 {
	wasmCode, err := ioutil.ReadFile("../testdata/reflect.wasm")
	require.NoError(t, err)

	codeID, err := keeper.NewDefaultPermissionKeeper(neutron.WasmKeeper).Create(ctx, addr, wasmCode, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody, Address: ""})
	require.NoError(t, err)

	return codeID
}

func instantiateReflectContract(t *testing.T, ctx sdk.Context, neutron *app.App, funder sdk.AccAddress, codeID uint64) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(neutron.WasmKeeper)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	require.NoError(t, err)

	return addr
}
