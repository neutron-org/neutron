package v2_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v4/testutil"
	v2 "github.com/neutron-org/neutron/v4/x/tokenfactory/migrations/v2"
	"github.com/neutron-org/neutron/v4/x/tokenfactory/types"
	"github.com/neutron-org/neutron/v4/x/tokenfactory/types/v1beta1"
	"github.com/stretchr/testify/suite"
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
	oldParams := v1beta1.Params{
		DenomCreationFee:        sdk.NewCoins(sdk.NewCoin("untrn", math.OneInt())),
		DenomCreationGasConsume: types.DefaultDenomCreationGasConsume,
		FeeCollectorAddress:     "test_addr",
	}
	store := ctx.KVStore(storeKey)
	bz, err := cdc.Marshal(&oldParams)
	suite.Require().NoError(err)

	store.Set(types.ParamsKey, bz)

	// Run migration
	suite.NoError(v2.MigrateStore(ctx, cdc, storeKey))

	// Check params are correct
	newParams := app.TokenFactoryKeeper.GetParams(ctx)
	suite.Require().EqualValues(oldParams.DenomCreationFee, newParams.DenomCreationFee)
	suite.Require().EqualValues(newParams.DenomCreationGasConsume, newParams.DenomCreationGasConsume)
	suite.Require().EqualValues(newParams.FeeCollectorAddress, newParams.FeeCollectorAddress)
	suite.Require().EqualValues(newParams.WhitelistedHooks, v2.WhitelistedHooks)
}
