package test

import (
	"encoding/json"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/wasmbinding/bindings"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
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
	testFromAddress = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"
	//testConnectionId = "connection-0"
	//testInterchainAccountId = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"
)

func TestInterchainAccountAddress(t *testing.T) {
	//cfg := app.GetDefaultConfig()
	//cfg.Seal()
	owner := keeper.RandomAccountAddress(t)
	//neutron, ctx := SetupCustomApp(t, owner) // TODO: probably don't need same address to be the owner of reflected contract, so try with other address
	testingstruct := testutil.SetupIBCConnection(t)
	testConnectionId := testingstruct.Path.EndpointA.ConnectionID
	neutron, ok := testingstruct.ChainA.App.(*app.App)
	require.True(t, ok)
	ctx := neutron.NewContext(true, tmproto.Header{Height: neutron.LastBlockHeight()})

	storeReflectCode(t, ctx, neutron, owner)
	cInfo := neutron.WasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	fundAccount(t, ctx, neutron, owner, defaultFunds)

	// Register interchain account
	msg := ictxtypes.MsgRegisterInterchainAccount{
		FromAddress:         testFromAddress,  // contract address
		ConnectionId:        testConnectionId, // new connection id
		InterchainAccountId: testFromAddress,  // owner
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
	wasmCode, err := ioutil.ReadFile("../testdata/reflect.wasm")
	require.NoError(t, err)

	// TODO: rewrite using this
	//codeId, err := keeper.NewDefaultPermissionKeeper(neutron.WasmKeeper).Create(ctx, addr, wasmCode, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody, Address: ""})

	src := wasmtypes.StoreCodeProposalFixture(func(p *wasmtypes.StoreCodeProposal) {
		p.RunAs = addr.String()
		p.WASMByteCode = wasmCode
	})
	govKeeper.SetProposalID(ctx, govtypes.DefaultStartingProposalID)
	govKeeper.SetDepositParams(ctx, govtypes.DefaultDepositParams())
	govKeeper.SetVotingParams(ctx, govtypes.DefaultVotingParams())
	govKeeper.SetTallyParams(ctx, govtypes.DefaultTallyParams())
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

func SetupCustomApp(t *testing.T, addr sdk.AccAddress) (*app.App, sdk.Context) {
	neutron, _ := setupTestingApp()
	ctx := neutron.NewContext(true, tmproto.Header{Height: neutron.LastBlockHeight()})
	wasmKeeper := neutron.WasmKeeper

	//wasmparams := wasmKeeper.GetParams(ctx)
	//wasmparams.CodeUploadAccess = wasmtypes.AllowEverybody
	neutron.WasmKeeper.SetParams(ctx, wasmtypes.DefaultParams())
	//paramsTest := wasmKeeper.GetParams(ctx)
	//err := wasmKeeper.Va
	//panic(neutron.WasmKeeper.GetParams(ctx))

	storeReflectCode(t, ctx, neutron, addr)
	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	return neutron, ctx
}

// setupTestingApp initializes the IBC-go testing application
func setupTestingApp() (*app.App, map[string]json.RawMessage) {
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
		//app.GetEnabledProposals(),
		wasm.EnableAllProposals,
		simapp.EmptyAppOptions{},
		nil,
	)
	genesisState := app.NewDefaultGenesisState(testApp.AppCodec())
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	testApp.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	return testApp, genesisState
}
