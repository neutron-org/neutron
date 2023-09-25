package nextupgrade_test

import (
	"github.com/neutron-org/neutron/app/upgrades/nextupgrade"
	"testing"

	adminmoduletypes "github.com/cosmos/admin-module/x/adminmodule/types"

	crontypes "github.com/neutron-org/neutron/x/cron/types"
	feeburnertypes "github.com/neutron-org/neutron/x/feeburner/types"
	feerefundertypes "github.com/neutron-org/neutron/x/feerefunder/types"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	tokenfactorytypes "github.com/neutron-org/neutron/x/tokenfactory/types"

	"github.com/neutron-org/neutron/app/params"

	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/testutil"
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
}

func (suite *UpgradeTestSuite) TestGlobalFeesUpgrade() {
	//ctrl := gomock.NewController(suite.T())
	//defer ctrl.Finish()

	var (
		app               = suite.GetNeutronZoneApp(suite.ChainA)
		globalFeeSubspace = app.GetSubspace(globalfee.ModuleName)
		ctx               = suite.ChainA.GetContext()
	)
	//app.WasmMsgServer = mock_upgrades.NewMockWasmMsgServer(ctrl)
	//feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))

	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMinGasPrices))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage))

	upgrade := upgradetypes.Plan{
		Name:   nextupgrade.UpgradeName,
		Info:   "testing_turn_off_contract_migrations",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMinGasPrices))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage))

	var globalMinGasPrices sdk.DecCoins
	globalFeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
	requiredGlobalFees := sdk.DecCoins{
		sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.MustNewDecFromStr("0.026")),
		sdk.NewDecCoinFromDec("ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349", sdk.MustNewDecFromStr("0.25")),
		sdk.NewDecCoinFromDec("untrn", sdk.MustNewDecFromStr("0.9")),
	}
	suite.Require().Equal(requiredGlobalFees, globalMinGasPrices)

	var actualBypassFeeMessages []string
	globalFeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes, &actualBypassFeeMessages)
	requiredBypassMinFeeMsgTypes := []string{
		sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
	}
	suite.Require().Equal(requiredBypassMinFeeMsgTypes, actualBypassFeeMessages)

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
		Name:   nextupgrade.UpgradeName,
		Info:   "testing_turn_off_contract_migrations",
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

	//emulate lack of ProposalIDKey like on a real mainnet
	store := ctx.KVStore(app.GetKey(adminmoduletypes.StoreKey))
	store.Delete(adminmoduletypes.ProposalIDKey)

	_, err := app.AdminmoduleKeeper.GetProposalID(ctx)
	suite.Require().Error(err)

	upgrade := upgradetypes.Plan{
		Name:   nextupgrade.UpgradeName,
		Info:   "testing_turn_off_contract_migrations",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	id, err := app.AdminmoduleKeeper.GetProposalID(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), id)
}
