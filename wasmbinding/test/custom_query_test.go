package test

import (
	"encoding/json"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/wasmbinding/bindings"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"io/ioutil"
	"testing"
)

var defaultFunds = sdk.NewCoins(
	sdk.NewInt64Coin("stake", 100000000),
)

var (
	testFromAddress  = "test_from_address"
	testConnectionId = "connection-0"
	//testInterchainAccountId = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"
)

func TestInterchainAccountAddress(t *testing.T) {
	owner := RandomAccountAddress()
	neutron, ctx := SetupCustomApp(t, owner) // TODO: probably don't need same address to be the owner of reflected contract, so try with other address

	fundAccount(t, ctx, neutron, owner, defaultFunds)

	// Register interchain account
	msg := ictxtypes.MsgRegisterInterchainAccount{
		FromAddress:         testFromAddress,  // contract address
		ConnectionId:        testConnectionId, // new connection id
		InterchainAccountId: owner.String(),   // owner
	}
	_, err := neutron.InterchainTxsKeeper.RegisterInterchainAccount(sdk.WrapSDKContext(ctx), &msg)
	require.NoError(t, err)

	// Instantiate reflect contract
	reflect := instantiateReflectContract(t, ctx, neutron, owner)
	require.NotEmpty(t, reflect)

	// query account address
	query := bindings.NeutronQuery{
		InterchainAccountAddress: &bindings.InterchainAccountAddress{
			OwnerAddress: owner.String(),
			ConnectionID: testConnectionId,
		},
	}
	resp := bindings.InterchainAccountAddressResponse{}
	queryCustom(t, ctx, neutron, reflect, query, &resp)

	expected := "KEKW"
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

func storeReflectCode(t *testing.T, ctx sdk.Context, neutron *app.App, addr sdk.AccAddress) {
	govKeeper := neutron.GovKeeper
	wasmCode, err := ioutil.ReadFile("../testdata/neutron_reflect.wasm")
	require.NoError(t, err)

	src := wasmtypes.StoreCodeProposalFixture(func(p *wasmtypes.StoreCodeProposal) {
		p.RunAs = addr.String()
		p.WASMByteCode = wasmCode
	})

	// when stored
	storedProposal, err := govKeeper.SubmitProposal(ctx, src)
	require.NoError(t, err)

	// and proposal execute
	handler := govKeeper.Router().GetRoute(storedProposal.ProposalRoute())
	err = handler(ctx, storedProposal.GetContent())
	require.NoError(t, err)
}

func instantiateReflectContract(t *testing.T, ctx sdk.Context, neutron *app.App, funder sdk.AccAddress) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(neutron.WasmKeeper)
	codeID := uint64(1)
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

// we need to make this deterministic (same every test run), as content might affect gas costs
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func RandomAccountAddress() sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

func RandomBech32AccountAddress() string {
	return RandomAccountAddress().String()
}

func SetupCustomApp(t *testing.T, addr sdk.AccAddress) (*app.App, sdk.Context) {
	neutron, _ := SetupTestingApp()
	ctx := neutron.NewContext(true, tmproto.Header{Height: neutron.LastBlockHeight()})
	wasmKeeper := neutron.WasmKeeper

	storeReflectCode(t, ctx, neutron, addr)

	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	return neutron, ctx
}

// SetupTestingApp initializes the IBC-go testing application
func SetupTestingApp() (*app.App, map[string]json.RawMessage) {
	encoding := app.MakeEncodingConfig()
	db := dbm.NewMemDB()
	testApp := app.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		app.GetEnabledProposals(),
		simapp.EmptyAppOptions{},
		nil,
	)
	return testApp, app.NewDefaultGenesisState(testApp.AppCodec())
}
