package nextupgrade_test

import (
	"testing"

	crontypes "github.com/neutron-org/neutron/v2/x/cron/types"
	feeburnertypes "github.com/neutron-org/neutron/v2/x/feeburner/types"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	nextupgrade "github.com/neutron-org/neutron/v2/app/upgrades/nextupgrade"
	"github.com/neutron-org/neutron/v2/testutil"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

const treasuryAddress = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext()
	subspace, _ := app.ParamsKeeper.GetSubspace(crontypes.StoreKey)
	pcron := crontypes.DefaultParams()
	subspace.SetParamSet(ctx, &pcron)
}

func (suite *UpgradeTestSuite) TestAuctionUpgrade() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext()
	)
	upgrade := upgradetypes.Plan{
		Name:   nextupgrade.UpgradeName,
		Info:   "nextupgrade",
		Height: 100,
	}

	feeParams := feeburnertypes.NewParams(feeburnertypes.DefaultNeutronDenom, treasuryAddress)
	suite.Require().NoError(app.FeeBurnerKeeper.SetParams(ctx, feeParams))

	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	// get the auction module's params
	params, err := app.AuctionKeeper.GetParams(ctx)
	suite.Require().NoError(err)

	// check that the params are correct
	params.MaxBundleSize = nextupgrade.AuctionParamsMaxBundleSize
	params.ReserveFee = nextupgrade.AuctionParamsReserveFee
	params.MinBidIncrement = nextupgrade.AuctionParamsMinBidIncrement
	params.FrontRunningProtection = nextupgrade.AuctionParamsFrontRunningProtection
	params.ProposerFee = nextupgrade.AuctionParamsProposerFee

	addr, err := sdk.AccAddressFromBech32(treasuryAddress)
	suite.Require().NoError(err)

	suite.Require().Equal(addr.Bytes(), params.EscrowAccountAddress)
}
