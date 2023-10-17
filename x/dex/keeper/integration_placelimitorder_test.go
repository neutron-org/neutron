package keeper_test

import (
	"math"
	"time"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/neutron-org/neutron/x/dex/types"
)

// Core tests w/ GTC limitOrders //////////////////////////////////////////////
func (s *DexTestSuite) TestPlaceLimitOrderInSpread1To0() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -5, 5
	s.aliceDeposits(NewDeposit(10, 10, 0, 5))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-5)
	s.assertCurr0To1(5)

	// WHEN
	// place limit order for B at tick -1
	s.aliceLimitSells("TokenA", -1, 10)
	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)

	// THEN
	// assert currentTick1To0 moved
	s.assertCurr1To0(-1)
}

func (s *DexTestSuite) TestPlaceLimitOrderInSpread0To1() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -5, 5
	s.aliceDeposits(NewDeposit(10, 10, 0, 5))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-5)
	s.assertCurr0To1(5)

	// WHEN
	// place limit order for A at tick 1
	s.aliceLimitSells("TokenB", 1, 10)
	s.assertAliceBalances(40, 30)
	s.assertDexBalances(10, 20)

	// THEN
	// assert currentTick0To1 moved
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestPlaceLimitOrderInSpreadMinMaxNotAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -5, 5
	s.aliceDeposits(NewDeposit(10, 10, 0, 5))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)

	// WHEN
	// place limit order for B at tick -1
	s.aliceLimitSells("TokenA", -1, 10)
	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)

	// THEN
	// assert min, max not moved
}

func (s *DexTestSuite) TestPlaceLimitOrderOutOfSpread0To1NotAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	// WHEN
	// place limit order out of spread (for A at tick 3)
	s.aliceLimitSells("TokenB", 3, 10)
	s.assertAliceBalances(40, 30)
	s.assertDexBalances(10, 20)

	// THEN
	// assert currentTick0To1 not moved
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestPlaceLimitOrderOutOfSpread1To0NotAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	// WHEN
	// place limit order out of spread (for B at tick -3)
	s.aliceLimitSells("TokenA", -3, 10)
	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)

	// THEN
	// assert currentTick1To0 not moved
	s.assertCurr1To0(-1)
}

func (s *DexTestSuite) TestPlaceLimitOrderOutOfSpreadMinAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	// WHEN
	// place limit order out of spread (for B at tick -3)
	s.aliceLimitSells("TokenA", -3, 10)
	s.assertAliceBalances(30, 40)
	s.assertDexBalances(20, 10)

	// THEN
	// assert min moved
}

func (s *DexTestSuite) TestPlaceLimitOrderOutOfSpreadMaxAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	// WHEN
	// place limit order out of spread (for A at tick 3)
	s.aliceLimitSells("TokenB", 3, 10)
	s.assertAliceBalances(40, 30)
	s.assertDexBalances(10, 20)

	// THEN
	// assert max moved
}

func (s *DexTestSuite) TestPlaceLimitOrderOutOfSpreadMinNotAdjusted() {
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
	// place limit order in spread (for B at tick -3)
	s.aliceLimitSells("TokenA", -3, 10)
	s.assertAliceBalances(20, 40)
	s.assertDexBalances(30, 10)

	// THEN
	// assert min not moved
}

func (s *DexTestSuite) TestPlaceLimitOrderOutOfSpreadMaxNotAdjusted() {
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
	// place limit order in spread (for A at tick 3)
	s.aliceLimitSells("TokenB", 3, 10)
	s.assertAliceBalances(40, 20)
	s.assertDexBalances(10, 30)

	// THEN
	// assert max not moved
}

