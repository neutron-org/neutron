package keeper_test

import (
	"encoding/json"
	"fmt"
	"os"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	dextypes "github.com/neutron-org/neutron/v6/x/dex/types"
	icqtypes "github.com/neutron-org/neutron/v6/x/interchainqueries/types"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
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

type SendMsgTestCase struct {
	desc       string
	msg        func(denom string) *banktypes.MsgSend
	expectPass bool
}

func (suite *KeeperTestSuite) TestBeforeSendHook() {
	for _, tc := range []struct {
		desc     string
		wasmFile string
		sendMsgs []SendMsgTestCase
	}{
		{
			desc:     "should not allow sending 100 amount of *any* denom",
			wasmFile: "./testdata/no100.wasm", // https://github.com/neutron-org/neutron-dev-contracts/tree/chore/additional-tf-test-contracts/contracts/no100
			sendMsgs: []SendMsgTestCase{
				{
					desc: "sending 1 of factorydenom should not error",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 1)),
						)
					},
					expectPass: true,
				},
				{
					desc: "sending 1 of non-factorydenom should not error",
					msg: func(_ string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin("foo", 1)),
						)
					},
					expectPass: true,
				},
				{
					desc: "sending 100 of factorydenom should error",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 100)),
						)
					},
					expectPass: false,
				},
				{
					desc: "sending 100 of non-factorydenom should work",
					msg: func(_ string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin("foo", 100)),
						)
					},
					expectPass: true,
				},
				{
					desc: "having 100 coin within coins should not work",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							suite.TestAccs[0],
							suite.TestAccs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 100), sdk.NewInt64Coin("foo", 1)),
						)
					},
					expectPass: false,
				},
			},
		},
	} {
		suite.Run(fmt.Sprintf("Case %suite", tc.desc), func() {
			// setup test
			suite.Setup()

			senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
			suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[0])

			// upload and instantiate wasm code
			wasmCode, err := os.ReadFile(tc.wasmFile)
			suite.Require().NoError(err, "test: %v", tc.desc)
			codeID, _, err := suite.contractKeeper.Create(suite.ChainA.GetContext(), suite.TestAccs[0], wasmCode, nil)
			suite.Require().NoError(err, "test: %v", tc.desc)
			cosmwasmAddress, _, err := suite.contractKeeper.Instantiate(suite.ChainA.GetContext(), codeID, suite.TestAccs[0], suite.TestAccs[0], []byte("{}"), "", sdk.NewCoins())
			suite.Require().NoError(err, "test: %v", tc.desc)

			// create new denom
			res, err := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
			suite.Require().NoError(err, "test: %v", tc.desc)
			denom := res.GetNewTokenDenom()

			// mint enough coins to the creator
			_, err = suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 1000000000)))
			suite.Require().NoError(err)
			// mint some non token factory denom coins for testing
			suite.FundAcc(sdk.MustAccAddressFromBech32(suite.TestAccs[0].String()), sdk.Coins{sdk.NewInt64Coin("foo", 100000000000)})

			params := types.DefaultParams()
			params.WhitelistedHooks = []*types.WhitelistedHook{{DenomCreator: suite.TestAccs[0].String(), CodeID: 1}}
			err = suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper.SetParams(suite.ChainA.GetContext(), params)
			suite.Require().NoError(err)

			// set beforesend hook to the new denom
			_, err = suite.msgServer.SetBeforeSendHook(suite.ChainA.GetContext(), types.NewMsgSetBeforeSendHook(suite.TestAccs[0].String(), denom, cosmwasmAddress.String()))
			suite.Require().NoError(err, "test: %v", tc.desc)

			for _, sendTc := range tc.sendMsgs {
				_, err := suite.bankMsgServer.Send(suite.ChainA.GetContext(), sendTc.msg(denom))
				if sendTc.expectPass {
					suite.Require().NoError(err, "test: %v", sendTc.desc)
				} else {
					suite.Require().Error(err, "test: %v", sendTc.desc)
				}
			}
		})
	}
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

