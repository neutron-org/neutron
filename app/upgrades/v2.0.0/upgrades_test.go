package v200_test

import (
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	adminmoduletypes "github.com/cosmos/admin-module/x/adminmodule/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	crontypes "github.com/neutron-org/neutron/v2/x/cron/types"
	feeburnertypes "github.com/neutron-org/neutron/v2/x/feeburner/types"
	feerefundertypes "github.com/neutron-org/neutron/v2/x/feerefunder/types"
	icqtypes "github.com/neutron-org/neutron/v2/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v2/x/interchaintxs/types"
	tokenfactorytypes "github.com/neutron-org/neutron/v2/x/tokenfactory/types"

	"github.com/neutron-org/neutron/v2/app/params"

	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	"github.com/stretchr/testify/suite"

	v200 "github.com/neutron-org/neutron/v2/app/upgrades/v2.0.0"
	"github.com/neutron-org/neutron/v2/testutil"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext()
	subspace, _ := app.ParamsKeeper.GetSubspace(crontypes.StoreKey)
	pcron := crontypes.DefaultParams()
	subspace.SetParamSet(ctx, &pcron)

	subspace, _ = app.ParamsKeeper.GetSubspace(feeburnertypes.StoreKey)
	p := feeburnertypes.NewParams(feeburnertypes.DefaultNeutronDenom, "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh")
	subspace.SetParamSet(ctx, &p)

	subspace, _ = app.ParamsKeeper.GetSubspace(feerefundertypes.StoreKey)
	pFeeRefunder := feerefundertypes.DefaultParams()
	subspace.SetParamSet(ctx, &pFeeRefunder)

	subspace, _ = app.ParamsKeeper.GetSubspace(tokenfactorytypes.StoreKey)
	pTokenfactory := tokenfactorytypes.DefaultParams()
	subspace.SetParamSet(ctx, &pTokenfactory)

	subspace, _ = app.ParamsKeeper.GetSubspace(icqtypes.StoreKey)
	pICQTypes := icqtypes.DefaultParams()
	subspace.SetParamSet(ctx, &pICQTypes)

	subspace, _ = app.ParamsKeeper.GetSubspace(interchaintxstypes.StoreKey)
	pICAtx := interchaintxstypes.DefaultParams()
	subspace.SetParamSet(ctx, &pICAtx)

	codeIDBefore := suite.StoreTestCode(ctx, sdk.AccAddress("neutron1weweewe"), "testdata/neutron_interchain_txs.wasm")
	suite.InstantiateTestContract(ctx, sdk.AccAddress("neutron1weweewe"), codeIDBefore)
}

func (suite *UpgradeTestSuite) TestGlobalFeesUpgrade() {
	var (
		app               = suite.GetNeutronZoneApp(suite.ChainA)
		globalFeeSubspace = app.GetSubspace(globalfee.ModuleName)
		ctx               = suite.ChainA.GetContext()
	)

	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMinGasPrices))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage))

	upgrade := upgradetypes.Plan{
		Name:   v200.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMinGasPrices))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage))

	var globalMinGasPrices sdk.DecCoins
	globalFeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
	requiredGlobalFees := sdk.DecCoins{
		sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.MustNewDecFromStr("0.02")),
		sdk.NewDecCoinFromDec("ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349", sdk.MustNewDecFromStr("0.2")),
		sdk.NewDecCoinFromDec("untrn", sdk.MustNewDecFromStr("0.56")),
	}
	suite.Require().Equal(requiredGlobalFees, globalMinGasPrices)

	var actualBypassFeeMessages []string
	globalFeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes, &actualBypassFeeMessages)
	suite.Require().Equal(0, len(actualBypassFeeMessages))

	var actualTotalBypassMinFeeMsgGasUsage uint64
	globalFeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, &actualTotalBypassMinFeeMsgGasUsage)
	requiredTotalBypassMinFeeMsgGasUsage := uint64(1_000_000)
	suite.Require().Equal(requiredTotalBypassMinFeeMsgGasUsage, actualTotalBypassMinFeeMsgGasUsage)
}

func (suite *UpgradeTestSuite) TestRewardDenomsUpgrade() {
	var (
		app                 = suite.GetNeutronZoneApp(suite.ChainA)
		ccvConsumerSubspace = app.GetSubspace(ccvconsumertypes.ModuleName)
		ctx                 = suite.ChainA.GetContext()
	)

	suite.Require().True(ccvConsumerSubspace.Has(ctx, ccvconsumertypes.KeyRewardDenoms))

	// emulate mainnet/testnet state
	ccvConsumerSubspace.Set(ctx, ccvconsumertypes.KeyRewardDenoms, &[]string{params.DefaultDenom})

	var denomsBefore []string
	ccvConsumerSubspace.Get(ctx, ccvconsumertypes.KeyRewardDenoms, &denomsBefore)
	suite.Require().Equal(denomsBefore, []string{params.DefaultDenom})

	upgrade := upgradetypes.Plan{
		Name:   v200.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	suite.Require().True(ccvConsumerSubspace.Has(ctx, ccvconsumertypes.KeyRewardDenoms))

	var denoms []string
	ccvConsumerSubspace.Get(ctx, ccvconsumertypes.KeyRewardDenoms, &denoms)
	requiredDenoms := []string{params.DefaultDenom, "ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349"}
	suite.Require().Equal(requiredDenoms, denoms)
}

func (suite *UpgradeTestSuite) TestAdminModuleUpgrade() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext()
	)

	// emulate lack of ProposalIDKey like on a real mainnet
	store := ctx.KVStore(app.GetKey(adminmoduletypes.StoreKey))
	store.Delete(adminmoduletypes.ProposalIDKey)

	_, err := app.AdminmoduleKeeper.GetProposalID(ctx)
	suite.Require().Error(err)

	upgrade := upgradetypes.Plan{
		Name:   v200.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	id, err := app.AdminmoduleKeeper.GetProposalID(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), id)
}

