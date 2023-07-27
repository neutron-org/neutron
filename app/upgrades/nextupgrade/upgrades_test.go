package nextupgrade_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
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

	upgrade := upgradetypes.Plan{
		Name:   nextupgrade.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMinGasPrices))

	var globalMinGasPrices sdk.DecCoins
	globalFeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, &globalMinGasPrices)

	requiredGlobalFees := sdk.DecCoins{
		sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.MustNewDecFromStr("0.026")),
		sdk.NewDecCoinFromDec("ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349", sdk.MustNewDecFromStr("0.25")),
		sdk.NewDecCoinFromDec("untrn", sdk.MustNewDecFromStr("0.9")),
	}
	suite.Require().Equal(requiredGlobalFees, globalMinGasPrices)
}
