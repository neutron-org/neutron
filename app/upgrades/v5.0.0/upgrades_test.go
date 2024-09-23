package v400_test

import (
	"testing"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/suite"

	v500 "github.com/neutron-org/neutron/v4/app/upgrades/v5.0.0"
	"github.com/neutron-org/neutron/v4/testutil/common/sample"

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
		ctx       = suite.ChainA.GetContext().WithChainID("neutron-1")
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

	_, err := msgServer.Deposit(ctx, &dextypes.MsgDeposit{
		Creator:         sample.AccAddress(),
		Receiver:        sample.AccAddress(),
		TokenA:          "TokenA",
		TokenB:          "TokenB",
		TickIndexesAToB: []int64{1},
		Fees:            []uint64{1},
		AmountsA:        []math.Int{math.OneInt()},
		AmountsB:        []math.Int{math.ZeroInt()},
		Options:         []*dextypes.DepositOptions{{}},
	})

	suite.ErrorIs(err, dextypes.ErrDexPaused)
}
