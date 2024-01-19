package v202_test

import (
	"testing"

	crontypes "github.com/neutron-org/neutron/v2/x/cron/types"
	feeburnertypes "github.com/neutron-org/neutron/v2/x/feeburner/types"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	v202 "github.com/neutron-org/neutron/v2/app/upgrades/v2.0.2"
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

	feeParams := feeburnertypes.NewParams(feeburnertypes.DefaultNeutronDenom, treasuryAddress)
	suite.Require().NoError(app.FeeBurnerKeeper.SetParams(ctx, feeParams))

	upgrade := upgradetypes.Plan{
		Name:   v202.UpgradeName,
		Info:   "ads",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	// get the auction module's params
	params, err := app.AuctionKeeper.GetParams(ctx)
	suite.Require().NoError(err)

	// check that the params are correct
	params.MaxBundleSize = v202.AuctionParamsMaxBundleSize
	params.ReserveFee = v202.AuctionParamsReserveFee
	params.MinBidIncrement = v202.AuctionParamsMinBidIncrement
	params.FrontRunningProtection = v202.AuctionParamsFrontRunningProtection
	params.ProposerFee = v202.AuctionParamsProposerFee

	addr, err := sdk.AccAddressFromBech32(treasuryAddress)
	suite.Require().NoError(err)

	suite.Require().Equal(addr.Bytes(), params.EscrowAccountAddress)
}