func (s *DexTestSuite) TestPlaceLimitOrderExistingLiquidityA() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// deposit 10 of token A at tick 0 fee 1
	s.aliceLimitSells("TokenA", -1, 10)
	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertLimitLiquidityAtTick("TokenA", -1, 10)
	s.assertAliceLimitLiquidityAtTick("TokenA", 10, -1)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)

	// // WHEN
	// // place limit order on same tick (for B at tick -1)
	// s.aliceLimitSells("TokenA", -1, 10)

	// // THEN
	// // assert 20 of token A deposited at tick 0 fee 0 and ticks unchanged
	// s.assertLimitLiquidityAtTick("TokenA", -1, 20)
	// s.assertAliceLimitLiquidityAtTick("TokenA", 20, -1)
	// s.assertAliceBalances(30, 50)
	// s.assertDexBalances(20, 0)
	// s.assertCurr1To0(-1)
	// s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestPlaceLimitOrderExistingLiquidityB() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// deposit 10 of token B at tick 1 fee 0
	s.aliceLimitSells("TokenB", 1, 10)
	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertLimitLiquidityAtTick("TokenB", 1, 10)
	s.assertAliceLimitLiquidityAtTick("TokenB", 10, 1)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)

	// WHEN
	// place limit order on same tick (for A at tick 1)
	s.aliceLimitSells("TokenB", 1, 10)

	// THEN
	// assert 20 of token B deposited at tick 0 fee 0 and ticks unchanged
	s.assertLimitLiquidityAtTick("TokenB", 1, 20)
	s.assertAliceLimitLiquidityAtTick("TokenB", 20, 1)
	s.assertAliceBalances(50, 30)
	s.assertDexBalances(0, 20)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestPlaceLimitOrderNoLOPlaceLODoesntIncrementPlaceTrancheKey() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// no previous LO on existing tick
	s.aliceDeposits(NewDeposit(10, 0, 0, 1))
	s.assertPoolLiquidity(10, 0, 0, 1)
	s.assertFillAndPlaceTrancheKeys("TokenA", -1, "", "")

	// WHEN
	// placing order on same tick
	trancheKey := s.aliceLimitSells("TokenA", -1, 10)

	// THEN
	// fill and place tranche keys don't change
	s.assertFillAndPlaceTrancheKeys("TokenA", -1, trancheKey, trancheKey)
}

func (s *DexTestSuite) TestPlaceLimitOrderUnfilledLOPlaceLODoesntIncrementPlaceTrancheKey() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// unfilled limit order exists on tick -1
	trancheKey := s.aliceLimitSells("TokenA", -1, 10)
	s.assertLimitLiquidityAtTick("TokenA", -1, 10)
	s.assertFillAndPlaceTrancheKeys("TokenA", -1, trancheKey, trancheKey)

	// WHEN
	// placing order on same tick
	s.aliceLimitSells("TokenA", -1, 10)

	// THEN
	// fill and place tranche keys don't change
	s.assertFillAndPlaceTrancheKeys("TokenA", -1, trancheKey, trancheKey)
}

func (s *DexTestSuite) TestPlaceLimitOrderPartiallyFilledLOPlaceLOIncrementsPlaceTrancheKey() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// partially filled limit order exists on tick -1
	trancheKey0 := s.aliceLimitSells("TokenA", -1, 10)
	s.bobLimitSells("TokenB", -10, 5, types.LimitOrderType_FILL_OR_KILL)
	s.assertFillAndPlaceTrancheKeys("TokenA", -1, trancheKey0, "")

	// WHEN
	// placing order on same tick
	trancheKey1 := s.aliceLimitSells("TokenA", -1, 10)

	// THEN
	// place tranche key changes
	s.assertFillAndPlaceTrancheKeys("TokenA", -1, trancheKey0, trancheKey1)
}

func (s *DexTestSuite) TestPlaceLimitOrderFilledLOPlaceLODoesntIncrementsPlaceTrancheKey() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// filled LO with partially filled place tranche
	trancheKey0 := s.aliceLimitSells("TokenA", -1, 10)
	s.bobLimitSells("TokenB", -10, 10, types.LimitOrderType_FILL_OR_KILL)
	trancheKey1 := s.aliceLimitSells("TokenA", -1, 10)
	s.assertFillAndPlaceTrancheKeys("TokenA", -1, trancheKey0, trancheKey1)

	// WHEN
	// placing order on same tick
	s.aliceLimitSells("TokenA", -1, 5)

	// THEN
	// fill and place tranche keys don't change
	s.assertFillAndPlaceTrancheKeys("TokenA", -1, trancheKey0, trancheKey1)
}

func (s *DexTestSuite) TestPlaceLimitOrderInsufficientFunds() {
	// GIVEN
	// alice has no funds
	s.assertAliceBalances(0, 0)

	// WHEN
	// place limit order selling non zero amount of token A for token B
	// THEN
	// deposit should fail with InsufficientFunds error

	err := sdkerrors.ErrInsufficientFunds
	s.assertAliceLimitSellFails(err, "TokenA", 0, 10)
}