func (suite *UpgradeTestSuite) TestTokenFactoryUpgrade() {
	var (
		app                  = suite.GetNeutronZoneApp(suite.ChainA)
		tokenFactorySubspace = app.GetSubspace(tokenfactorytypes.ModuleName)
		ctx                  = suite.ChainA.GetContext()
	)

	var denomGasBefore uint64
	tokenFactorySubspace.Get(ctx, tokenfactorytypes.KeyDenomCreationGasConsume, &denomGasBefore)
	suite.Require().Equal(denomGasBefore, uint64(0))

	// emulate mainnet state
	tokenFactorySubspace.Set(ctx, tokenfactorytypes.KeyDenomCreationFee, sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdk.NewInt(100_000_000))))
	tokenFactorySubspace.Set(ctx, tokenfactorytypes.KeyFeeCollectorAddress, "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh")

	var denomFeeBefore sdk.Coins
	tokenFactorySubspace.Get(ctx, tokenfactorytypes.KeyDenomCreationFee, &denomFeeBefore)
	suite.Require().Equal(denomFeeBefore, sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdk.NewInt(100_000_000))))

	upgrade := upgradetypes.Plan{
		Name:   v200.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	var denomGas uint64
	tokenFactorySubspace.Get(ctx, tokenfactorytypes.KeyDenomCreationGasConsume, &denomGas)
	requiredGasDenom := uint64(0)
	suite.Require().Equal(requiredGasDenom, denomGas)

	var denomFee sdk.Coins
	tokenFactorySubspace.Get(ctx, tokenfactorytypes.KeyDenomCreationFee, &denomFee)
	requiredFeeDenom := sdk.NewCoins()
	suite.Require().Equal(len(requiredFeeDenom), len(denomFee))

	var feeCollector string
	tokenFactorySubspace.Get(ctx, tokenfactorytypes.KeyFeeCollectorAddress, &feeCollector)
	requiredFeeCollector := ""
	suite.Require().Equal(requiredFeeCollector, feeCollector)
}

func (suite *UpgradeTestSuite) TestRegisterInterchainAccountCreationFee() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext()
	)

	suite.FundAcc(sdk.AccAddress("neutron1weweewe"), sdk.NewCoins(sdk.NewCoin("untrn", sdk.NewInt(1_000_000))))
	contractKeeper := keeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	// store contract for register ica w/o fees
	codeIDBefore := suite.StoreTestCode(ctx, sdk.AccAddress("neutron1_ica"), "testdata/neutron_interchain_txs.wasm")
	contractAddressBeforeUpgrade := suite.InstantiateTestContract(ctx, sdk.AccAddress("neutron1_ica"), codeIDBefore)

	upgrade := upgradetypes.Plan{
		Name:   v200.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	lastCodeID := app.InterchainTxsKeeper.GetICARegistrationFeeFirstCodeID(ctx)
	// ensure that wasm module stores next code id
	suite.Require().Equal(lastCodeID, codeIDBefore+1)

	// store contract after upgrade
	codeID := suite.StoreTestCode(ctx, sdk.AccAddress("neutron1_ica"), "testdata/neutron_interchain_txs.wasm")
	contractAddressAfterUpgrade := suite.InstantiateTestContract(ctx, sdk.AccAddress("neutron1_ica"), codeID)
	// register w/o actual fees
	jsonStringBeforeUpgrade := `{"register": {"connection_id":"connection-1","interchain_account_id":"test-2"}}`
	byteEncodedMsgBeforeUpgrade := []byte(jsonStringBeforeUpgrade)
	_, err := contractKeeper.Execute(ctx, contractAddressBeforeUpgrade, sdk.AccAddress("neutron1_ica"), byteEncodedMsgBeforeUpgrade, nil)
	suite.Require().NoError(err)

	// register with fees
	jsonStringAfterUpgrade := `{"register": {"connection_id":"connection-1","interchain_account_id":"test-3"}}`
	byteEncodedMsgAfterUpgrade := []byte(jsonStringAfterUpgrade)
	_, err = contractKeeper.Execute(ctx, contractAddressAfterUpgrade, sdk.AccAddress("neutron1weweewe"), byteEncodedMsgAfterUpgrade, sdk.NewCoins(sdk.NewCoin("untrn", sdk.NewInt(1_000_000))))
	suite.Require().NoError(err)

	// failed register due lack of fees (fees required)
	_, err = contractKeeper.Execute(ctx, contractAddressAfterUpgrade, sdk.AccAddress("neutron1weweewe"), byteEncodedMsgAfterUpgrade, nil)
	suite.ErrorIs(err, errors.ErrInsufficientFunds)
}