// TestInfiniteTrackBeforeSend tests gas metering with infinite loop contract
// to properly test if we are gas metering trackBeforeSend properly.
func (suite *KeeperTestSuite) TestInfiniteTrackBeforeSend() {
	for _, tc := range []struct {
		name            string
		wasmFile        string
		tokenToSend     sdk.Coins
		useFactoryDenom bool
		blockBeforeSend bool
		expectedError   bool
	}{
		{
			name:            "sending tokenfactory denom from module to module with infinite contract should panic",
			wasmFile:        "./testdata/infinite_track_beforesend.wasm", // https://github.com/neutron-org/neutron-dev-contracts/tree/chore/additional-tf-test-contracts/contracts/infinite-track-beforesend
			useFactoryDenom: true,
		},
		{
			name:            "sending tokenfactory denom from account to account with infinite contract should panic",
			wasmFile:        "./testdata/infinite_track_beforesend.wasm",
			useFactoryDenom: true,
			blockBeforeSend: true,
		},
		{
			name:            "sending non-tokenfactory denom from module to module with infinite contract should not panic",
			wasmFile:        "./testdata/infinite_track_beforesend.wasm",
			tokenToSend:     sdk.NewCoins(sdk.NewInt64Coin("foo", 1000000)),
			useFactoryDenom: false,
		},
		{
			name:            "Try using no 100",
			wasmFile:        "./testdata/no100.wasm", // https://github.com/neutron-org/neutron-dev-contracts/tree/chore/additional-tf-test-contracts/contracts/no100
			useFactoryDenom: true,
		},
	} {
		suite.Run(fmt.Sprintf("Case %suite", tc.name), func() {
			// setup test
			suite.Setup()

			// load wasm file
			wasmCode, err := os.ReadFile(tc.wasmFile)
			suite.Require().NoError(err)

			// instantiate wasm code
			codeID, _, err := suite.contractKeeper.Create(suite.ChainA.GetContext(), suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress(), wasmCode, nil)
			suite.Require().NoError(err, "test: %v", tc.name)
			cosmwasmAddress, _, err := suite.contractKeeper.Instantiate(suite.ChainA.GetContext(), codeID, suite.TestAccs[0], suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress(), []byte("{}"), "", sdk.NewCoins())
			suite.Require().NoError(err, "test: %v", tc.name)

			params := types.DefaultParams()
			params.WhitelistedHooks = []*types.WhitelistedHook{{DenomCreator: suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress().String(), CodeID: 1}}
			err = suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper.SetParams(suite.ChainA.GetContext(), params)
			suite.Require().NoError(err)

			// create new denom
			res, err := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress().String(), "bitcoin"))
			suite.Require().NoError(err, "test: %v", tc.name)
			factoryDenom := res.GetNewTokenDenom()

			var tokenToSend sdk.Coins
			if tc.useFactoryDenom {
				tokenToSend = sdk.NewCoins(sdk.NewInt64Coin(factoryDenom, 100))
			} else {
				tokenToSend = tc.tokenToSend
			}

			// send the mint module tokenToSend
			if tc.blockBeforeSend {
				suite.FundAcc(suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress(), tokenToSend)
			} else {
				suite.FundModuleAcc(icqtypes.ModuleName, tokenToSend)
			}

			// set beforesend hook to the new denom
			// we register infinite loop contract here to test if we are gas metering properly
			_, err = suite.msgServer.SetBeforeSendHook(suite.ChainA.GetContext(), types.NewMsgSetBeforeSendHook(suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress().String(), factoryDenom, cosmwasmAddress.String()))
			suite.Require().NoError(err, "test: %v", tc.name)

			ctx := suite.ChainA.GetContext().WithGasMeter(storetypes.NewGasMeter(30_000_000))

			if tc.blockBeforeSend {
				err = suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.SendCoins(ctx, suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress(), suite.ChainA.SenderAccounts[1].SenderAccount.GetAddress(), tokenToSend)
				suite.Require().Error(err)
			} else {
				// track before send suppresses in any case, thus we expect no error
				err = suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.SendCoinsFromModuleToModule(ctx, icqtypes.ModuleName, dextypes.ModuleName, tokenToSend)
				suite.Require().NoError(err)

				// send should happen regardless of trackBeforeSend results
				distributionModuleAddress := suite.GetNeutronZoneApp(suite.ChainA).AccountKeeper.GetModuleAddress(dextypes.ModuleName)
				distributionModuleBalances := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetAllBalances(suite.ChainA.GetContext(), distributionModuleAddress)
				suite.Require().True(distributionModuleBalances.Equal(tokenToSend))
			}
		})
	}
}
