package v5_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/testutil"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	dexkeeper "github.com/neutron-org/neutron/v4/x/dex/keeper"
	v5 "github.com/neutron-org/neutron/v4/x/dex/migrations/v5"
	"github.com/neutron-org/neutron/v4/x/dex/types"
)

type V4DexMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V4DexMigrationTestSuite))
}

func (suite *V4DexMigrationTestSuite) TestPoolMigrationSingleShareHolder() {
	var (
		app           = suite.GetNeutronZoneApp(suite.ChainA)
		ctx           = suite.ChainA.GetContext()
		alice         = []byte("alice")
		pairID        = &types.PairID{Token0: "TokenA", Token1: "TokenB"}
		depositAmount = math.NewInt(10_000)
	)

	// create a pool with 1 shareholder
	FundAccount(app.BankKeeper, ctx, alice, sdk.NewCoins(sdk.NewCoin("TokenA", depositAmount)))
	shares, err := suite.makeDeposit(ctx, app.DexKeeper, alice, pairID, depositAmount, math.ZeroInt(), 0, 1)
	suite.NoError(err)

	// run migrations
	suite.NoError(v5.MigrateStore(ctx, app.DexKeeper))

	// assert pool and shareholder balance are unchanged
	poolID, err := types.ParsePoolIDFromDenom(shares[0].Denom)
	suite.NoError(err)

	pool, _ := app.DexKeeper.GetPoolByID(ctx, poolID)

	suite.True(pool.LowerTick0.ReservesMakerDenom.Equal(depositAmount), "Pool value changed")
	aliceBalance := app.BankKeeper.GetAllBalances(ctx, alice)
	suite.True(aliceBalance.Equal(shares))
}

func (suite *V4DexMigrationTestSuite) TestPoolMigrationMultiShareHolder() {
	var (
		app            = suite.GetNeutronZoneApp(suite.ChainA)
		ctx            = suite.ChainA.GetContext()
		alice          = []byte("alice")
		bob            = []byte("bob")
		pairID         = &types.PairID{Token0: "TokenA", Token1: "TokenB"}
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

	// run migrations
	suite.NoError(v5.MigrateStore(ctx, app.DexKeeper))

	// assert that all users have withdrawn from the pool
	poolID, err := types.ParsePoolIDFromDenom(shares[0].Denom)
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
	if err := bankKeeper.MintCoins(ctx, types.ModuleName, amounts); err != nil {
		panic(err)
	}

	if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, amounts); err != nil {
		panic(err)
	}
}

func (suite *V4DexMigrationTestSuite) makeDeposit(
	ctx sdk.Context,
	k dexkeeper.Keeper,
	addr sdk.AccAddress,
	pairID *types.PairID,
	amount0, amount1 math.Int,
	tick int64,
	fee uint64,
) (sharesIssued sdk.Coins, err error) {
	deposit0, deposit1, sharesIssued, _, err := k.DepositCore(ctx, pairID, addr, addr, []math.Int{amount0}, []math.Int{amount1}, []int64{tick}, []uint64{fee}, []*types.DepositOptions{{}})
	suite.True(deposit0[0].Equal(amount0))
	suite.True(deposit1[0].Equal(amount1))

	return sharesIssued, err
}
