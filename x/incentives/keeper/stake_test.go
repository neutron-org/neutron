package keeper_test

import (
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

func (suite *IncentivesTestSuite) TestStakeLifecycle() {
	addr0 := suite.SetupAddr(0)

	// setup dex deposit and stake of those shares
	stake := suite.SetupDepositAndStake(depositStakeSpec{
		depositSpecs: []depositSpec{
			{
				addr:   addr0,
				token0: sdk.NewInt64Coin("TokenA", 1000),
				token1: sdk.NewInt64Coin("TokenB", 1000),
				tick:   0,
				fee:    1,
			},
		},
		stakeDistEpochOffset: -2,
	})

	retrievedStake, err := suite.App.IncentivesKeeper.GetStakeByID(suite.Ctx, stake.ID)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrievedStake)

	// unstake the full amount
	_, err = suite.App.IncentivesKeeper.Unstake(suite.Ctx, stake, sdk.Coins{})
	suite.Require().NoError(err)
	balances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr0)
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(suite.LPDenom0, 2000)), balances)
	_, err = suite.App.IncentivesKeeper.GetStakeByID(suite.Ctx, stake.ID)
	// should be deleted
	suite.Require().Error(err)
}

func (suite *IncentivesTestSuite) TestMultipleStakeLifecycle() {
	addr0 := suite.SetupAddr(0)

	// setup dex deposit and stake of those shares
	stake := suite.SetupDepositAndStake(depositStakeSpec{
		depositSpecs: []depositSpec{
			{
				addr:   addr0,
				token0: sdk.NewInt64Coin("TokenA", 1000),
				token1: sdk.NewInt64Coin("TokenB", 1000),
				tick:   0,
				fee:    1,
			},
			{
				addr:   addr0,
				token0: sdk.NewInt64Coin("TokenA", 1000),
				token1: sdk.NewInt64Coin("TokenB", 1000),
				tick:   1,
				fee:    1,
			},
		},
		stakeDistEpochOffset: -2,
	})

	retrievedStake, err := suite.App.IncentivesKeeper.GetStakeByID(suite.Ctx, stake.ID)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrievedStake)

	// unstake the full amount
	_, err = suite.App.IncentivesKeeper.Unstake(suite.Ctx, stake, sdk.Coins{})
	suite.Require().NoError(err)
	balances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr0)
	suite.Require().Equal(
		sdk.NewCoins(
			sdk.NewInt64Coin(suite.LPDenom0, 2000),
			sdk.NewInt64Coin(suite.LPDenom1, 2000),
		), balances)
	_, err = suite.App.IncentivesKeeper.GetStakeByID(suite.Ctx, stake.ID)
	// should be deleted
	suite.Require().Error(err)
}

func (suite *IncentivesTestSuite) TestStakeUnstakePartial() {
	addr0 := suite.SetupAddr(0)

	// setup dex deposit and stake of those shares
	stake := suite.SetupDepositAndStake(depositStakeSpec{
		depositSpecs: []depositSpec{
			{
				addr:   addr0,
				token0: sdk.NewInt64Coin("TokenA", 1000),
				token1: sdk.NewInt64Coin("TokenB", 1000),
				tick:   0,
				fee:    1,
			},
		},
		stakeDistEpochOffset: -2,
	})

	retrievedStake, err := suite.App.IncentivesKeeper.GetStakeByID(suite.Ctx, stake.ID)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrievedStake)

	// unstake the partial amount
	_, err = suite.App.IncentivesKeeper.Unstake(
		suite.Ctx,
		stake,
		sdk.Coins{sdk.NewInt64Coin(suite.LPDenom0, 900)},
	)
	suite.Require().NoError(err)
	balances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr0)
	suite.Require().ElementsMatch(sdk.NewCoins(sdk.NewInt64Coin(suite.LPDenom0, 900)), balances)
	// should still be accessible
	retrievedStake, err = suite.App.IncentivesKeeper.GetStakeByID(suite.Ctx, stake.ID)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrievedStake)
	suite.Require().
		ElementsMatch(sdk.NewCoins(sdk.NewInt64Coin(suite.LPDenom0, 1100)), retrievedStake.Coins)

	// fin.
}