func (s *DexTestSuite) TestLimitOrderPartialFillDepositCancel() {
	s.fundAliceBalances(100, 100)
	s.fundBobBalances(100, 100)
	s.assertDexBalances(0, 0)

	trancheKey0 := s.aliceLimitSells("TokenB", 0, 50)

	s.assertAliceBalances(100, 50)
	s.assertBobBalances(100, 100)
	s.assertDexBalances(0, 50)
	s.assertCurrentTicks(math.MinInt64, 0)

	s.bobLimitSells("TokenA", 10, 10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	s.assertAliceBalances(100, 50)
	s.assertBobBalances(90, 110)
	s.assertDexBalances(10, 40)
	s.assertCurrentTicks(math.MinInt64, 0)

	trancheKey1 := s.aliceLimitSells("TokenB", 0, 50)

	s.assertAliceBalances(100, 0)
	s.assertBobBalances(90, 110)
	s.assertDexBalances(10, 90)
	s.assertCurrentTicks(math.MinInt64, 0)

	s.aliceCancelsLimitSell(trancheKey0)

	s.assertAliceBalances(100, 40)
	s.assertBobBalances(90, 110)
	s.assertDexBalances(10, 50)
	s.assertCurrentTicks(math.MinInt64, 0)

	s.bobLimitSells("TokenA", 10, 10, types.LimitOrderType_FILL_OR_KILL)

	s.assertAliceBalances(100, 40)
	s.assertBobBalances(80, 120)
	s.assertDexBalances(20, 40)

	s.aliceCancelsLimitSell(trancheKey1)

	s.assertAliceBalances(100, 80)
	s.assertBobBalances(80, 120)
	s.assertDexBalances(20, 0)

	s.aliceWithdrawsLimitSell(trancheKey0)

	s.assertAliceBalances(110, 80)
	s.assertBobBalances(80, 120)
	s.assertDexBalances(10, 0)

	s.aliceWithdrawsLimitSell(trancheKey1)

	s.assertAliceBalances(120, 80)
	s.assertBobBalances(80, 120)
	s.assertDexBalances(0, 0)
}

// Fill Or Kill limit orders ///////////////////////////////////////////////////////////
func (s *DexTestSuite) TestPlaceLimitOrderFoKNoLiq() {
	s.fundAliceBalances(10, 0)
	// GIVEN no liquidity
	// THEN alice's LimitOrder fails
	s.assertAliceLimitSellFails(
		types.ErrFoKLimitOrderNotFilled,
		"TokenA",
		0,
		10,
		types.LimitOrderType_FILL_OR_KILL,
	)

	s.assertDexBalances(0, 0)
	s.assertAliceBalances(10, 0)
}

func (s *DexTestSuite) TestPlaceLimitOrderFoKWithLPFills() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)
	// GIVEN LP liq at tick -1
	s.bobDeposits(NewDeposit(0, 20, -1, 1))
	// WHEN alice submits FoK limitOrder
	s.aliceLimitSells("TokenA", 0, 10, types.LimitOrderType_FILL_OR_KILL)
	// THEN alice's LimitOrder fills via swap and auto-withdraws
	s.assertDexBalances(10, 10)
	s.assertAliceBalances(0, 10)

	// No maker LO is placed
	s.assertFillAndPlaceTrancheKeys("TokenA", 1, "", "")
	s.assertLimitLiquidityAtTick("TokenA", 1, 0)
}

func (s *DexTestSuite) TestPlaceLimitOrderFoKFailsWithInsufficientLiq() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)
	// GIVEN LP liq at tick -1 of 9 tokenB
	s.bobDeposits(NewDeposit(0, 9, -1, 1))
	// WHEN alice submits FoK limitOrder for 10 at tick 0 it fails
	s.assertAliceLimitSellFails(
		types.ErrFoKLimitOrderNotFilled,
		"TokenA",
		0,
		10,
		types.LimitOrderType_FILL_OR_KILL,
	)
}

func (s *DexTestSuite) TestPlaceLimitOrder0FoKFailsWithLowLimit() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)
	// GIVEN LP liq at tick -1 of 20 tokenB
	s.bobDeposits(NewDeposit(0, 20, -1, 1))
	// WHEN alice submits FoK limitOrder for 10 at tick -1 it fails
	s.assertAliceLimitSellFails(
		types.ErrFoKLimitOrderNotFilled,
		"TokenA",
		-1,
		10,
		types.LimitOrderType_FILL_OR_KILL,
	)
}

