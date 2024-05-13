package v400_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	comettypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/suite"

	v400 "github.com/neutron-org/neutron/v4/app/upgrades/v4.0.0"
	"github.com/neutron-org/neutron/v4/testutil"
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
	ctx := suite.ChainA.GetContext()
	t := suite.T()

	upgrade := upgradetypes.Plan{
		Name:   v400.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	params, err := app.MarketMapKeeper.GetParams(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(params.MarketAuthority, "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z")
	suite.Require().Equal(params.Version, uint64(0))
}

func (suite *UpgradeTestSuite) TestEnableVoteExtensionsUpgrade() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext()
	t := suite.T()

	oldParams, err := app.ConsensusParamsKeeper.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	// VoteExtensionsEnableHeight must be updated after the upgrade on upgrade height
	// but the rest of params must be the same
	oldParams.Params.Abci = &comettypes.ABCIParams{VoteExtensionsEnableHeight: ctx.BlockHeight() + 4}
	// it is automatically tracked in upgrade handler, we need to set it manually for tests
	oldParams.Params.Version = &comettypes.VersionParams{App: 0}

	upgrade := upgradetypes.Plan{
		Name:   v400.UpgradeName,
		Info:   "some text here",
		Height: ctx.BlockHeight(),
	}
	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	newParams, err := app.ConsensusParamsKeeper.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	suite.Require().Equal(oldParams, newParams)
}
