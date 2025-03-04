package keeper_test

import (
	"math"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v5/x/dex/types"
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
	resp := s.aliceDeposits(
		NewDeposit(10, 0, 0, 1),
		NewDeposit(10, 0, 0, 3),
	)

	// THEN
	// assert 20 of token B deposited at tick 1 fee 0 and ticks unchanged
	s.True(resp.Reserve0Deposited[0].Equal(sdkmath.NewInt(10000000)))
	s.True(resp.Reserve0Deposited[1].Equal(sdkmath.NewInt(10000000)))
	s.EqualValues([]sdkmath.Int{sdkmath.ZeroInt(), sdkmath.ZeroInt()}, resp.Reserve1Deposited)
	s.EqualValues([]*types.FailedDeposit(nil), resp.FailedDeposits)
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
	resp := s.aliceDeposits(
		NewDeposit(0, 10, 0, 1),
		NewDeposit(0, 10, 0, 3),
	)

	// THEN
	// assert 20 of token B deposited at tick 1 fee 0 and ticks unchanged
	s.EqualValues([]sdkmath.Int{sdkmath.ZeroInt(), sdkmath.ZeroInt()}, resp.Reserve0Deposited)
	s.True(resp.Reserve1Deposited[0].Equal(sdkmath.NewInt(10000000)))
	s.True(resp.Reserve1Deposited[1].Equal(sdkmath.NewInt(10000000)))
	s.EqualValues([]*types.FailedDeposit(nil), resp.FailedDeposits)
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
	s.assertAliceDepositFails(err, NewDepositWithOptions(0, 5, 0, 1, types.DepositOptions{DisableAutoswap: true}))
}

func (s *DexTestSuite) TestDepositNilOptions() {
	s.fundAliceBalances(0, 1)
	msg := &types.MsgDeposit{
		Creator:         s.alice.String(),
		Receiver:        s.alice.String(),
		TokenA:          "TokenA",
		TokenB:          "TokenB",
		AmountsA:        []sdkmath.Int{sdkmath.ZeroInt()},
		AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
		TickIndexesAToB: []int64{0},
		Fees:            []uint64{1},
		Options:         []*types.DepositOptions{nil}, // WHEN options are nil
	}
	_, err := s.msgServer.Deposit(s.Ctx, msg)
	s.Assert().NoError(err)
}

func (s *DexTestSuite) TestDepositSingleLowTickUnderflowFails() {
	s.fundAliceBalances(0, 40_000_000_000_0)

	// GIVEN
	// deposit 50 of token B at tick -352436 fee 0
	// THEN 0 shares would be issued so deposit fails
	s.assertAliceDepositFails(
		types.ErrDepositShareUnderflow,
		NewDeposit(0, 26457, -240_000, 0),
	)
}

func (s *DexTestSuite) TestDepositSingleInvalidFeeFails() {
	s.fundAliceBalances(0, 50)

	// WHEN Deposit at fee 43 (invalid)
	// THEN FAILURE
	s.assertAliceDepositFails(
		types.ErrInvalidFee,
		NewDeposit(0, 50, 10, 43),
	)
}

func (s *DexTestSuite) TestDepositSinglewWhitelistedLPWithInvalidFee() {
	s.fundAliceBalances(0, 50)

	// Whitelist alice's address
	params := s.App.DexKeeper.GetParams(s.Ctx)
	params.WhitelistedLps = []string{s.alice.String()}
	err := s.App.DexKeeper.SetParams(s.Ctx, params)
	s.NoError(err)

	// WHEN Deposit at fee 43 (invalid)
	// THEN no error
	s.aliceDeposits(
		NewDeposit(0, 50, 10, 43),
	)
}

func (s *DexTestSuite) TestDepositSingleToken0BELFails() {
	s.fundAliceBalances(50, 50)

	// GIVEN TokenB liquidity at tick 2002-2004
	s.aliceDeposits(NewDeposit(0, 10, 2001, 1),
		NewDeposit(0, 10, 2002, 1),
		NewDeposit(0, 10, 2003, 1),
	)
	// WHEN alice deposits TokenA at tick -2003 (BEL)
	// THEN FAILURE
	s.assertAliceDepositFails(
		types.ErrDepositBehindEnemyLines,
		NewDepositWithOptions(50, 0, 2004, 1, types.DepositOptions{FailTxOnBel: true}),
	)
}

func (s *DexTestSuite) TestDepositSingleToken1BELFails() {
	s.fundAliceBalances(50, 50)

	// GIVEN TokenA liquidity at tick 2002-2005
	s.aliceDeposits(NewDeposit(10, 0, -2001, 1),
		NewDeposit(10, 0, -2003, 1),
		NewDeposit(10, 0, -2004, 1),
	)
	// WHEN alice deposits TokenB at tick -2003 (BEL)
	// THEN FAILURE
	s.assertAliceDepositFails(
		types.ErrDepositBehindEnemyLines,
		NewDepositWithOptions(0, 50, -2004, 1, types.DepositOptions{FailTxOnBel: true}),
	)
}