func (s *DexTestSuite) TestPlaceLimitOrder1FoKFailsWithHighLimit() {
	s.fundAliceBalances(0, 10)
	s.fundBobBalances(20, 0)
	// GIVEN LP liq at tick 20 of 20 tokenA
	s.bobDeposits(NewDeposit(20, 0, 20, 1))
	// WHEN alice submits FoK limitOrder for 10 at tick -1 it fails
	s.assertAliceLimitSellFails(
		types.ErrFoKLimitOrderNotFilled,
		"TokenB",
		21,
		10,
		types.LimitOrderType_FILL_OR_KILL,
	)
}

func (s *DexTestSuite) TestPlaceLimitOrderFoK0TotalAmountInNotUsed() {
	s.fundAliceBalances(9998, 0)
	s.fundBobBalances(0, 5000)
	// GIVEN LP liq at tick 20,000 & 20,001 of 1000 TokenB
	s.bobDeposits(
		NewDeposit(0, 1000, 20000, 1),
		NewDeposit(0, 1000, 20001, 1),
	)
	// WHEN alice submits FoK limitOrder for 9998 it succeeds
	// even though trueAmountIn < specifiedAmountIn due to rounding
	s.aliceLimitSells("TokenA", 21000, 9998, types.LimitOrderType_FILL_OR_KILL)
	s.assertAliceBalances(6, 1352)
}

func (s *DexTestSuite) TestPlaceLimitOrderFoK1TotalAmountInNotUsed() {
	s.fundAliceBalances(0, 9998)
	s.fundBobBalances(5000, 0)
	// GIVEN LP liq at tick -20,000 & -20,001 of 1000 tokenA
	s.bobDeposits(
		NewDeposit(1000, 0, -20000, 1),
		NewDeposit(1000, 0, -20001, 1),
	)
	// WHEN alice submits FoK limitOrder for 9998 it succeeds
	// even though trueAmountIn < specifiedAmountIn due to rounding
	s.aliceLimitSells("TokenB", -21000, 9998, types.LimitOrderType_FILL_OR_KILL)
	s.assertAliceBalances(1352, 6)
}

func (s *DexTestSuite) TestPlaceLimitOrderFoKMaxOutUsed() {
	s.fundAliceBalances(0, 50)
	s.fundBobBalances(50, 0)
	// GIVEN LP 50 TokenA at tick 600
	s.bobDeposits(
		NewDeposit(50, 0, 600, 1),
	)
	// WHEN alice submits FoK limitOrder of 50 TokenB with maxOut of 20
	s.aliceLimitSellsWithMaxOut("TokenB", 0, 50, 20)

	// THEN alice swap 19 TokenB and gets back 20 TokenA
	s.assertAliceBalances(20, 31)
}

func (s *DexTestSuite) TestPlaceLimitOrderFoKMaxOutUsedMultiTick() {
	s.fundAliceBalances(0, 50)
	s.fundBobBalances(50, 0)
	// GIVEN LP 30 TokenA at tick 600-602
	s.bobDeposits(
		NewDeposit(10, 0, 600, 1),
		NewDeposit(10, 0, 601, 1),
		NewDeposit(10, 0, 602, 1),
	)
	// WHEN alice submits FoK limitOrder of 50 TokenB with maxOut of 20
	s.aliceLimitSellsWithMaxOut("TokenB", 0, 50, 20)

	// THEN alice swap 20 TokenB and gets back 20 TokenA
	s.assertAliceBalances(20, 30)
}

// Immediate Or Cancel LimitOrders ////////////////////////////////////////////////////////////////////

func (s *DexTestSuite) TestPlaceLimitOrderIoCNoLiq() {
	s.fundAliceBalances(10, 0)
	// GIVEN no liquidity
	// WHEN alice submits IoC limitOrder
	s.aliceLimitSells("TokenA", 0, 10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	// THEN alice's LimitOrder is not filled
	s.assertDexBalances(0, 0)
	s.assertAliceBalances(10, 0)

	// No maker LO is placed
	s.assertLimitLiquidityAtTick("TokenA", 1, 0)
	s.assertFillAndPlaceTrancheKeys("TokenA", 1, "", "")
}

func (s *DexTestSuite) TestPlaceLimitOrderIoCWithLPFills() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)
	// GIVEN LP liq at tick -1
	s.bobDeposits(NewDeposit(0, 20, -1, 1))
	// WHEN alice submits IoC limitOrder
	s.aliceLimitSells("TokenA", 0, 10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	// THEN alice's LimitOrder fills via swap and auto-withdraws
	s.assertDexBalances(10, 10)
	s.assertAliceBalances(0, 10)

	// No maker LO is placed
	s.assertLimitLiquidityAtTick("TokenA", 1, 0)
	s.assertFillAndPlaceTrancheKeys("TokenA", 1, "", "")
}

