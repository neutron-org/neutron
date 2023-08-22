package nextupgrade_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/neutron-org/neutron/app/upgrades/nextupgrade"
	"github.com/neutron-org/neutron/testutil"
	"github.com/stretchr/testify/suite"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
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
		Name:   nextupgrade.UpgradeName,
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
