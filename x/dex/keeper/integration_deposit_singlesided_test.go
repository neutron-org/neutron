package keeper_test

import (
	"math"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v3/x/dex/types"
)

func (s *DexTestSuite) TestDepositSingleSidedInSpread1To0() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -5, 5
	s.aliceDeposits(NewDeposit(10, 10, 0, 5))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-5)
	s.assertCurr0To1(5)

	// WHEN
	// deposit in spread (10 of A at tick 0 fee 1)
	s.aliceDeposits(NewDeposit(10, 0, 0, 1))
	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)

	// THEN
	// assert currentTick1To0 moved
	s.assertCurr1To0(-1)
}

func (s *DexTestSuite) TestDepositSingleSidedInSpread0To1() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -5, 5
	s.aliceDeposits(NewDeposit(10, 10, 0, 5))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-5)
	s.assertCurr0To1(5)

	// WHEN
	// deposit in spread (10 of B at tick 0 fee 1)
	s.aliceDeposits(NewDeposit(0, 10, 0, 1))
	s.assertAliceBalances(40, 30)
	s.assertDexBalances(10, 20)

	// THEN
	// assert currentTick0To1 moved
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestDepositSingleSidedInSpreadMinMaxNotAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -5, 5
	s.aliceDeposits(NewDeposit(10, 10, 0, 5))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)

	// WHEN
	// deposit in spread (10 of A at tick 0 fee 1)
	s.aliceDeposits(NewDeposit(10, 0, 0, 1))
	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)

	// THEN
	// assert min, max not moved
}

func (s *DexTestSuite) TestDepositSingleSidedOutOfSpread0To1NotAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	// WHEN
	// deposit out of spread (10 of B at tick 0 fee 3)
	s.aliceDeposits(NewDeposit(0, 10, 0, 3))
	s.assertAliceBalances(40, 30)
	s.assertDexBalances(10, 20)

	// THEN
	// assert currentTick0To1 not moved
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestDepositSingleSidedOutOfSpread1To0NotAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	// WHEN
	// deposit out of spread (10 of A at tick 0 fee 3)
	s.aliceDeposits(NewDeposit(10, 0, 0, 3))
	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)

	// THEN
	// assert currentTick1To0 not moved
	s.assertCurr1To0(-1)
}

func (s *DexTestSuite) TestDepositSingleSidedOutOfSpreadMinAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	// WHEN
	// deposit out of spread (10 of A at tick 0 fee 3)
	s.aliceDeposits(NewDeposit(10, 0, 0, 3))
	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)

	// THEN
	// assert min moved
}

func (s *DexTestSuite) TestDepositSingleSidedOutOfSpreadMaxAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	// WHEN
	// deposit out of spread (10 of B at tick 0 fee 3)
	s.aliceDeposits(NewDeposit(0, 10, 0, 3))
	s.assertAliceBalances(40, 30)
	s.assertDexBalances(10, 20)

	// THEN
	// assert max moved
}

func (s *DexTestSuite) TestDepositSingleSidedOutOfSpreadMinNotAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)
	// deposit new min at -5
	s.aliceDeposits(NewDeposit(10, 0, 0, 5))
	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)

	// WHEN
	// deposit in spread (10 of A at tick 0 fee 3)
	s.aliceDeposits(NewDeposit(10, 0, 0, 3))
	s.assertAliceBalances(20, 40)
	s.assertDexBalances(30, 10)

	// THEN
	// assert min not moved
}

func (s *DexTestSuite) TestDepositSingleSidedOutOfSpreadMaxNotAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)
	// deposit new max at 5
	s.aliceDeposits(NewDeposit(0, 10, 0, 5))
	s.assertAliceBalances(40, 30)
	s.assertDexBalances(10, 20)

	// WHEN
	// deposit out of spread (10 of B at tick 0 fee 3)
	s.aliceDeposits(NewDeposit(0, 10, 0, 3))
	s.assertAliceBalances(40, 20)
	s.assertDexBalances(10, 30)

	// THEN
	// assert max not moved
}

func (s *DexTestSuite) TestDepositSingleSidedExistingLiquidityA() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// deposit 10 of token A at tick 0 fee 1
	s.aliceDeposits(NewDeposit(10, 0, 0, 1))
	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertPoolLiquidity(10, 0, 0, 1)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)

	// WHEN
	// deposit 10 of token A on the same tick
	s.aliceDeposits(NewDeposit(10, 0, 0, 1))

	// THEN
	// assert 20 of token A deposited at tick 0 fee 0 and ticks unchanged
	s.assertAliceBalances(30, 50)
	s.assertDexBalances(20, 0)
	s.assertPoolLiquidity(20, 0, 0, 1)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestDepositSingleSidedExistingLiquidityB() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// deposit 10 of token B at tick 1 fee 0
	s.aliceDeposits(NewDeposit(0, 10, 0, 1))
	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertPoolLiquidity(0, 10, 0, 1)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)

	// WHEN
	// deposit 10 of token B on the same tick
	s.aliceDeposits(NewDeposit(0, 10, 0, 1))

	// THEN
	// assert 20 of token B deposited at tick 0 fee 0 and ticks unchanged
	s.assertPoolLiquidity(0, 20, 0, 1)
	s.assertAliceBalances(50, 30)
	s.assertDexBalances(0, 20)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestDepositSingleSidedCreatingArbToken0() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// deposit 10 of token B at tick 0 fee 1
	s.aliceDeposits(NewDeposit(0, 10, 0, 1))
	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertPoolLiquidity(0, 10, 0, 1)
	s.assertCurr0To1(1)

	// WHEN
	// depositing  above enemy lines at tick 1
	// THEN
	// deposit should not fail with BEL error, balances and liquidity should not change at deposited tick
	s.aliceDeposits(NewDeposit(10, 0, 4000, 1))

	// Bob arbs
	s.bobLimitSells("TokenB", -1, 50, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
	s.bobLimitSells("TokenA", 1, 10)
	s.assertBobBalancesInt(sdkmath.NewInt(50_000_000), sdkmath.NewInt(53_294_996))
}

