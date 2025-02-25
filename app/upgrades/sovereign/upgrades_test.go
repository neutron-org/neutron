package v600_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/neutron-org/neutron/v5/app/params"
	v600 "github.com/neutron-org/neutron/v5/app/upgrades/sovereign"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"

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

func (suite *UpgradeTestSuite) TopUpWallet(ctx sdk.Context, sender, contractAddress sdk.AccAddress) {
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(1_000_000))))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
	suite.Require().NoError(err)
}

func (suite *UpgradeTestSuite) TestUpgrade() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	_, addr, _ := bech32.DecodeAndConvert("neutron1jxxfkkxd9qfjzpvjyr9h3dy7l5693kx4y0zvay")
	suite.TopUpWallet(ctx, senderAddress, addr)
	_, addr, _ = bech32.DecodeAndConvert("neutron1tedsrwal9n2qlp6j3xcs3fjz9khx7z4reep8k3")
	suite.TopUpWallet(ctx, senderAddress, addr)
	_, addr, _ = bech32.DecodeAndConvert("neutron1xdlvhs2l2wq0cc3eskyxphstns3348elwzvemh")
	suite.TopUpWallet(ctx, senderAddress, addr)

	err := v600.DeICS(ctx, *app.StakingKeeper, app.ConsumerKeeper, app.BankKeeper)
	require.NoError(t, err)

	vals, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Greater(t, len(vals), 0)
	//for _, v := range vals {
	//	fmt.Println(v)
	//}

	err = v600.SetupRevenue(ctx, *app.RevenueKeeper)
	require.NoError(t, err)

	resp, err := app.RevenueKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, resp.TwapWindow, int64(900))

	//_, err = app.RevenueKeeper.GetParams(ctx)
	//require.NoError(t, err)
}
