package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v10/x/dex/types"
)

func (s *DexTestSuite) TestWithdrawMultiFailure() {
	s.fundAliceBalances(50, 0)

	// GIVEN
	// alice deposits 5 A, 5 B in tick 0 fee 1
	s.aliceDeposits(
		NewDeposit(5, 0, 0, 1),
		NewDeposit(5, 0, 1, 1),
	)
	s.assertAliceShares(0, 1, 5)
	s.assertAliceShares(1, 1, 5)

	// WHEN
	// alice withdraws more shares than she owns from the second tick
	// THEN
	// failure on second withdraw (insufficient shares) will trigger ErrNotEnoughShares
	err := types.ErrInsufficientShares
	s.aliceWithdrawFails(err,
		NewWithdrawal(5, 0, 1),
		NewWithdrawal(10, 1, 1),
	)
}

func (s *DexTestSuite) TestWithdrawMultiSuccess() {
	s.fundAliceBalances(50, 0)

	// GIVEN
	// alice deposits 5 A, 5 B in tick 0 fee 1
	s.aliceDeposits(
		NewDeposit(5, 0, 0, 1),
		NewDeposit(5, 0, 1, 1),
		NewDeposit(5, 0, 2, 1),
	)
	s.assertAliceShares(0, 1, 5)
	s.assertAliceShares(1, 1, 5)
	s.assertAliceShares(2, 1, 5)
	s.assertLiquidityAtTick(5, 0, 0, 1)
	s.assertLiquidityAtTick(5, 0, 1, 1)
	s.assertLiquidityAtTick(5, 0, 2, 1)
	s.assertAliceBalances(35, 0)
	s.assertDexBalances(15, 0)

	// WHEN
	// alice withdraws 5 shares from each pool
	s.aliceWithdraws(
		NewWithdrawal(5, 0, 1),
		NewWithdrawal(5, 1, 1),
		NewWithdrawal(5, 2, 1),
	)

	// THEN
	// all withdraws should work
	// i.e. no shares remaining and entire balance transferred out
	s.assertAliceShares(0, 1, 0)
	s.assertDexBalances(0, 0)
}

func (s *DexTestSuite) TestWithdrawWithSharesMultiSuccess() {
	s.fundAliceBalances(50, 0)

	// GIVEN
	// alice deposits 5 A, 5 B in tick 0 fee 1
	s.aliceDeposits(
		NewDeposit(5, 0, 0, 1),
		NewDeposit(5, 0, 1, 1),
		NewDeposit(5, 0, 2, 1),
	)
	s.assertAliceShares(0, 1, 5)
	s.assertAliceShares(1, 1, 5)
	s.assertAliceShares(2, 1, 5)
	s.assertLiquidityAtTick(5, 0, 0, 1)
	s.assertLiquidityAtTick(5, 0, 1, 1)
	s.assertLiquidityAtTick(5, 0, 2, 1)
	s.assertAliceBalances(35, 0)
	s.assertDexBalances(15, 0)

	// WHEN
	// alice withdraws 5 shares from each pool
	s.withdrawsWithShares(
		s.alice,
		sdk.Coins{
			types.NewPoolShares(0, sdkmath.NewInt(5).Mul(denomMultiple)),
			types.NewPoolShares(1, sdkmath.NewInt(5).Mul(denomMultiple)),
			types.NewPoolShares(2, sdkmath.NewInt(5).Mul(denomMultiple)),
		},
	)

	// THEN
	// all withdraws should work
	// i.e. no shares remaining and entire balance transferred out
	s.assertAliceShares(0, 1, 0)
	s.assertDexBalances(0, 0)
}

func (s *DexTestSuite) TestWithdrawWithSharesMultiFailure() {
	s.fundAliceBalances(50, 0)

	// GIVEN
	// alice deposits 5 A, 5 B in tick 0 fee 1
	s.aliceDeposits(
		NewDeposit(5, 0, 0, 1),
		NewDeposit(5, 0, 1, 1),
		NewDeposit(5, 0, 2, 1),
	)

	// WHEN alice withdraws more shares than she owns from the third pool
	// THEN failure on third withdraw will trigger ErrInsufficientShares
	s.withdrawsWithSharesFails(s.alice,
		types.ErrInsufficientShares,
		sdk.Coins{
			types.NewPoolShares(0, sdkmath.NewInt(5).Mul(denomMultiple)),
			types.NewPoolShares(1, sdkmath.NewInt(5).Mul(denomMultiple)),
			types.NewPoolShares(2, sdkmath.NewInt(100).Mul(denomMultiple)),
		},
	)
}