func (s *DexTestSuite) TestPlaceLimitOrderIoCWithLPPartialFill() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)
	// GIVEN LP of 5 tokenB at tick -1
	s.bobDeposits(NewDeposit(0, 5, -1, 1))
	// WHEN alice submits IoC limitOrder for 10 tokenA
	s.aliceLimitSells("TokenA", 0, 10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	// THEN alice's LimitOrder swaps 5 TokenA and auto-withdraws
	s.assertDexBalances(5, 0)
	s.assertAliceBalances(5, 5)

	// No maker LO is placed
	s.assertFillAndPlaceTrancheKeys("TokenA", 1, "", "")
	s.assertLimitLiquidityAtTick("TokenA", 1, 0)
}

func (s *DexTestSuite) TestPlaceLimitOrderIoCWithLPNoFill() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)
	// GIVEN LP of 5 tokenB at tick -1
	s.bobDeposits(NewDeposit(0, 5, -1, 1))
	// WHEN alice submits IoC limitOrder for 10 tokenA below current 0To1 price
	s.aliceLimitSells("TokenA", -1, 10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	// THEN alice's LimitOrder doesn't fill and is canceled
	s.assertDexBalances(0, 5)
	s.assertAliceBalances(10, 0)
	// No maker LO is placed
	s.assertFillAndPlaceTrancheKeys("TokenA", 1, "", "")
	s.assertLimitLiquidityAtTick("TokenA", 1, 0)
}

// Just In Time Limit Orders //////////////////////////////////////////////////

func (s *DexTestSuite) TestPlaceLimitOrderJITFills() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)

	// GIVEN Alice submits JIT limitOrder for 10 tokenA at tick 0
	trancheKey := s.aliceLimitSells("TokenA", 0, 10, types.LimitOrderType_JUST_IN_TIME)
	s.assertLimitLiquidityAtTick("TokenA", 0, 10)
	s.assertAliceBalances(0, 0)

	// WHEN bob swaps through all the liquidity
	s.bobLimitSells("TokenB", -10, 10, types.LimitOrderType_FILL_OR_KILL)

	// THEN all liquidity is depleted
	s.assertLimitLiquidityAtTick("TokenA", 0, 0)
	// Alice can withdraw 10 TokenB
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(0, 10)
}

func (s *DexTestSuite) TestPlaceLimitOrderJITBehindEnemyLines() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)

	// GIVEN 10 LP liquidity for token exists at tick 0
	s.bobDeposits(NewDeposit(0, 10, 0, 1))

	// WHEN alice places a JIT limit order for TokenA at tick 1 (above enemy lines)
	trancheKey := s.aliceLimitSells("TokenA", 1, 10, types.LimitOrderType_JUST_IN_TIME)
	s.assertLimitLiquidityAtTick("TokenA", 1, 10)
	s.assertAliceBalances(0, 0)
	// AND bob swaps through all the liquidity
	s.bobLimitSells("TokenB", -10, 10, types.LimitOrderType_FILL_OR_KILL)

	// THEN all liquidity is depleted
	s.assertLimitLiquidityAtTick("TokenA", 1, 0)
	// Alice can withdraw 9 TokenB
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(0, 9)
}

func (s *DexTestSuite) TestPlaceLimitOrderJITNextBlock() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)

	// GIVEN Alice submits JIT limitOrder for 10 tokenA at tick 0 for block N
	trancheKey := s.aliceLimitSells("TokenA", 0, 10, types.LimitOrderType_JUST_IN_TIME)
	s.assertLimitLiquidityAtTick("TokenA", 0, 10)
	s.assertAliceBalances(0, 0)

	// WHEN we move to block N+1
	s.nextBlockWithTime(time.Now())
	s.App.EndBlock(abci.RequestEndBlock{Height: 0})

	// THEN there is no liquidity available
	s.assertLimitLiquidityAtTick("TokenA", 0, 0)
	// Alice can withdraw the entirety of the unfilled limitOrder
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(10, 0)
}

