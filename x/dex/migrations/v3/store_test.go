package v3_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v4/testutil"
	"github.com/neutron-org/neutron/v4/utils/math"
	v3 "github.com/neutron-org/neutron/v4/x/dex/migrations/v3"
	"github.com/neutron-org/neutron/v4/x/dex/types"

	v2types "github.com/neutron-org/neutron/v4/x/dex/types/v2"
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
}
