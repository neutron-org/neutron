package keeper_test

import (
	"encoding/json"
	"fmt"
	"os"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dextypes "github.com/neutron-org/neutron/v3/x/dex/types"
	"github.com/neutron-org/neutron/v3/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestTrackBeforeSendWasm() {
	for _, tc := range []struct {
		name     string
		wasmFile string
	}{
		{
			name: "Test bank hook tracking contract ",
			// https://github.com/neutron-org/neutron-dev-contracts/tree/feat/balance-tracker-contract/contracts/balance-tracker
			wasmFile: "./testdata/balance_tracker.wasm",
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			// setup test
			suite.Setup()

			senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
			suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[0])

			// load wasm file
			wasmCode, err := os.ReadFile(tc.wasmFile)
			suite.Require().NoError(err)

			// create new denom
			res, err := suite.msgServer.CreateDenom(sdk.WrapSDKContext(suite.ChainA.GetContext()), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "testdenom"))
			suite.Require().NoError(err, "test: %v", tc.name)
			factoryDenom := res.GetNewTokenDenom()

			// instantiate wasm code
			tokenFactoryModuleAddr := suite.GetNeutronZoneApp(suite.ChainA).AccountKeeper.GetModuleAddress("tokenfactory")
			contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper)
			codeID, _, err := contractKeeper.Create(suite.ChainA.GetContext(), suite.TestAccs[0], wasmCode, nil)
			suite.Require().NoError(err, "test: %v", tc.name)
			initMsg, _ := json.Marshal(
				map[string]interface{}{
					"tracked_denom":               factoryDenom,
					"tokenfactory_module_address": tokenFactoryModuleAddr,
				},
			)
			cosmwasmAddress, _, err := contractKeeper.Instantiate(
				suite.ChainA.GetContext(), codeID, suite.TestAccs[0], suite.TestAccs[0], initMsg, "", sdk.NewCoins(),
			)
			suite.Require().NoError(err, "test: %v", tc.name)

			// set beforesend hook to the new denom
			_, err = suite.msgServer.SetBeforeSendHook(sdk.WrapSDKContext(suite.ChainA.GetContext()), types.NewMsgSetBeforeSendHook(suite.TestAccs[0].String(), factoryDenom, cosmwasmAddress.String()))
			suite.Require().NoError(err, "test: %v", tc.name)

			tokenToSend := sdk.NewCoin(factoryDenom, sdk.NewInt(100))

			// mint tokens
			_, err = suite.msgServer.Mint(sdk.WrapSDKContext(suite.ChainA.GetContext()), types.NewMsgMint(suite.TestAccs[0].String(), tokenToSend))
			suite.Require().NoError(err)

			queryResp, err := suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper.QuerySmart(suite.ChainA.GetContext(), cosmwasmAddress, []byte(`{"total_supply_at":{}}`))
			suite.Require().NoError(err)
			suite.Require().Equal("\"100\"", string(queryResp))
		})
	}
}

func (suite *KeeperTestSuite) TestDisableBlockBeforeSend() {
	// setup test
	suite.Setup()

	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[0])

	// create new denom
	res, err := suite.msgServer.CreateDenom(sdk.WrapSDKContext(suite.ChainA.GetContext()), types.NewMsgCreateDenom(senderAddress.String(), "testdenom"))
	suite.Require().NoError(err)
	factoryDenom := res.GetNewTokenDenom()

	tokenToSend := sdk.NewCoin(factoryDenom, sdk.NewInt(100))
	halfTokenToSend := sdk.NewCoin(factoryDenom, sdk.NewInt(50))
	// mint tokens
	_, err = suite.msgServer.Mint(sdk.WrapSDKContext(suite.ChainA.GetContext()), types.NewMsgMint(senderAddress.String(), tokenToSend))
	suite.Require().NoError(err)

	// Send some tokens to the dex
	err = suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.SendCoinsFromAccountToModule(suite.ChainA.GetContext(), senderAddress, dextypes.ModuleName, sdk.NewCoins(halfTokenToSend))
	suite.Require().NoError(err)

	// set beforesend hook to the new denom with an invalid contract address
	invalidContractAddr := "neutron10h9stc5v6ntgeygf5xf945njqq5h32r54rf7kf"
	_, err = suite.msgServer.SetBeforeSendHook(sdk.WrapSDKContext(suite.ChainA.GetContext()), types.NewMsgSetBeforeSendHook(senderAddress.String(), factoryDenom, invalidContractAddr))
	suite.Require().NoError(err)

	// Trying to send more tokens to dex fails
	err = suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.SendCoinsFromAccountToModule(suite.ChainA.GetContext(), senderAddress, dextypes.ModuleName, sdk.NewCoins(halfTokenToSend))
	suite.Require().Error(err)

	// Sending tokens from dex (disablesBlockSend) to account succeeds
	err = suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.SendCoinsFromModuleToAccount(suite.ChainA.GetContext(), dextypes.ModuleName, senderAddress, sdk.NewCoins(halfTokenToSend))
	suite.Require().NoError(err)
}