// GoodTilLimitOrders //////////////////////////////////////////////////

func (s *DexTestSuite) TestPlaceLimitOrderGoodTilFills() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)
	tomorrow := time.Now().AddDate(0, 0, 1)
	// GIVEN Alice submits JIT limitOrder for 10 tokenA expiring tomorrow
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", 0, 10, tomorrow)
	s.assertLimitLiquidityAtTick("TokenA", 0, 10)
	s.assertAliceBalances(0, 0)

	// WHEN bob swaps through all the liquidity
	s.bobLimitSells("TokenB", -10, 10, types.LimitOrderType_FILL_OR_KILL)

	// THEN all liquidity is depleted
	s.assertLimitLiquidityAtTick("TokenA", 0, 0)
	// Alice can withdraw 10 TokenB
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(0, 10)
}

func (s *DexTestSuite) TestPlaceLimitOrderGoodTilExpires() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)
	tomorrow := time.Now().AddDate(0, 0, 1)
	// GIVEN Alice submits JIT limitOrder for 10 tokenA expiring tomorrow
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", 0, 10, tomorrow)
	s.assertLimitLiquidityAtTick("TokenA", 0, 10)
	s.assertAliceBalances(0, 0)

	// When two days go by and multiple blocks are created (ie. purge is run)
	s.nextBlockWithTime(time.Now().AddDate(0, 0, 2))
	s.App.EndBlock(abci.RequestEndBlock{Height: 0})
	// THEN there is no liquidity available
	s.assertLimitLiquidityAtTick("TokenA", 0, 0)
	// Alice can withdraw the entirety of the unfilled limitOrder
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(10, 0)
}

func (s *DexTestSuite) TestPlaceLimitOrderGoodTilExpiresNotPurged() {
	// This is testing the case where the limitOrder has expired but has not yet been purged
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)
	tomorrow := time.Now().AddDate(0, 0, 1)
	// GIVEN Alice submits JIT limitOrder for 10 tokenA expiring tomorrow
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", 0, 10, tomorrow)
	s.assertLimitLiquidityAtTick("TokenA", 0, 10)
	s.assertAliceBalances(0, 0)

	// When two days go by
	// for simplicity sake we never run endBlock, it reality it would be run, but gas limit would be hit
	s.nextBlockWithTime(time.Now().AddDate(0, 0, 2))

	// THEN there is no liquidity available
	s.assertLimitLiquidityAtTick("TokenA", 0, 0)
	// Alice can cancel the entirety of the unfilled limitOrder
	s.aliceCancelsLimitSell(trancheKey)
	s.assertAliceBalances(10, 0)
}

func (s *DexTestSuite) TestPlaceLimitOrderGoodTilHandlesTimezoneCorrectly() {
	s.fundAliceBalances(10, 0)
	timeInPST, _ := time.Parse(time.RFC3339, "2050-01-02T15:04:05-08:00")
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", 0, 10, timeInPST)
	tranche := s.App.DexKeeper.GetLimitOrderTranche(s.Ctx, &types.LimitOrderTrancheKey{
		TradePairID:           defaultTradePairID1To0,
		TickIndexTakerToMaker: 0,
		TrancheKey:            trancheKey,
	})

	s.Assert().Equal(tranche.ExpirationTime.Unix(), timeInPST.Unix())
}

func (s *DexTestSuite) TestPlaceLimitOrderGoodTilAlreadyExpiredFails() {
	s.fundAliceBalances(10, 0)

	now := time.Now()
	yesterday := time.Now().AddDate(0, 0, -1)
	s.nextBlockWithTime(now)

	_, err := s.msgServer.PlaceLimitOrder(s.GoCtx, &types.MsgPlaceLimitOrder{
		Creator:          s.alice.String(),
		Receiver:         s.alice.String(),
		TokenIn:          "TokenA",
		TokenOut:         "TokenB",
		TickIndexInToOut: 0,
		AmountIn:         sdkmath.NewInt(50),
		OrderType:        types.LimitOrderType_GOOD_TIL_TIME,
		ExpirationTime:   &yesterday,
	})
	s.Assert().ErrorIs(err, types.ErrExpirationTimeInPast)
}
