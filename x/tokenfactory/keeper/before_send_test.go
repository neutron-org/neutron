package keeper_test

import (
	"encoding/json"
	"os"

	"cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) initBalanceTrackContract(denom string) (sdk.AccAddress, uint64, string) {
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[0])

	// https://github.com/neutron-org/neutron-dev-contracts/tree/feat/balance-tracker-contract/contracts/balance-tracker
	wasmFile := "./testdata/balance_tracker.wasm"

	// load wasm file
	wasmCode, err := os.ReadFile(wasmFile)
	suite.Require().NoError(err)

	// create new denom
	res, err := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), denom))
	suite.Require().NoError(err)
	factoryDenom := res.GetNewTokenDenom()

	// instantiate wasm code
	tokenFactoryModuleAddr := suite.GetNeutronZoneApp(suite.ChainA).AccountKeeper.GetModuleAddress("tokenfactory")
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper)
	codeID, _, err := contractKeeper.Create(suite.ChainA.GetContext(), suite.TestAccs[0], wasmCode, nil)
	suite.Require().NoError(err)
	initMsg, _ := json.Marshal(
		map[string]interface{}{
			"tracked_denom":               factoryDenom,
			"tokenfactory_module_address": tokenFactoryModuleAddr,
		},
	)
	cosmwasmAddress, _, err := contractKeeper.Instantiate(
		suite.ChainA.GetContext(), codeID, suite.TestAccs[0], suite.TestAccs[0], initMsg, "", sdk.NewCoins(),
	)
	suite.Require().NoError(err)

	return cosmwasmAddress, codeID, factoryDenom
}

func (suite *KeeperTestSuite) TestTrackBeforeSendWasm() {
	suite.Setup()

	cosmwasmAddress, codeID, factoryDenom := suite.initBalanceTrackContract("testdenom")

	// Whitelist the hook
	params := types.DefaultParams()
	params.WhitelistedHooks = []*types.WhitelistedHook{{DenomCreator: suite.TestAccs[0].String(), CodeID: codeID}}
	err := suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper.SetParams(suite.ChainA.GetContext(), params)
	suite.Require().NoError(err)

	// set beforeSendHook for the new denom
	_, err = suite.msgServer.SetBeforeSendHook(suite.ChainA.GetContext(), types.NewMsgSetBeforeSendHook(suite.TestAccs[0].String(), factoryDenom, cosmwasmAddress.String()))
	suite.Require().NoError(err)

	tokenToSend := sdk.NewCoin(factoryDenom, math.NewInt(100))

	// mint tokens
	_, err = suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), tokenToSend))
	suite.Require().NoError(err)

	queryResp, err := suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper.QuerySmart(suite.ChainA.GetContext(), cosmwasmAddress, []byte(`{"total_supply_at":{}}`))
	suite.Require().NoError(err)
	// Whitelisted contract is called correctly
	suite.Require().Equal("\"100\"", string(queryResp))
}

func (suite *KeeperTestSuite) TestAddNonWhitelistedHooksFails() {
	suite.Setup()

	cosmwasmAddress, _, factoryDenom := suite.initBalanceTrackContract("testdenom")

	// WHEN to set beforeSendHook
	_, err := suite.msgServer.SetBeforeSendHook(suite.ChainA.GetContext(), types.NewMsgSetBeforeSendHook(suite.TestAccs[0].String(), factoryDenom, cosmwasmAddress.String()))
	// THEN fails with BeforeSendHookNotWhitelisted
	suite.Require().ErrorIs(err, types.ErrBeforeSendHookNotWhitelisted)
}

func (suite *KeeperTestSuite) TestNonWhitelistedHooksNotCalled() {
	suite.Setup()

	cosmwasmAddress, codeID, factoryDenom := suite.initBalanceTrackContract("testdenom")

	// Whitelist the hook first so it can be added
	params := types.DefaultParams()
	params.WhitelistedHooks = []*types.WhitelistedHook{{DenomCreator: suite.TestAccs[0].String(), CodeID: codeID}}
	err := suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper.SetParams(suite.ChainA.GetContext(), params)
	suite.Require().NoError(err)

	// set beforeSendHook for the new denom
	_, err = suite.msgServer.SetBeforeSendHook(suite.ChainA.GetContext(), types.NewMsgSetBeforeSendHook(suite.TestAccs[0].String(), factoryDenom, cosmwasmAddress.String()))
	suite.Require().NoError(err)

	// Remove whitelist for the hook
	params = types.DefaultParams()
	err = suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper.SetParams(suite.ChainA.GetContext(), params)
	suite.Require().NoError(err)

	tokenToSend := sdk.NewCoin(factoryDenom, math.NewInt(100))

	// mint tokens
	_, err = suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), tokenToSend))
	suite.Require().NoError(err)

	queryResp, err := suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper.QuerySmart(suite.ChainA.GetContext(), cosmwasmAddress, []byte(`{"total_supply_at":{}}`))
	suite.Require().NoError(err)
	// Whitelisted contract is not called
	suite.Require().Equal("\"0\"", string(queryResp))
}
