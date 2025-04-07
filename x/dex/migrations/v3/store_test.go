package v3_test

import (
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/store/prefix"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/testutil"
	"github.com/neutron-org/neutron/v6/utils/math"
	v3 "github.com/neutron-org/neutron/v6/x/dex/migrations/v3"
	"github.com/neutron-org/neutron/v6/x/dex/types"
	v2types "github.com/neutron-org/neutron/v6/x/dex/types/v2"
	"github.com/neutron-org/neutron/v6/x/dex/utils"
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
	unixMs := uint64(timestamp.UnixMilli()) //nolint:gosec
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
