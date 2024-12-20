package v505_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v505 "github.com/neutron-org/neutron/v5/app/upgrades/v5.0.5"
	"github.com/neutron-org/neutron/v5/testutil"
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

func (suite *UpgradeTestSuite) TestUpgrade() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	upgrade := upgradetypes.Plan{
		Name:   v505.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}

	var escrowAddresses []sdk.AccAddress
	transferChannels := app.IBCKeeper.ChannelKeeper.GetAllChannelsWithPortPrefix(ctx, app.TransferKeeper.GetPort(ctx))
	for _, channel := range transferChannels {
		escrowAddresses = append(escrowAddresses, transfertypes.GetEscrowAddress(channel.PortId, channel.ChannelId))
	}
	require.Greater(t, len(escrowAddresses), 0)
	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	for _, escrowAddress := range escrowAddresses {
		require.True(t, app.TokenFactoryKeeper.IsEscrowAddress(ctx, escrowAddress))
	}
	require.False(t, app.TokenFactoryKeeper.IsEscrowAddress(ctx, []byte{1, 2, 3, 4, 5}))
}
