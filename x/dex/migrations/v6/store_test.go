package v6_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v7/testutil"
	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	v6 "github.com/neutron-org/neutron/v7/x/dex/migrations/v6"
	"github.com/neutron-org/neutron/v7/x/dex/types"
)

type V6DexMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V6DexMigrationTestSuite))
}

func (suite *V6DexMigrationTestSuite) TestFieldUpdates() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)
	tk0 := &types.LimitOrderTrancheKey{TradePairId: types.MustNewTradePairID("TokenA", "TokenB"), TickIndexTakerToMaker: 1, TrancheKey: "123"}
	tranche0 := &types.LimitOrderTranche{
		Key:                tk0,
		ReservesMakerDenom: math.NewInt(1),
		ReservesTakerDenom: math.NewInt(2),
		TotalTakerDenom:    math.NewInt(3),
	}
	tk1 := &types.LimitOrderTrancheKey{TradePairId: types.MustNewTradePairID("TokenA", "TokenB"), TickIndexTakerToMaker: 2, TrancheKey: "123"}
	tranche1 := &types.LimitOrderTranche{
		Key:                tk1,
		ReservesMakerDenom: math.NewInt(4),
		ReservesTakerDenom: math.NewInt(5),
		TotalTakerDenom:    math.NewInt(6),
	}

	pk0 := &types.PoolReservesKey{TradePairId: types.MustNewTradePairID("TokenA", "TokenB"), TickIndexTakerToMaker: 1, Fee: 1}
	pool0 := &types.PoolReserves{
		Key:                pk0,
		ReservesMakerDenom: math.NewInt(10),
	}
	pk1 := &types.PoolReservesKey{TradePairId: types.MustNewTradePairID("TokenA", "TokenB"), TickIndexTakerToMaker: 2, Fee: 1}
	pool1 := &types.PoolReserves{
		Key:                pk1,
		ReservesMakerDenom: math.NewInt(10000),
	}

	app.DexKeeper.SetLimitOrderTranche(ctx, tranche0)
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche1)
	app.DexKeeper.SetInactiveLimitOrderTranche(ctx, tranche0)
	app.DexKeeper.SetInactiveLimitOrderTranche(ctx, tranche1)
	app.DexKeeper.SetPoolReserves(ctx, pool0)
	app.DexKeeper.SetPoolReserves(ctx, pool1)

	suite.NoError(v6.MigrateStore(ctx, cdc, storeKey))

	migratedTranche0 := app.DexKeeper.GetLimitOrderTranche(ctx, tk0)
	migratedTranche1 := app.DexKeeper.GetLimitOrderTranche(ctx, tk1)
	inactiveMigratedTranche0, _ := app.DexKeeper.GetInactiveLimitOrderTranche(ctx, tk0)
	inactiveMigratedTranche1, _ := app.DexKeeper.GetInactiveLimitOrderTranche(ctx, tk1)
	migratedPool0, _ := app.DexKeeper.GetPoolReserves(ctx, pk0)
	migratedPool1, _ := app.DexKeeper.GetPoolReserves(ctx, pk1)

	suite.Equal(math_utils.MustNewPrecDecFromStr("1"), migratedTranche0.DecReservesMakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("2"), migratedTranche0.DecReservesTakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("3"), migratedTranche0.DecTotalTakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("4"), migratedTranche1.DecReservesMakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("5"), migratedTranche1.DecReservesTakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("6"), migratedTranche1.DecTotalTakerDenom)

	suite.Equal(math_utils.MustNewPrecDecFromStr("1"), inactiveMigratedTranche0.DecReservesMakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("2"), inactiveMigratedTranche0.DecReservesTakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("3"), inactiveMigratedTranche0.DecTotalTakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("4"), inactiveMigratedTranche1.DecReservesMakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("5"), inactiveMigratedTranche1.DecReservesTakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("6"), inactiveMigratedTranche1.DecTotalTakerDenom)

	suite.Equal(math_utils.MustNewPrecDecFromStr("10"), migratedPool0.DecReservesMakerDenom)
	suite.Equal(math_utils.MustNewPrecDecFromStr("10000"), migratedPool1.DecReservesMakerDenom)

}
