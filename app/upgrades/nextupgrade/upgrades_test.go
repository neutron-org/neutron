package nextupgrade_test

import (
	"testing"

	feeburnertypes "github.com/neutron-org/neutron/v3/x/feeburner/types"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/stretchr/testify/require"

	nextupgrade "github.com/neutron-org/neutron/v3/app/upgrades/nextupgrade"
	"github.com/neutron-org/neutron/v3/testutil"
)

var consAddr = sdk.ConsAddress("addr1_______________")

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

const treasuryAddress = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
	suite.Require().NoError(
		suite.GetNeutronZoneApp(suite.ChainA).FeeBurnerKeeper.SetParams(
			suite.ChainA.GetContext(), feeburnertypes.NewParams(feeburnertypes.DefaultNeutronDenom, treasuryAddress),
		))
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

func (suite *UpgradeTestSuite) TestSlashingUpgrade() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext()
	t := suite.T()
	params := slashingtypes.Params{SignedBlocksWindow: 100}

	unrealMissedBlocksCounter := int64(500)
	// store old signing info and bitmap entries
	info := slashingtypes.ValidatorSigningInfo{
		Address:             consAddr.String(),
		MissedBlocksCounter: unrealMissedBlocksCounter, // set unrealistic value of missed blocks
	}
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, info)

	for i := int64(0); i < params.SignedBlocksWindow; i++ {
		// all even blocks are missed
		require.NoError(t, app.SlashingKeeper.SetMissedBlockBitmapValue(ctx, consAddr, i, i%2 == 0))
	}

	upgrade := upgradetypes.Plan{
		Name:   nextupgrade.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	postUpgradeInfo, ok := app.SlashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
	require.True(t, ok)
	require.Equal(t, postUpgradeInfo.MissedBlocksCounter, int64(50))
}
