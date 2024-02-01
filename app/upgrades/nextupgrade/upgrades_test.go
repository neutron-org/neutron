package nextupgrade_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v2/app/upgrades/nextupgrade"
	"github.com/neutron-org/neutron/v2/testutil"
)

var consAddr = sdk.ConsAddress("addr1_______________")

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
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
