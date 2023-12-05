package v2_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v2/testutil/apptesting"
	math_utils "github.com/neutron-org/neutron/v2/utils/math"
	v3 "github.com/neutron-org/neutron/v2/x/dex/migrations/v3"
	"github.com/neutron-org/neutron/v2/x/dex/types"
	typesv2 "github.com/neutron-org/neutron/v2/x/dex/types/v2"
	"github.com/neutron-org/neutron/v2/x/dex/utils"
)

type DexMigrationTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestDexMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(DexMigrationTestSuite))
}

func (suite *DexMigrationTestSuite) SetupTest() {
	suite.Setup()
}

var defaultTradePair = &types.TradePairID{
	MakerDenom: "TokenA",
	TakerDenom: "TokenB",
}

func newLOTickLiquidityV2(i uint64, time *time.Time) typesv2.TickLiquidity {
	newLO := newLimitOrderV2(i, time)

	return typesv2.TickLiquidity{
		Liquidity: &typesv2.TickLiquidity_LimitOrderTranche{
			LimitOrderTranche: &newLO,
		},
	}
}

func newLimitOrderV2(i uint64, time *time.Time) typesv2.LimitOrderTranche {
	return typesv2.LimitOrderTranche{
		Key: &typesv2.LimitOrderTrancheKey{
			TradePairId:           defaultTradePair,
			TickIndexTakerToMaker: 1,
			TrancheKey:            fmt.Sprintf("%dXXX", i),
		},
		ReservesMakerDenom: math.OneInt(),
		ReservesTakerDenom: math.ZeroInt(),
		TotalMakerDenom:    math.OneInt(),
		TotalTakerDenom:    math.ZeroInt(),
		ExpirationTime:     time,
		PriceTakerToMaker:  math_utils.OnePrecDec(),
	}
}

func newLimitOrderExpirationV2(i uint64, time time.Time) typesv2.LimitOrderExpiration {
	return typesv2.LimitOrderExpiration{
		ExpirationTime: time,
		TrancheRef:     []byte(fmt.Sprintf("%dXXX", i)),
	}
}

func v2LimitOrderExpirationKey(
	goodTilDate time.Time,
	trancheRef []byte,
) []byte {
	unixMs := uint64(goodTilDate.UnixMilli())
	str := utils.Uint64ToSortableString(unixMs)
	goodTilDateBytes := []byte(str)

	var key []byte

	key = append(key, goodTilDateBytes...)
	key = append(key, []byte("/")...)

	key = append(key, trancheRef...)
	key = append(key, []byte("/")...)

	return key
}

func (suite *DexMigrationTestSuite) assertV2TrancheEqualsV3(v2Tranche typesv2.LimitOrderTranche, v3Tranche types.LimitOrderTranche) {
	suite.Equal(v2Tranche.Key.TradePairId, v3Tranche.Key.TradePairId)
	suite.Equal(v2Tranche.Key.TickIndexTakerToMaker, v3Tranche.Key.TickIndexTakerToMaker)
	suite.Equal(v2Tranche.Key.TrancheKey, v3Tranche.Key.TrancheKey)
	suite.Equal(v2Tranche.ReservesMakerDenom, v3Tranche.ReservesMakerDenom)
	suite.Equal(v2Tranche.ReservesTakerDenom, v3Tranche.ReservesTakerDenom)
	suite.Equal(v2Tranche.TotalMakerDenom, v3Tranche.TotalMakerDenom)
	suite.Equal(v2Tranche.TotalTakerDenom, v3Tranche.TotalTakerDenom)
	suite.Equal(v2Tranche.PriceTakerToMaker, v3Tranche.PriceTakerToMaker)

	switch v2Tranche.ExpirationTime {
	case &time.Time{}:
		suite.Equal(int64(0), v3Tranche.ExpirationTime)
		suite.Equal(types.LimitOrderType_JUST_IN_TIME, v3Tranche.OrderType)
	case nil:
		suite.Equal(int64(0), v3Tranche.ExpirationTime)
		suite.Equal(types.LimitOrderType_GOOD_TIL_CANCELLED, v3Tranche.OrderType)

	default:
		suite.Equal(v2Tranche.ExpirationTime.Unix(), v3Tranche.ExpirationTime)
		suite.Equal(types.LimitOrderType_GOOD_TIL_TIME, v3Tranche.OrderType)
	}
}

func (suite *DexMigrationTestSuite) assertV2ExpirationEqualsV3(v2Expiration typesv2.LimitOrderExpiration, v3Expiration types.LimitOrderExpiration) {
	suite.Equal(v2Expiration.TrancheRef, v3Expiration.TrancheRef)

	switch v2Expiration.ExpirationTime {
	case time.Time{}:
		suite.Equal(int64(0), v3Expiration.ExpirationTime)
	default:
		suite.Equal(v2Expiration.ExpirationTime.Unix(), v3Expiration.ExpirationTime)
	}
}

