package keeper_test

import (
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestWithdrawMultiFailure() {
	s.fundAliceBalances(50, 50)
	// GIVEN
	// alice deposits 5 A, 5 B in tick 0 fee 0
	s.aliceDeposits(NewDeposit(5, 5, 0, 1))
	s.assertAliceShares(0, 1, 10)
	s.assertLiquidityAtTick(5, 5, 0, 1)
	s.assertAliceBalances(45, 45)
	s.assertDexBalances(5, 5)

	// WHEN
	// alice withdraws 6 shares, then 10 shares
	// THEN
	// failure on second withdraw (insufficient shares) will trigger ErrNotEnoughShares
	err := types.ErrInsufficientShares
	s.aliceWithdrawFails(err,
		NewWithdrawal(6, 0, 1),
		NewWithdrawal(10, 0, 1),
	)
}

func (s *DexTestSuite) TestWithdrawMultiSuccess() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// alice deposits 5 A, 5 B in tick 0 fee 1
	s.aliceDeposits(NewDeposit(5, 5, 0, 1))
	s.assertAliceShares(0, 1, 10)
	s.assertLiquidityAtTick(5, 5, 0, 1)
	s.assertAliceBalances(45, 45)
	s.assertDexBalances(5, 5)

	// WHEN
	// alice withdraws 6 shares, then 4 shares
	s.aliceWithdraws(
		NewWithdrawal(6, 0, 1),
		NewWithdrawal(4, 0, 1),
	)

	// THEN
	// both withdraws should work
	// i.e. no shares remaining and entire balance transferred out
	s.assertAliceShares(0, 1, 0)
	s.assertLiquidityAtTick(0, 0, 0, 1)
	s.assertAliceBalances(50, 50)
	s.assertDexBalances(0, 0)
}
