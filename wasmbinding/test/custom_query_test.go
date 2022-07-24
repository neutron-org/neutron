package test

import (
	"encoding/json"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/wasmbinding/bindings"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

var defaultFunds = sdk.NewCoins(
	sdk.NewInt64Coin("stake", 100000000),
)

func TestInterchainAccountAddress(t *testing.T) {
	owner := keeper.RandomAccountAddress(t)
	// Setup IBC chains and create connection between them
	ibcStruct := testutil.SetupIBCConnection(t)
	neutron, ok := ibcStruct.ChainA.App.(*app.App)
	require.True(t, ok)

	ctx := neutron.NewContext(true, ibcStruct.ChainA.CurrentHeader)

	// Store code and instantiate reflect contract
	codeId := storeReflectCode(t, ctx, neutron, owner)
	fundAccount(t, ctx, neutron, owner, defaultFunds)
	contractAddress := instantiateReflectContract(t, ctx, neutron, owner, codeId)
	require.NotEmpty(t, contractAddress)

	// Query account address
	icaOwner, err := ictxtypes.NewICAOwner(testutil.TestOwnerAddress, "owner_id")
	require.NoError(t, err)

	query := bindings.NeutronQuery{
		InterchainAccountAddress: &bindings.InterchainAccountAddress{
			OwnerAddress: icaOwner.String(),
			ConnectionID: ibcStruct.Path.EndpointA.ConnectionID,
		},
	}
	resp := bindings.InterchainAccountAddressResponse{}
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

func fundAccount(t *testing.T, ctx sdk.Context, neutron *app.App, addr sdk.AccAddress, coins sdk.Coins) {
	err := simapp.FundAccount(
		neutron.BankKeeper,
		ctx,
		addr,
		coins,
	)
	require.NoError(t, err)
}
