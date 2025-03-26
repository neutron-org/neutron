package v4_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/testutil"
	"github.com/neutron-org/neutron/v6/utils/math"
	v4 "github.com/neutron-org/neutron/v6/x/dex/migrations/v4"
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

	// Write tranche with incorrect price
	trancheKey := &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: -50,
		TrancheKey:            "123",
	}
	tranche := &types.LimitOrderTranche{
		Key:               trancheKey,
		PriceTakerToMaker: math.ZeroPrecDec(),
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)

	// also create inactive tranche
	app.DexKeeper.SetInactiveLimitOrderTranche(ctx, tranche)

	// Write poolReserves with incorrect prices
	poolKey := &types.PoolReservesKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: 60000,
		Fee:                   1,
	}
	poolReserves := &types.PoolReserves{
		Key:               poolKey,
		PriceTakerToMaker: math.ZeroPrecDec(),
	}
	app.DexKeeper.SetPoolReserves(ctx, poolReserves)

	// Run migration
	suite.NoError(v4.MigrateStore(ctx, cdc, storeKey))

	// Check LimitOrderTranche has correct price
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)

	suite.True(newTranche.PriceTakerToMaker.Equal(math.MustNewPrecDecFromStr("1.005012269623051203500693815")))

	// check InactiveLimitOrderTranche has correct price
	inactiveTranche, _ := app.DexKeeper.GetInactiveLimitOrderTranche(ctx, trancheKey)
	suite.True(inactiveTranche.PriceTakerToMaker.Equal(math.MustNewPrecDecFromStr("1.005012269623051203500693815")))

	// Check PoolReserves has the correct prices
	newPool, _ := app.DexKeeper.GetPoolReserves(ctx, poolKey)
	suite.True(newPool.PriceTakerToMaker.Equal(math.MustNewPrecDecFromStr("0.002479495864288162666675934")))
}
