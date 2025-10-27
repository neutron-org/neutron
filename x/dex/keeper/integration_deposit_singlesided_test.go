package keeper_test

import (
	"math"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v8/x/dex/types"
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

func (s *DexTestSuite) TestDepositSingleToken0BELWithSwapPartial() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(0, 30)

	// GIVEN TokenB liquidity at tick 2002-2004
	s.bobDeposits(NewDeposit(0, 10, 2001, 1),
		NewDeposit(0, 10, 2002, 1),
		NewDeposit(0, 10, 2003, 1),
	)
	// WHEN alice deposits TokenA at tick -2005 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(50, 0, 2006, 1, types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN some of alice's tokenA is swapped and she deposits ~13TokenA & ~30TokenB
	// A = 50 - 30 * 1.0001^~2003 = 13.3
	// SharesIssued = 13.3 + 30 * 1.0001^2006 = ~50

	s.Equal(sdkmath.NewInt(13347290), resp.Reserve0Deposited[0])
	s.Equal(sdkmath.NewInt(30000000), resp.Reserve1Deposited[0])
	s.Equal(sdkmath.NewInt(50010996), resp.SharesIssued[0].Amount)
	s.assertAliceBalances(0, 0)

	s.assertLiquidityAtTickInt(sdkmath.NewInt(13347289), sdkmath.NewInt(30000000), 2006, 1)
}

func (s *DexTestSuite) TestDepositSingleToken0BELWithSwapAll() {
	s.fundAliceBalances(25, 0)
	s.fundBobBalances(0, 30)

	// GIVEN TokenB liquidity at tick 2002-2004
	s.bobDeposits(NewDeposit(0, 10, 2001, 1),
		NewDeposit(0, 10, 2002, 1),
		NewDeposit(0, 10, 2003, 1),
	)
	// WHEN alice deposits TokenA at tick -2005 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(25, 0, 2006, 1, types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN all of alice's TokenA is swapped and she deposits 0TokenA & ~20TokenB
	// B = 25 / 1.0001^~2003 = 20.4
	// SharesIssued = 20.4 * 1.0001^2006 = 25

	s.True(resp.Reserve0Deposited[0].IsZero())
	s.Equal(sdkmath.NewInt(20463288), resp.Reserve1Deposited[0])
	s.Equal(sdkmath.NewInt(25008666), resp.SharesIssued[0].Amount)
	s.assertAliceBalances(0, 0)

	s.assertLiquidityAtTickInt(sdkmath.ZeroInt(), sdkmath.NewInt(20463287), 2006, 1)
}

func (s *DexTestSuite) TestDepositSingleToken1BELWithSwapPartial() {
	s.fundAliceBalances(0, 50)
	s.fundBobBalances(20, 0)

	// GIVEN TokenA liquidity at tick 5002 & 5003
	s.bobDeposits(
		NewDeposit(10, 0, -5001, 1),
		NewDeposit(10, 0, -5002, 1),
	)
	// WHEN alice deposits TokenB at tick -5004 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(0, 50, -5005, 1, types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN some of alice's tokenB is swapped and she deposits 20TokenA & ~17TokenB
	// A = 20 (from swap)
	// B = 50 - 20 * 1.0001^~5002 = ~17
	// SharesIssued = 20 +  17 * 1.0001^-5005 = 30.3

	s.Equal(sdkmath.NewInt(20000000), resp.Reserve0Deposited[0])
	s.Equal(sdkmath.NewInt(17018155), resp.Reserve1Deposited[0])
	s.Equal(sdkmath.NewInt(30317131), resp.SharesIssued[0].Amount)
	s.assertAliceBalances(0, 0)

	s.assertLiquidityAtTickInt(sdkmath.NewInt(20000000), sdkmath.NewInt(17018154), -5005, 1)
}

func (s *DexTestSuite) TestDepositSingleToken1BELWithSwapAll() {
	s.fundAliceBalances(0, 5)
	s.fundBobBalances(20, 0)

	// GIVEN TokenA liquidity at tick -5000 & -5001
	s.bobDeposits(
		NewDeposit(10, 0, 5001, 1),
		NewDeposit(10, 0, 5002, 1),
	)
	// WHEN alice deposits TokenB at tick 5000 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(0, 5, 4999, 1,
			types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true},
		),
	)

	// THEN all of alice's TokenB is swapped and she deposits ~15TokenA & 0TokenB
	// A = 5 / 1.0001^-5001 = 8.2
	// SharesIssued = 8.2

	s.Equal(sdkmath.NewInt(8244225), resp.Reserve0Deposited[0])
	s.True(resp.Reserve1Deposited[0].IsZero())
	s.Equal(sdkmath.NewInt(8244224), resp.SharesIssued[0].Amount)
	s.assertAliceBalances(0, 0)

	s.assertLiquidityAtTickInt(sdkmath.NewInt(8244224), sdkmath.ZeroInt(), 4999, 1)
}

