package v7_test

import (
	"testing"

	"cosmossdk.io/math"
	v7 "github.com/neutron-org/neutron/v10/x/dex/migrations/v7"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v10/testutil"
	math_utils "github.com/neutron-org/neutron/v10/utils/math"
	"github.com/neutron-org/neutron/v10/x/dex/types"
)

type V7DexMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V7DexMigrationTestSuite))
}

func (suite *V7DexMigrationTestSuite) TestFieldUpdates() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	trancheUser0 := &types.LimitOrderTrancheUser{
		TrancheKey:      "123",
		Address:         "alice",
		SharesOwned:     math.NewInt(100),
		SharesWithdrawn: math.ZeroInt(),
	}

	trancheUser1 := &types.LimitOrderTrancheUser{
		TrancheKey:      "123",
		Address:         "bob",
		SharesOwned:     math.NewInt(100),
		SharesWithdrawn: math.NewInt(10),
	}

	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser0)
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, trancheUser1)

	suite.NoError(v7.MigrateStore(ctx, cdc, storeKey))

	migratedTrancheUser0, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, trancheUser0.Address, trancheUser0.TrancheKey)
	suite.True(found)
	migratedTrancheUser1, found := app.DexKeeper.GetLimitOrderTrancheUser(ctx, trancheUser1.Address, trancheUser1.TrancheKey)
	suite.True(found)

	suite.Equal(math.ZeroInt(), migratedTrancheUser0.SharesWithdrawn)
	suite.Equal(math_utils.ZeroPrecDec(), migratedTrancheUser0.DecSharesWithdrawn)
	suite.Equal(math.NewInt(10), migratedTrancheUser1.SharesWithdrawn)
	suite.Equal(math_utils.NewPrecDec(10), migratedTrancheUser1.DecSharesWithdrawn)
}
