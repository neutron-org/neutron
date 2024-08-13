package v5_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v4/testutil"
	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	"github.com/neutron-org/neutron/v4/utils/math"
	v5 "github.com/neutron-org/neutron/v4/x/dex/migrations/v5"
	"github.com/neutron-org/neutron/v4/x/dex/types"
)

type V5DexMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V5DexMigrationTestSuite))
}

func (suite *V5DexMigrationTestSuite) TestUnswappedTranche() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// Create tranche with empty TotalTakerDenom
	trancheKey := &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	tranche := &types.LimitOrderTranche{
		Key:               trancheKey,
		PriceTakerToMaker: math.ZeroPrecDec(),
		TotalMakerDenom:   sdkmath.NewInt(100),
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)

	// Create a couple trancheUsers for the tranche
	addr1 := sample.AccAddress()
	addr2 := sample.AccAddress()
	trancheUser1 := &types.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr1,
		SharesOwned:           sdkmath.NewInt(25),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	trancheUser2 := &types.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser1)
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser2)

	// Run migration
	suite.NoError(v5.MigrateStore(ctx, cdc, storeKey))

	// Check LimitOrderTranche.TotalMakerDenom is still 0
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.TotalTakerDenom.IsZero())

	// Check tranche users are not deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr1, trancheKey.TrancheKey)
	suite.True(found)
	_, found = app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.True(found)

}

func (suite *V5DexMigrationTestSuite) TestSwappedUnwithdrawnTranche() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// Create tranche that has been swapped through
	trancheKey := &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	tranche := &types.LimitOrderTranche{
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
	trancheUser1 := &types.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr1,
		SharesOwned:           sdkmath.NewInt(25),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	trancheUser2 := &types.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser1)
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser2)

	// Run migration
	suite.NoError(v5.MigrateStore(ctx, cdc, storeKey))

	// Check LimitOrderTranche.TotalMakerDenom is still 80
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.TotalTakerDenom.Equal(sdkmath.NewInt(80)))

	// Check tranche users are not deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr1, trancheKey.TrancheKey)
	suite.True(found)
	_, found = app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.True(found)

}

func (suite *V5DexMigrationTestSuite) TestSwappedWithdrawnTranche() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// Create tranche with that has been swapped through
	trancheKey := &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	tranche := &types.LimitOrderTranche{
		Key:                trancheKey,
		PriceTakerToMaker:  math.ZeroPrecDec(),
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

	// Run migration
	suite.NoError(v5.MigrateStore(ctx, cdc, storeKey))

	// Check LimitOrderTranche TotalMakerDenom is still 80
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.TotalTakerDenom.Equal(sdkmath.NewInt(80)))

	// Check tranche users are not deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr1, trancheKey.TrancheKey)
	suite.True(found)
	_, found = app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.True(found)
}

func (suite *V5DexMigrationTestSuite) TestSwappedCanceledTranche() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// Create tranche with that has been swapped through, withdrawn and canceled
	// In this case the math was previously incorrect and TotalTakerDenom is too high
	trancheKey := &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	tranche := &types.LimitOrderTranche{
		Key:                trancheKey,
		PriceTakerToMaker:  math.ZeroPrecDec(),
		TotalMakerDenom:    sdkmath.NewInt(100),
		TotalTakerDenom:    sdkmath.NewInt(65),
		ReservesTakerDenom: sdkmath.NewInt(60),
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)

	// Create a single tranche user. Simulating that another user has already withdrawn and then canceled.

	addr2 := sample.AccAddress()
	trancheUser2 := &types.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser2)

	// Run migration
	suite.NoError(v5.MigrateStore(ctx, cdc, storeKey))

	// Check LimitOrderTranche TotalMakerDenom is updated to 60
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.TotalTakerDenom.Equal(sdkmath.NewInt(60)))

	// Check tranche user is not deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.True(found)
}

func (suite *V5DexMigrationTestSuite) TestOrphanedTrancheUser() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// Create two tranche users with no associated limitOrderTranche
	trancheKey := &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	addr1 := sample.AccAddress()
	addr2 := sample.AccAddress()
	trancheUser1 := &types.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser1)
	trancheUser2 := &types.LimitOrderTrancheUser{
		TradePairId:           trancheKey.TradePairId,
		TickIndexTakerToMaker: trancheKey.TickIndexTakerToMaker,
		TrancheKey:            trancheKey.TrancheKey,
		Address:               addr2,
		SharesOwned:           sdkmath.NewInt(75),
		SharesWithdrawn:       sdkmath.ZeroInt(),
	}
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser2)

	// Run migration
	suite.NoError(v5.MigrateStore(ctx, cdc, storeKey))

	// Check tranche users are deleted
	_, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr1, trancheKey.TrancheKey)
	suite.False(found)
	_, found = app.DexKeeper.GetLimitOrderTrancheUser(ctx, addr2, trancheKey.TrancheKey)
	suite.False(found)
}