func (s *DexTestSuite) TestDepositSingleSidedCreatingArbToken1() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// deposit 10 of token A at tick 0 fee 1
	s.aliceDeposits(NewDeposit(10, 0, 0, 1))
	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertPoolLiquidity(10, 0, 0, 1)
	s.assertCurr1To0(-1)

	// WHEN
	// depositing above enemy lines at tick -1
	// THEN
	// deposit should not fail with BEL error, balances and liquidity should not change at deposited tick

	s.aliceDeposits(NewDeposit(0, 10, -4000, 0))

	// Bob arbs
	s.bobLimitSells("TokenA", -1, 50, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
	s.bobLimitSells("TokenB", -1, 10)
	s.assertBobBalancesInt(sdkmath.NewInt(53_295_666), sdkmath.NewInt(50_000_000))
}

func (s *DexTestSuite) TestDepositSingleSidedMultiA() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// deposit 10 of token A at tick 0 fee 1
	s.aliceDeposits(NewDeposit(10, 0, 0, 1))
	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertPoolLiquidity(10, 0, 0, 1)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)

	// WHEN
	// multi deposit
	s.aliceDeposits(
		NewDeposit(10, 0, 0, 1),
		NewDeposit(10, 0, 0, 3),
	)

	// THEN
	// assert 20 of token B deposited at tick 1 fee 0 and ticks unchanged
	s.assertAliceBalances(20, 50)
	s.assertDexBalances(30, 0)
	s.assertPoolLiquidity(20, 0, 0, 1)
	s.assertPoolLiquidity(10, 0, 0, 3)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestDepositSingleSidedMultiB() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// deposit 10 of token B at tick 1 fee 0
	s.aliceDeposits(NewDeposit(0, 10, 0, 1))
	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertPoolLiquidity(0, 10, 0, 1)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)

	// WHEN
	// multi deposit at
	s.aliceDeposits(
		NewDeposit(0, 10, 0, 1),
		NewDeposit(0, 10, 0, 3),
	)

	// THEN
	// assert 20 of token B deposited at tick 1 fee 0 and ticks unchanged
	s.assertAliceBalances(50, 20)
	s.assertDexBalances(0, 30)
	s.assertPoolLiquidity(0, 20, 0, 1)
	s.assertPoolLiquidity(0, 10, 0, 3)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestDepositSingleSidedLowerTickOutsideRange() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// no existing liquidity

	// WHEN
	// depositing at the lower end of the acceptable range for ticks
	// THEN
	// deposit should fail with TickOutsideRange

	tickIndex := -1 * int(types.MaxTickExp)
	err := types.ErrTickOutsideRange
	s.assertAliceDepositFails(err, NewDeposit(10, 0, tickIndex, 1))
}

func (s *DexTestSuite) TestDepositSingleSidedUpperTickOutsideRange() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// no existing liquidity

	// WHEN
	// depositing at the lower end of the acceptable range for ticks
	// THEN
	// deposit should fail with TickOutsideRange

	tickIndex := int(types.MaxTickExp)
	err := types.ErrTickOutsideRange
	s.assertAliceDepositFails(err, NewDeposit(0, 10, tickIndex, 1))
}

func (s *DexTestSuite) TestDepositSingleSidedZeroTrueAmountsFail() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// alice deposits 5 A, 0 B at tick 0 fee 0
	s.aliceDeposits(NewDeposit(5, 0, 0, 1))

	// WHEN
	// alice deposits 0 A, 5 B at tick 0 fee 0
	// THEN
	// second deposit's ratio is different than pool after the first, so amounts will be rounded to 0,0 and tx will fail

	err := types.ErrZeroTrueDeposit
	s.assertAliceDepositFails(err, NewDeposit(0, 5, 0, 1))
}

// NOTE: The error checking for ShareUnderflow is completely subsumed by the ensureFairTruePrice check
// it no longer possible to manually check this test case. Leaving the example test here should things change in the future
// func (s *DexTestSuite) TestDepositSingleLowTickUnderflowFails() {
// 	s.fundAliceBalances(0, 40_000_000_000_0)

// 	// GIVEN
// 	// deposit 50 of token B at tick -352436 fee 0
// 	// THEN 0 shares would be issued so deposit fails
// 	s.assertAliceDepositFails(
// 		types.ErrDepositShareUnderflow,
// 		NewDeposit(0, 26457, -240_000, 0),
// 	)
// }

func (s *DexTestSuite) TestDepositSingleInvalidFeeFails() {
	s.fundAliceBalances(0, 50)

	// GIVEN
	// Deposit at fee 43 (invalid)
	// THEN FAILURE
	s.assertAliceDepositFails(
		types.ErrInvalidFee,
		NewDeposit(0, 50, 10, 43),
	)
}
