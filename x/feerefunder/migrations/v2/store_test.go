package v2_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v7/testutil"
	v2 "github.com/neutron-org/neutron/v7/x/feerefunder/migrations/v2"
	"github.com/neutron-org/neutron/v7/x/feerefunder/types"
)

type V2FeerefunderTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V2FeerefunderTestSuite))
}

func (suite *V2FeerefunderTestSuite) TestParamsUpdate() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// Setup before state
	params := app.FeeKeeper.GetParams(ctx)
	params.FeeEnabled = false
	suite.NoError(app.FeeKeeper.SetParams(ctx, params))

	// Before upgrade
	params = app.FeeKeeper.GetParams(ctx)
	suite.Require().False(params.FeeEnabled)

	// Run migration
	suite.NoError(v2.MigrateStore(ctx, cdc, storeKey))

	// After upgrade
	params = app.FeeKeeper.GetParams(ctx)
	suite.Require().True(params.FeeEnabled)
}
