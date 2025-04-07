package v3_test

import (
	"testing"

	v3 "github.com/neutron-org/neutron/v6/x/tokenfactory/migrations/v3"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/testutil"
	v2 "github.com/neutron-org/neutron/v6/x/tokenfactory/migrations/v2"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

type V3TokenfactoryMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V3TokenfactoryMigrationTestSuite))
}

func (suite *V3TokenfactoryMigrationTestSuite) TestParamsUpgrade() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// Write old state
	oldParams := v3.Params{
		DenomCreationFee:        sdk.NewCoins(sdk.NewCoin("untrn", math.OneInt())),
		DenomCreationGasConsume: types.DefaultDenomCreationGasConsume,
		FeeCollectorAddress:     "test_addr",
		WhitelistedHooks:        v2.WhitelistedHooks,
	}
	store := ctx.KVStore(storeKey)
	bz, err := cdc.Marshal(&oldParams)
	suite.Require().NoError(err)

	store.Set(types.ParamsKey, bz)

	// Run migration
	suite.NoError(v3.MigrateStore(ctx, cdc, storeKey))

	// Check params are correct
	newParams := app.TokenFactoryKeeper.GetParams(ctx)
	suite.Require().EqualValues(newParams.DenomCreationFee, oldParams.DenomCreationFee)
	suite.Require().EqualValues(newParams.DenomCreationGasConsume, oldParams.DenomCreationGasConsume)
	suite.Require().EqualValues(newParams.FeeCollectorAddress, oldParams.FeeCollectorAddress)
	suite.Require().EqualValues(newParams.WhitelistedHooks, oldParams.WhitelistedHooks)
	suite.Require().EqualValues(newParams.TrackBeforeSendGasLimit, types.DefaultTrackBeforeSendGasLimit)
}
