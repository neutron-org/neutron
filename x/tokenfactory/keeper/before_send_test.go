package keeper_test

import (
	"encoding/json"
	"fmt"
	"os"

	"cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/tokenfactory/types"
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
			res, err := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "testdenom"))
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
			_, err = suite.msgServer.SetBeforeSendHook(suite.ChainA.GetContext(), types.NewMsgSetBeforeSendHook(suite.TestAccs[0].String(), factoryDenom, cosmwasmAddress.String()))
			suite.Require().NoError(err, "test: %v", tc.name)

			tokenToSend := sdk.NewCoin(factoryDenom, math.NewInt(100))

			// mint tokens
			_, err = suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), tokenToSend))
			suite.Require().NoError(err)

			queryResp, err := suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper.QuerySmart(suite.ChainA.GetContext(), cosmwasmAddress, []byte(`{"total_supply_at":{}}`))
			suite.Require().NoError(err)

			// Contract has not been called because it is not whitelisted
			suite.Require().Equal("\"0\"", string(queryResp))

			// Whitelist the hook
			params := types.DefaultParams()
			params.WhitelistedHooks = []*types.HookWhitelist{{DenomCreator: senderAddress.String(), CodeID: codeID}}
			err = suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper.SetParams(suite.ChainA.GetContext(), params)
			suite.Require().NoError(err)

			// mint tokens again
			_, err = suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), tokenToSend))
			suite.Require().NoError(err)

			queryResp, err = suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper.QuerySmart(suite.ChainA.GetContext(), cosmwasmAddress, []byte(`{"total_supply_at":{}}`))
			suite.Require().NoError(err)

			// Whitelisted contract has now been called
			suite.Require().Equal("\"100\"", string(queryResp))
		})
	}
}
