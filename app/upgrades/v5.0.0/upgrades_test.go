package v400_test

import (
	"testing"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	v500 "github.com/neutron-org/neutron/v4/app/upgrades/v5.0.0"
	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v4/testutil"
	dexkeeper "github.com/neutron-org/neutron/v4/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
}

func (suite *UpgradeTestSuite) TestUpgradeDexPause() {
	var (
		app       = suite.GetNeutronZoneApp(suite.ChainA)
		ctx       = suite.ChainA.GetContext()
		msgServer = dexkeeper.NewMsgServerImpl(app.DexKeeper)
	)

	params := app.DexKeeper.GetParams(ctx)

	suite.False(params.Paused)

	upgrade := upgradetypes.Plan{
		Name:   v500.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	params = app.DexKeeper.GetParams(ctx)

	suite.True(params.Paused)

	price := math_utils.OnePrecDec()
	_, err := msgServer.PlaceLimitOrder(ctx, &dextypes.MsgPlaceLimitOrder{
		Creator:        sample.AccAddress(),
		Receiver:       sample.AccAddress(),
		TokenIn:        "TokenA",
		TokenOut:       "TokenB",
		LimitSellPrice: &price,
		AmountIn:       math.OneInt(),
	})

	suite.ErrorIs(err, dextypes.ErrDexPaused)

}
