package v5_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/testutil"
	"github.com/neutron-org/neutron/v6/utils/math"
	v5 "github.com/neutron-org/neutron/v6/x/dex/migrations/v5"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

type V4DexMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V4DexMigrationTestSuite))
}

func (suite *V4DexMigrationTestSuite) TestPriceUpdates() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// Write tranches
	trancheKey := &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -1150,
		TrancheKey:            "123",
	}
	tranche := &types.LimitOrderTranche{
		Key:               trancheKey,
		PriceTakerToMaker: math.ZeroPrecDec(),
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)

	// also create inactive tranche
	app.DexKeeper.SetInactiveLimitOrderTranche(ctx, tranche)

	// Write poolReserves
	poolKey := &types.PoolReservesKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: 256000,
		Fee:                   1,
	}
	poolReserves := &types.PoolReserves{
		Key: poolKey,
	}
	app.DexKeeper.SetPoolReserves(ctx, poolReserves)

	// Run migration
	suite.NoError(v5.MigrateStore(ctx, cdc, storeKey))

	// Check LimitOrderTranche has correct price
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.MakerPrice.Equal(math.MustNewPrecDecFromStr("0.891371268935227562508365227")))

	// check InactiveLimitOrderTranche has correct price
	inactiveTranche, _ := app.DexKeeper.GetInactiveLimitOrderTranche(ctx, trancheKey)
	suite.True(inactiveTranche.MakerPrice.Equal(math.MustNewPrecDecFromStr("0.891371268935227562508365227")))

	// Check PoolReserves has the correct prices
	newPool, _ := app.DexKeeper.GetPoolReserves(ctx, poolKey)
	suite.True(newPool.MakerPrice.Equal(math.MustNewPrecDecFromStr("131033661522.558812694915985539856164620")))
}
