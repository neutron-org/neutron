package v400_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	v402 "github.com/neutron-org/neutron/v4/app/upgrades/v4.0.2"
	"github.com/neutron-org/neutron/v4/testutil"
	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	"github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/types"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
}

func (suite *UpgradeTestSuite) TestUnswappedTranche() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext().WithChainID("pion-1")
	)

	// Create tranche with empty TotalTakerDenom
	trancheKey := &dextypes.LimitOrderTrancheKey{
		TradePairId:           dextypes.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	tranche := &dextypes.LimitOrderTranche{
		Key:               trancheKey,
		PriceTakerToMaker: math.MustNewPrecDecFromStr("0.9950127279"),
		TotalMakerDenom:   sdkmath.NewInt(100),
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)

	// Create a couple trancheUsers for the tranche
	addr1 := sample.AccAddress()
	addr2 := sample.AccAddress()
	trancheUser1 := &dextypes.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr1,
		SharesOwned:           sdkmath.NewInt(25),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	trancheUser2 := &dextypes.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser1)
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser2)

	// Run upgrade
	upgrade := upgradetypes.Plan{
		Name:   v402.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	// Check LimitOrderTranche.TotalMakerDenom is still 0
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.TotalTakerDenom.IsZero())

	// Check tranche users are not deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr1, trancheKey.TrancheKey)
	suite.True(found)
	_, found = app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.True(found)
}

func (suite *UpgradeTestSuite) TestSwappedWithdrawnTranche() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext().WithChainID("pion-1")
	)

	// Create tranche with that has been swapped through
	trancheKey := &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	tranche := &types.LimitOrderTranche{
		Key:                trancheKey,
		PriceTakerToMaker:  math.MustNewPrecDecFromStr("0.9950127279"),
		TotalMakerDenom:    sdkmath.NewInt(100),
		TotalTakerDenom:    sdkmath.NewInt(80),
		ReservesTakerDenom: sdkmath.NewInt(0),
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)

	// Create a couple trancheUsers for the tranche; both users have withdrawn the swap profit
	addr1 := sample.AccAddress()
	addr2 := sample.AccAddress()
	trancheUser1 := &types.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr1,
		SharesOwned:           sdkmath.NewInt(25),
		SharesWithdrawn:       sdkmath.NewInt(20),
	}
	trancheUser2 := &types.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.NewInt(60),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser1)
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser2)

	// Run Upgrade
	upgrade := upgradetypes.Plan{
		Name:   v402.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	// Check LimitOrderTranche TotalMakerDenom is still 80
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.TotalTakerDenom.Equal(sdkmath.NewInt(80)))

	// Check tranche users are not deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr1, trancheKey.TrancheKey)
	suite.True(found)
	_, found = app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.True(found)
}

func (suite *UpgradeTestSuite) TestSwappedUnwithdrawnTranche() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext().WithChainID("pion-1")
	)

	// Create tranche that has been swapped through
	trancheKey := &dextypes.LimitOrderTrancheKey{
		TradePairId:           dextypes.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	tranche := &dextypes.LimitOrderTranche{
		Key:                trancheKey,
		PriceTakerToMaker:  math.ZeroPrecDec(),
		TotalMakerDenom:    sdkmath.NewInt(20),
		TotalTakerDenom:    sdkmath.NewInt(80),
		ReservesTakerDenom: sdkmath.NewInt(80),
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)

	// Create a couple trancheUsers for the tranche
	addr1 := sample.AccAddress()
	addr2 := sample.AccAddress()
	trancheUser1 := &dextypes.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr1,
		SharesOwned:           sdkmath.NewInt(25),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	trancheUser2 := &dextypes.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser1)
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser2)

	// Run Upgrade
	upgrade := upgradetypes.Plan{
		Name:   v402.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	// Check LimitOrderTranche.TotalMakerDenom is still 80
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.TotalTakerDenom.Equal(sdkmath.NewInt(80)))

	// Check tranche users are not deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr1, trancheKey.TrancheKey)
	suite.True(found)
	_, found = app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.True(found)

}

func (suite *UpgradeTestSuite) TestSwappedCanceledTranche() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext().WithChainID("pion-1")
	)

	// Create tranche with that has been swapped through, withdrawn and canceled
	// In this case the math was previously incorrect and TotalTakerDenom is too high
	trancheKey := &dextypes.LimitOrderTrancheKey{
		TradePairId:           dextypes.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	tranche := &dextypes.LimitOrderTranche{
		Key:                trancheKey,
		PriceTakerToMaker:  math.MustNewPrecDecFromStr("0.9950127279"),
		TotalMakerDenom:    sdkmath.NewInt(100),
		TotalTakerDenom:    sdkmath.NewInt(65),
		ReservesTakerDenom: sdkmath.NewInt(60),
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)

	// Create a single tranche user. Simulating that another user has already withdrawn and then canceled.

	addr2 := sample.AccAddress()
	trancheUser2 := &dextypes.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser2)

	// Run Upgrade
	upgrade := upgradetypes.Plan{
		Name:   v402.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	// Check LimitOrderTranche TotalMakerDenom is updated to 60
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.TotalTakerDenom.Equal(sdkmath.NewInt(60)))

	// Check tranche user is not deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.True(found)
}

func (suite *UpgradeTestSuite) TestOrphanedTrancheUser() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext().WithChainID("pion-1")
	)

	// Create two tranche users with no associated limitOrderTranche
	trancheKey := &dextypes.LimitOrderTrancheKey{
		TradePairId:           dextypes.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	addr1 := sample.AccAddress()
	addr2 := sample.AccAddress()
	trancheUser1 := &dextypes.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr1,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.OneInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser1)
	trancheUser2 := &dextypes.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.OneInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser2)

	// Run Upgrade
	upgrade := upgradetypes.Plan{
		Name:   v402.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	// Check tranche users are deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr1, trancheKey.TrancheKey)
	suite.False(found)
	_, found = app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.False(found)
}
