package v500_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v500 "github.com/neutron-org/neutron/v5/app/upgrades/v5.0.0"
	"github.com/neutron-org/neutron/v5/testutil/common/sample"

	"github.com/neutron-org/neutron/v5/testutil"
	dexkeeper "github.com/neutron-org/neutron/v5/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v5/x/dex/types"
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

func (suite *UpgradeTestSuite) TestOracleUpgrade() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	upgrade := upgradetypes.Plan{
		Name:   v500.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	params, err := app.MarketMapKeeper.GetParams(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(params.MarketAuthorities[0], "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z")
	suite.Require().Equal(params.MarketAuthorities[1], v500.MarketMapAuthorityMultisig)
	suite.Require().Equal(params.Admin, "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z")
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

func (suite *UpgradeTestSuite) TestUpgradeSetRateLimitContractMainnet() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext().WithChainID("neutron-1")
	)

	params := app.RateLimitingICS4Wrapper.IbcratelimitKeeper.GetParams(ctx)

	suite.Equal(params.ContractAddress, "")

	upgrade := upgradetypes.Plan{
		Name:   v500.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	params = app.RateLimitingICS4Wrapper.IbcratelimitKeeper.GetParams(ctx)

	suite.Equal(params.ContractAddress, v500.MainnetRateLimitContract)
}

func (suite *UpgradeTestSuite) TestUpgradeSetRateLimitContractTestnet() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext().WithChainID("pion-1")
	)

	params := app.RateLimitingICS4Wrapper.IbcratelimitKeeper.GetParams(ctx)

	suite.Equal(params.ContractAddress, "")

	upgrade := upgradetypes.Plan{
		Name:   v500.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	params = app.RateLimitingICS4Wrapper.IbcratelimitKeeper.GetParams(ctx)

	suite.Equal(params.ContractAddress, v500.TestnetRateLimitContract)
}

func (suite *UpgradeTestSuite) TestUpgradeSetRateLimitContractUnknownChain() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext().WithChainID("unknown-chain")
	)

	params := app.RateLimitingICS4Wrapper.IbcratelimitKeeper.GetParams(ctx)

	suite.Equal(params.ContractAddress, "")

	upgrade := upgradetypes.Plan{
		Name:   v500.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.EqualError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade), fmt.Sprintf("unknown chain id %s", ctx.ChainID()))
}