func (s *DexTestSuite) TestDepositSingleToken1BELWithSwapAll2() {
	s.fundAliceBalances(0, 20)
	s.fundBobBalances(10, 0)

	// GIVEN TokenA liquidity at tick 10,003
	s.bobDeposits(NewDeposit(10, 0, -10002, 1))
	// WHEN alice deposits TokenB at tick -10,004 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(0, 20, -10005, 1,
			types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN all of alice's TokenB is swapped
	// and she deposits ~7.3TokenA & 0TokenB
	// A = 20 / 1.0001^10003 = 7.3
	// SharesIssued = 7.3

	s.Equal(sdkmath.NewInt(7355750), resp.Reserve0Deposited[0])
	s.True(resp.Reserve1Deposited[0].IsZero())
	s.Equal(sdkmath.NewInt(7355749), resp.SharesIssued[0].Amount)
	s.assertAliceBalancesInt(sdkmath.ZeroInt(), sdkmath.ZeroInt())

	s.assertLiquidityAtTickInt(sdkmath.NewInt(7355749), sdkmath.ZeroInt(), -10005, 1)
}

func (s *DexTestSuite) TestDepositSingleToken1SameTickWithSwap() {
	s.fundAliceBalances(0, 20)
	s.fundBobBalances(10, 0)

	// GIVEN TokenA liquidity at tick 10,004
	s.bobDeposits(NewDeposit(10, 0, -10003, 1))
	// WHEN alice deposits TokenB at tick -10,004 (double sided liquidity; but not BEL )
	resp := s.aliceDeposits(
		NewDepositWithOptions(0, 20, -10005, 1, types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN no swap happens all of alice's TokenB is deposited at -10004
	s.True(resp.Reserve0Deposited[0].IsZero())
	s.Equal(sdkmath.NewInt(20000000), resp.Reserve1Deposited[0])
	s.Equal(sdkmath.NewInt(7354278), resp.SharesIssued[0].Amount)
	s.assertAliceBalancesInt(sdkmath.ZeroInt(), sdkmath.ZeroInt())

	s.assertLiquidityAtTickInt(sdkmath.ZeroInt(), sdkmath.NewInt(20000000), -10005, 1)
}

func (s *DexTestSuite) TestDepositSingleToken0SameTickWithSwap() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 10)

	// GIVEN TokenB liquidity at tick 1000
	s.bobDeposits(NewDeposit(0, 10, 999, 1))
	// WHEN alice deposits TokenA at tick -1000 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(10, 0, 1001, 1, types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN no swap happens all of alice's TokenA is deposited at tick -1000
	s.Equal(sdkmath.NewInt(10000000), resp.Reserve0Deposited[0])
	s.True(resp.Reserve1Deposited[0].IsZero())
	s.Equal(sdkmath.NewInt(10000000), resp.SharesIssued[0].Amount)
	s.assertAliceBalancesInt(sdkmath.ZeroInt(), sdkmath.ZeroInt())

	s.assertLiquidityAtTickInt(sdkmath.NewInt(10000000), sdkmath.ZeroInt(), 1001, 1)
}

func (s *DexTestSuite) TestDepositSingleToken0NotBELWithSwap() {
	s.fundAliceBalances(20, 0)
	s.fundBobBalances(0, 30)

	// GIVEN TokenB liquidity at tick 10,001
	s.bobDeposits(NewDeposit(0, 10, 10000, 1))
	// WHEN alice deposits TokenA at tick -49 (NOT BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(20, 0, 50, 1, types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN there is no swap and the deposit goes through as specified
	s.Equal(sdkmath.NewInt(20000000), resp.Reserve0Deposited[0])
	s.True(resp.Reserve1Deposited[0].IsZero())
	s.Equal(sdkmath.NewInt(20000000), resp.SharesIssued[0].Amount)
	s.assertAliceBalancesInt(sdkmath.ZeroInt(), sdkmath.ZeroInt())

	s.assertLiquidityAtTickInt(sdkmath.NewInt(20000000), sdkmath.ZeroInt(), 50, 1)
}

func (s *DexTestSuite) TestDepositSingleToken1NotBELWithSwap() {
	s.fundAliceBalances(0, 20)
	s.fundBobBalances(10, 0)

	// GIVEN TokenA liquidity at tick 10,003
	s.bobDeposits(NewDeposit(10, 0, -10002, 1))
	// WHEN alice deposits TokenB at tick -10,002 (NOT BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(0, 20, -10003, 1, types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN there is no swap and the deposit goes through as specified
	s.True(resp.Reserve0Deposited[0].IsZero())
	s.Equal(sdkmath.NewInt(20000000), resp.Reserve1Deposited[0])
	s.Equal(sdkmath.NewInt(7355749), resp.SharesIssued[0].Amount)
	s.assertAliceBalancesInt(sdkmath.ZeroInt(), sdkmath.ZeroInt())

	s.assertLiquidityAtTickInt(sdkmath.ZeroInt(), sdkmath.NewInt(20000000), -10003, 1)
}
