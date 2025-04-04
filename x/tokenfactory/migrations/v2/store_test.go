package v2_test

import (
	"encoding/json"
	"os"
	"testing"

	"cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/testutil"
	v2 "github.com/neutron-org/neutron/v6/x/tokenfactory/migrations/v2"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/types/v1beta1"
)

type V3DexMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V3DexMigrationTestSuite))
}

func (suite *V3DexMigrationTestSuite) TestParamsUpgrade() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// Write old state
	oldParams := v1beta1.Params{
		DenomCreationFee:        sdk.NewCoins(sdk.NewCoin("untrn", math.OneInt())),
		DenomCreationGasConsume: types.DefaultDenomCreationGasConsume,
		FeeCollectorAddress:     "test_addr",
	}
	store := ctx.KVStore(storeKey)
	bz, err := cdc.Marshal(&oldParams)
	suite.Require().NoError(err)

	store.Set(types.ParamsKey, bz)

	// Run migration
	suite.NoError(v2.MigrateStore(ctx, cdc, storeKey, app.TokenFactoryKeeper))

	// Check params are correct
	newParams := app.TokenFactoryKeeper.GetParams(ctx)
	suite.Require().EqualValues(oldParams.DenomCreationFee, newParams.DenomCreationFee)
	suite.Require().EqualValues(newParams.DenomCreationGasConsume, newParams.DenomCreationGasConsume)
	suite.Require().EqualValues(newParams.FeeCollectorAddress, newParams.FeeCollectorAddress)
	suite.Require().EqualValues(newParams.WhitelistedHooks, v2.WhitelistedHooks)
}

func (suite *V3DexMigrationTestSuite) TestHooksUpgrade() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
		addr1    = suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
		addr2    = suite.ChainA.SenderAccounts[1].SenderAccount.GetAddress()
	)

	wasmFile := "../../keeper/testdata/balance_tracker.wasm"
	wasmCode, err := os.ReadFile(wasmFile)
	suite.Require().NoError(err)

	// Setup contract
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper)
	codeID, _, err := contractKeeper.Create(ctx, addr1, wasmCode, nil)
	suite.Require().NoError(err)
	tokenFactoryModuleAddr := app.AccountKeeper.GetModuleAddress(types.ModuleName)
	initMsg, _ := json.Marshal(
		map[string]interface{}{
			"tracked_denom":               "test_denom",
			"tokenfactory_module_address": tokenFactoryModuleAddr,
		},
	)
	cwAddress, _, err := contractKeeper.Instantiate(ctx, codeID, addr1, addr1, initMsg, "", sdk.NewCoins())
	suite.Require().NoError(err)
	cwAddressStr := cwAddress.String()

	// Add Denoms and hooks
	factoryDenom1, err := app.TokenFactoryKeeper.CreateDenom(ctx, addr1.String(), "test1")
	suite.Require().NoError(err)
	store := app.TokenFactoryKeeper.GetDenomPrefixStore(ctx, factoryDenom1)
	store.Set([]byte(types.BeforeSendHookAddressPrefixKey), []byte(cwAddressStr))

	factoryDenom2, err := app.TokenFactoryKeeper.CreateDenom(ctx, addr2.String(), "test2")
	suite.Require().NoError(err)
	store = app.TokenFactoryKeeper.GetDenomPrefixStore(ctx, factoryDenom2)
	store.Set([]byte(types.BeforeSendHookAddressPrefixKey), []byte(cwAddressStr))

	factoryDenom3, err := app.TokenFactoryKeeper.CreateDenom(ctx, addr2.String(), "test3")
	suite.Require().NoError(err)
	store = app.TokenFactoryKeeper.GetDenomPrefixStore(ctx, factoryDenom2)
	store.Set([]byte(types.BeforeSendHookAddressPrefixKey), []byte(cwAddressStr))

	// Include the hook we want to whitelist in the params migration
	v2.WhitelistedHooks = []*types.WhitelistedHook{
		{
			CodeID:       codeID,
			DenomCreator: addr1.String(),
		},
	}

	// Run migration
	suite.NoError(v2.MigrateStore(ctx, cdc, storeKey, app.TokenFactoryKeeper))

	// The whitelisted hook is still there
	hook1 := app.TokenFactoryKeeper.GetBeforeSendHook(ctx, factoryDenom1)
	suite.Assert().Equal(cwAddressStr, hook1)

	// The non whitelisted hooks have been removed
	hook2 := app.TokenFactoryKeeper.GetBeforeSendHook(ctx, factoryDenom2)
	suite.Assert().Equal("", hook2)
	hook3 := app.TokenFactoryKeeper.GetBeforeSendHook(ctx, factoryDenom3)
	suite.Assert().Equal("", hook3)
}
