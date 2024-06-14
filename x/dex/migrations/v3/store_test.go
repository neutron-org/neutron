package v3_test

import (
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/store/prefix"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v4/testutil"
	"github.com/neutron-org/neutron/v4/utils/math"
	v3 "github.com/neutron-org/neutron/v4/x/dex/migrations/v3"
	"github.com/neutron-org/neutron/v4/x/dex/types"
	v2types "github.com/neutron-org/neutron/v4/x/dex/types/v2"
	"github.com/neutron-org/neutron/v4/x/dex/utils"
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
	oldParams := v2types.Params{
		FeeTiers:           []uint64{0, 1, 2, 3, 4, 5, 10, 20, 50, 100, 150, 200},
		MaxTrueTakerSpread: math.NewPrecDec(1),
	}
	store := ctx.KVStore(storeKey)
	bz, err := cdc.Marshal(&oldParams)
	suite.Require().NoError(err)

	store.Set(types.KeyPrefix(types.ParamsKey), bz)

	// Run migration
	suite.NoError(v3.MigrateStore(ctx, cdc, storeKey))

	// Check params are correct
	newParams := app.DexKeeper.GetParams(ctx)
	suite.Require().EqualValues(oldParams.FeeTiers, newParams.FeeTiers)
	suite.Require().EqualValues(newParams.Paused, types.DefaultPaused)
	suite.Require().EqualValues(newParams.GoodTilPurgeAllowance, types.DefaultGoodTilPurgeAllowance)
	suite.Require().EqualValues(newParams.MaxJitsPerBlock, types.DefaultMaxJITsPerBlock)
}

func v2TimeBytes(timestamp time.Time) []byte {
	unixMs := uint64(timestamp.UnixMilli())
	str := utils.Uint64ToSortableString(unixMs)
	return []byte(str)
}

func v2LimitOrderExpirationKey(
	goodTilDate time.Time,
	trancheRef []byte,
) []byte {
	var key []byte

	goodTilDateBytes := v2TimeBytes(goodTilDate)
	key = append(key, goodTilDateBytes...)
	key = append(key, []byte("/")...)

	key = append(key, trancheRef...)
	key = append(key, []byte("/")...)

	return key
}

func (suite *V3DexMigrationTestSuite) TestLimitOrderExpirationUpgrade() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)
	store := prefix.NewStore(
		ctx.KVStore(storeKey),
		types.KeyPrefix(types.LimitOrderExpirationKeyPrefix),
	)
	lOExpirations := make([]types.LimitOrderExpiration, 0)
	// Write old LimitOrderExpirations
	for i := 0; i < 10; i++ {
		expiration := types.LimitOrderExpiration{
			ExpirationTime: time.Now().AddDate(0, 0, i).UTC(),
			TrancheRef:     []byte(strconv.Itoa(i)),
		}
		lOExpirations = append(lOExpirations, expiration)
		b := cdc.MustMarshal(&expiration)
		store.Set(v2LimitOrderExpirationKey(
			expiration.ExpirationTime,
			expiration.TrancheRef,
		), b)

	}

	// Run migration
	suite.NoError(v3.MigrateStore(ctx, cdc, storeKey))

	// Check LimitOrderExpirations can be found with the new key
	for _, v := range lOExpirations {
		expiration, found := app.DexKeeper.GetLimitOrderExpiration(ctx, v.ExpirationTime, v.TrancheRef)
		suite.Require().True(found)
		suite.Require().EqualValues(v, *expiration)
	}

	// check that no extra LimitOrderExpirations exist
	allExp := app.DexKeeper.GetAllLimitOrderExpiration(ctx)
	suite.Require().Equal(len(lOExpirations), len(allExp))
}

func (suite *V3DexMigrationTestSuite) TestPriceUpdates() {
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
	// create active tranche
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)
	// create inactive tranche
	app.DexKeeper.SetInactiveLimitOrderTranche(ctx, tranche)

	// Write poolReserves with old precision
	poolKey := &types.PoolReservesKey{
		TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
		TickIndexTakerToMaker: 60000,
		Fee:                   1,
	}
	poolReserves := &types.PoolReserves{
		Key:                       poolKey,
		PriceTakerToMaker:         math.ZeroPrecDec(),
		PriceOppositeTakerToMaker: math.ZeroPrecDec(),
	}

	app.DexKeeper.SetPoolReserves(ctx, poolReserves)

	// Run migration
	suite.NoError(v3.MigrateStore(ctx, cdc, storeKey))

	// Check LimitOrderTranche has correct price
	newTranche := app.DexKeeper.GetLimitOrderTranche(ctx, trancheKey)
	suite.True(newTranche.PriceTakerToMaker.Equal(math.MustNewPrecDecFromStr("1.005012269623051203500693815")), "was : %v", newTranche.PriceTakerToMaker)

	// check InactiveLimitOrderTranche has correct price
	inactiveTranche, _ := app.DexKeeper.GetInactiveLimitOrderTranche(ctx, trancheKey)
	suite.True(inactiveTranche.PriceTakerToMaker.Equal(math.MustNewPrecDecFromStr("1.005012269623051203500693815")), "was : %v", newTranche.PriceTakerToMaker)

	// Check PoolReserves has the correct prices
	newPool, _ := app.DexKeeper.GetPoolReserves(ctx, poolKey)
	suite.True(newPool.PriceTakerToMaker.Equal(math.MustNewPrecDecFromStr("0.002479495864288162666675923")), "was : %v", newPool.PriceTakerToMaker)
	suite.True(newPool.PriceOppositeTakerToMaker.Equal(math.MustNewPrecDecFromStr("403.227141612124702272520931931")), "was : %v", newPool.PriceOppositeTakerToMaker)
}