func marshalV2TrancheKey(k typesv2.LimitOrderTrancheKey) []byte {
	var key []byte

	pairKeyBytes := []byte(k.TradePairId.MustPairID().CanonicalString())
	key = append(key, pairKeyBytes...)
	key = append(key, []byte("/")...)

	makerDenomBytes := []byte(k.TradePairId.MakerDenom)
	key = append(key, makerDenomBytes...)
	key = append(key, []byte("/")...)

	tickIndexBytes := types.Int64ToSortableBytes(k.TickIndexTakerToMaker)
	key = append(key, tickIndexBytes...)
	key = append(key, []byte("/")...)

	liquidityTypeBytes := []byte(types.LiquidityTypeLimitOrder)
	key = append(key, liquidityTypeBytes...)
	key = append(key, []byte("/")...)

	key = append(key, []byte(k.TrancheKey)...)
	key = append(key, []byte("/")...)

	return key
}

func (suite *DexMigrationTestSuite) TestTrancheUpgrade() {
	var (
		keeper   = suite.App.DexKeeper
		storeKey = suite.App.GetKey(types.StoreKey)
		cdc      = suite.App.AppCodec()
	)

	expirationTime1, _ := time.Parse("2006-Jan-02", "2023-Dec-02")
	expirationTime2, _ := time.Parse("2006-Jan-02", "2023-Dec-03")
	// Write old state
	// LOTicks
	store := prefix.NewStore(suite.Ctx.KVStore(storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	oldTicks := []typesv2.TickLiquidity{
		newLOTickLiquidityV2(0, &expirationTime1),
		newLOTickLiquidityV2(1, &expirationTime2),
		newLOTickLiquidityV2(2, &time.Time{}),
		newLOTickLiquidityV2(3, nil),
	}

	for _, v := range oldTicks {
		//nolint:gosec
		bz := cdc.MustMarshal(&v)
		key := marshalV2TrancheKey(*v.GetLimitOrderTranche().Key)
		store.Set(key, bz)
	}

	// InactiveLimitOrders
	store = prefix.NewStore(suite.Ctx.KVStore(storeKey), types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix))
	oldInactiveLOs := []typesv2.LimitOrderTranche{
		newLimitOrderV2(0, &expirationTime1),
		newLimitOrderV2(1, &time.Time{}),
		newLimitOrderV2(2, nil),
	}

	for _, v := range oldInactiveLOs {
		//nolint:gosec
		bz := cdc.MustMarshal(&v)
		key := marshalV2TrancheKey(*v.Key)
		store.Set(key, bz)
	}

	// LimitOrderExpirations
	store = prefix.NewStore(suite.Ctx.KVStore(storeKey), types.KeyPrefix(types.LimitOrderExpirationKeyPrefix))
	oldLOExpirations := []typesv2.LimitOrderExpiration{
		newLimitOrderExpirationV2(0, time.Time{}),
		newLimitOrderExpirationV2(1, expirationTime1),
		newLimitOrderExpirationV2(2, expirationTime2),
	}

	for _, v := range oldLOExpirations {
		//nolint:gosec
		bz := cdc.MustMarshal(&v)
		key := v2LimitOrderExpirationKey(v.ExpirationTime, v.TrancheRef)
		store.Set(key, bz)
	}

	// Run migration
	suite.NoError(v3.MigrateStore(suite.Ctx, cdc, storeKey))

	// check that LimitOrders have been successfully migrated
	newLOs := keeper.GetAllLimitOrderTrancheAtIndex(suite.Ctx, defaultTradePair, 0)
	for i, newLO := range newLOs {
		suite.assertV2TrancheEqualsV3(*oldTicks[i].GetLimitOrderTranche(), newLO)
	}

	// check that InactiveLimitsOrders have been successfully migrated
	newInactiveLOS := keeper.GetAllInactiveLimitOrderTranche(suite.Ctx)
	for i, newInactiveLO := range newInactiveLOS {
		suite.assertV2TrancheEqualsV3(oldInactiveLOs[i], *newInactiveLO)
	}

	// check that LimitOrderExpirations have been successfully migrated
	newLOExpirations := keeper.GetAllLimitOrderExpiration(suite.Ctx)
	for i, newLOExpiration := range newLOExpirations {
		suite.assertV2ExpirationEqualsV3(oldLOExpirations[i], *newLOExpiration)
	}
}
