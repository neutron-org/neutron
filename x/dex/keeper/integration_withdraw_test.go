package keeper_test

import (
	"math"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestPartialWithdrawOnlyA() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice deposits 10 of A at tick 0, fee tier 0
	// and then withdraws only 5 shares of A

	// DATA
	// Alice should be credited 10 total shares
	// Shares = amount0 + price1to0 * amount1
	// Shares = 10 + 0 * 0 = 10
	s.aliceDeposits(NewDeposit(10, 0, 0, 1))

	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)

	s.aliceWithdraws(NewWithdrawal(5, 0, 1))

	s.assertAliceBalances(45, 50)
	s.assertDexBalances(5, 0)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestPartialWithdrawOnlyB() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice deposits 10 of B at tick 0, fee tier 0
	// and then withdraws only 5 shares of B

	// DATA
	// Alice should be credited 10 total shares
	// Shares = amount0 + price1to0 * amount1
	// Shares = 10 + 0 * 0 = 10
	s.aliceDeposits(NewDeposit(0, 10, 0, 1))

	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)

	s.aliceWithdraws(NewWithdrawal(5, 0, 1))

	s.assertAliceBalances(50, 45)
	s.assertDexBalances(0, 5)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestFullWithdrawOnlyB() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice deposits 10 of B at tick 0, fee tier 0
	// and then withdraws 10 shares of B

	// DATA
	// Alice should be credited 10 total shares
	// Shares = amount0 + price1to0 * amount1
	// Shares = 10 + 0 * 0 = 10
	s.aliceDeposits(NewDeposit(0, 10, 0, 1))

	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)

	s.aliceWithdraws(NewWithdrawal(10, 0, 1))

	s.assertAliceBalances(50, 50)
	s.assertDexBalances(0, 0)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestCurrentTickUpdatesAfterDoubleSidedThenSingleSidedDepositAndPartialWithdrawal() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice deposits 10 of A and B with a spread (fee) of +- 3 ticks
	// Alice then deposits 10 A with a spread (fee) of -1 ticks
	// Finally Alice withdraws from the first pool they deposited to

	s.aliceDeposits(NewDeposit(10, 10, 0, 3))

	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-3)
	s.assertCurr0To1(3)

	s.aliceDeposits(NewDeposit(10, 0, 0, 1))

	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(3)

	// DEBUG
	s.aliceWithdraws(NewWithdrawal(10, 0, 3))

	s.assertAliceBalances(35, 45)
	s.assertDexBalances(15, 5)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(3)
}

func (s *DexTestSuite) TestCurrentTickUpdatesAfterDoubleSidedThenSingleSidedDepositAndFulllWithdrawal() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice deposits 10 of A and B with a spread (fee) of +- 3 ticks
	// Alice then deposits 10 A with a spread (fee) of -1 ticks
	// Finally Alice withdraws from the first pool they deposited to

	s.aliceDeposits(NewDeposit(10, 10, 0, 3))

	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-3)
	s.assertCurr0To1(3)

	s.aliceDeposits(NewDeposit(10, 0, 0, 1))

	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(3)

	s.aliceWithdraws(NewWithdrawal(20, 0, 3))

	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestTwoFullDoubleSidedRebalancedAtooMuchTick0() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)
	// CASE
	// Alice deposits 5 of A and 10 of B at tick 0, fee tier 0
	// Bob tries to deposit 10 of A and 10 of B
	// Thus Bob should only end up depositing 5 of A and 10 of B
	// Alice then withdraws
	// David then withdraws

	s.aliceDeposits(NewDepositWithOptions(5, 10, 0, 1, types.DepositOptions{DisableAutoswap: true}))

	s.assertAliceBalances(45, 40)
	s.assertBobBalances(50, 50)
	s.assertDexBalances(5, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	s.bobDeposits(NewDepositWithOptions(10, 10, 0, 1, types.DepositOptions{DisableAutoswap: true}))

	s.assertAliceBalances(45, 40)
	s.assertBobBalances(45, 40)
	s.assertDexBalances(10, 20)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	s.aliceWithdraws(NewWithdrawal(15, 0, 1))

	s.assertAliceBalances(50, 50)
	s.assertBobBalances(45, 40)
	s.assertDexBalances(5, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	s.bobWithdraws(NewWithdrawal(15, 0, 1))

	s.assertAliceBalances(50, 50)
	s.assertBobBalances(50, 50)
	s.assertDexBalances(0, 0)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestTwoFullDoubleSidedRebalancedBtooMuchTick0() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)
	// CASE
	// Alice deposits 10 of B and 5 of Aat tick 0, fee tier 0
	// Bob tries to deposit 10 of A and 10 of B
	// Thus Bob should only end up depositing 5 of A and 10 of B
	// Alice then withdraws
	// David then withdraws

	s.aliceDeposits(NewDepositWithOptions(10, 5, 0, 1, types.DepositOptions{DisableAutoswap: true}))

	s.assertAliceBalances(40, 45)
	s.assertBobBalances(50, 50)
	s.assertDexBalances(10, 5)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	s.bobDeposits(NewDepositWithOptions(10, 10, 0, 1, types.DepositOptions{DisableAutoswap: true}))

	s.assertAliceBalances(40, 45)
	s.assertBobBalances(40, 45)
	s.assertDexBalances(20, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	s.aliceWithdraws(NewWithdrawal(15, 0, 1))

	s.assertAliceBalances(50, 50)
	s.assertBobBalances(40, 45)
	s.assertDexBalances(10, 5)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	s.bobWithdraws(NewWithdrawal(15, 0, 1))

	s.assertAliceBalances(50, 50)
	s.assertBobBalances(50, 50)
	s.assertDexBalances(0, 0)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestWithdrawalFailsWhenNotEnoughShares() {
	s.fundAliceBalances(100, 0)

	// IF  Alice deposits 100
	s.aliceDeposits(NewDeposit(100, 0, 0, 1))

	// WHEN Alice tries to withdraw 200
	// THEN ensure error is thrown and Alice and Dex balances remain unchanged
	err := types.ErrInsufficientShares
	s.aliceWithdrawFails(err, NewWithdrawal(200, 0, 1))
}

func (s *DexTestSuite) TestWithdrawalFailsWithNonExistentPair() {
	s.fundAliceBalances(100, 0)

	// IF Alice Deposists 100
	s.aliceDeposits(NewDeposit(100, 0, 0, 1))

	// WHEN Alice tries to withdraw from a nonexistent tokenPair
	_, err := s.msgServer.Withdrawal(s.Ctx, &types.MsgWithdrawal{
		Creator:         s.alice.String(),
		Receiver:        s.alice.String(),
		TokenA:          "TokenX",
		TokenB:          "TokenZ",
		SharesToRemove:  []sdkmath.Int{sdkmath.NewInt(10)},
		TickIndexesAToB: []int64{0},
		Fees:            []uint64{0},
	})

	// NOTE: As code is currently written we hit not enough shares check
	// before validating pair existence. This is correct from a
	// UX perspective --users should not care whether tick is initialized
	s.Assert().ErrorIs(err, types.ErrInsufficientShares)
}
