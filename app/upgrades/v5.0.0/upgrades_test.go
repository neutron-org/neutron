package v500_test

import (
	"testing"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	v500 "github.com/neutron-org/neutron/v5/app/upgrades/v5.0.0"
	"github.com/neutron-org/neutron/v5/testutil/common/sample"

	"github.com/neutron-org/neutron/v5/testutil"
	dexkeeper "github.com/neutron-org/neutron/v5/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v5/x/dex/types"
	"github.com/stretchr/testify/suite"
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

func (suite *UpgradeTestSuite) TestPoolMigrationSingleShareHolder() {
	var (
		app           = suite.GetNeutronZoneApp(suite.ChainA)
		ctx           = suite.ChainA.GetContext()
		alice         = []byte("alice")
		pairID        = &dextypes.PairID{Token0: "TokenA", Token1: "TokenB"}
		depositAmount = math.NewInt(10_000)
	)

	// create a pool with 1 shareholder
	FundAccount(app.BankKeeper, ctx, alice, sdk.NewCoins(sdk.NewCoin("TokenA", depositAmount)))
	shares, err := suite.makeDeposit(ctx, app.DexKeeper, alice, pairID, depositAmount, math.ZeroInt(), 0, 1)
	suite.NoError(err)

	// run upgrade
	upgrade := upgradetypes.Plan{
		Name:   v500.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	// assert pool and shareholder balance are unchanged
	poolID, err := dextypes.ParsePoolIDFromDenom(shares[0].Denom)
	suite.NoError(err)

	pool, _ := app.DexKeeper.GetPoolByID(ctx, poolID)

	suite.True(pool.LowerTick0.ReservesMakerDenom.Equal(depositAmount), "Pool value changed")
	aliceBalance := app.BankKeeper.GetAllBalances(ctx, alice)
	suite.True(aliceBalance.Equal(shares))
}

func (suite *UpgradeTestSuite) TestPoolMigrationMultiShareHolder() {
	var (
		app            = suite.GetNeutronZoneApp(suite.ChainA)
		ctx            = suite.ChainA.GetContext()
		alice          = []byte("alice")
		bob            = []byte("bob")
		pairID         = &dextypes.PairID{Token0: "TokenA", Token1: "TokenB"}
		depositAmount  = math.NewInt(10_000)
		initialBalance = sdk.NewCoins(sdk.NewCoin("TokenA", depositAmount))
	)
	FundAccount(app.BankKeeper, ctx, alice, initialBalance)
	FundAccount(app.BankKeeper, ctx, bob, initialBalance)

	// create a pool with 2 shareholders
	shares, err := suite.makeDeposit(ctx, app.DexKeeper, alice, pairID, depositAmount, math.ZeroInt(), 0, 1)
	suite.NoError(err)
	aliceBalance := app.BankKeeper.GetAllBalances(ctx, alice)
	suite.True(aliceBalance.Equal(shares))

	shares, err = suite.makeDeposit(ctx, app.DexKeeper, bob, pairID, depositAmount, math.ZeroInt(), 0, 1)
	suite.NoError(err)
	bobBalance := app.BankKeeper.GetAllBalances(ctx, bob)
	suite.True(bobBalance.Equal(shares))

	// run upgrade
	upgrade := upgradetypes.Plan{
		Name:   v500.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	// assert that all users have withdrawn from the pool
	poolID, err := dextypes.ParsePoolIDFromDenom(shares[0].Denom)
	suite.NoError(err)

	pool, _ := app.DexKeeper.GetPoolByID(ctx, poolID)
	suite.True(pool.LowerTick0.ReservesMakerDenom.Equal(math.ZeroInt()), "Pool not withdrawn")

	// AND funds are returned to the users
	aliceBalance = app.BankKeeper.GetAllBalances(ctx, alice)
	suite.True(aliceBalance.Equal(initialBalance))

	bobBalance = app.BankKeeper.GetAllBalances(ctx, bob)
	suite.True(bobBalance.Equal(initialBalance))
}

func FundAccount(bankKeeper bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) {
	if err := bankKeeper.MintCoins(ctx, dextypes.ModuleName, amounts); err != nil {
		panic(err)
	}

	if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, dextypes.ModuleName, addr, amounts); err != nil {
		panic(err)
	}
}

func (suite *UpgradeTestSuite) makeDeposit(
	ctx sdk.Context,
	k dexkeeper.Keeper,
	addr sdk.AccAddress,
	pairID *dextypes.PairID,
	amount0, amount1 math.Int,
	tick int64,
	fee uint64,
) (sharesIssued sdk.Coins, err error) {
	deposit0, deposit1, sharesIssued, _, err := k.DepositCore(ctx, pairID, addr, addr, []math.Int{amount0}, []math.Int{amount1}, []int64{tick}, []uint64{fee}, []*dextypes.DepositOptions{{}})
	suite.True(deposit0[0].Equal(amount0))
	suite.True(deposit1[0].Equal(amount1))

	return sharesIssued, err
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
