package keeper_test

import (
	"math"
	"time"

	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
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

	// WHEN
	// place limit order on same tick (for B at tick -1)
	s.aliceLimitSells("TokenA", -1, 10)

	// THEN
	// assert 20 of token A deposited at tick 0 fee 0 and ticks unchanged
	s.assertLimitLiquidityAtTick("TokenA", -1, 20)
	s.assertAliceLimitLiquidityAtTick("TokenA", 20, -1)
	s.assertAliceBalances(30, 50)
	s.assertDexBalances(20, 0)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)
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

func (s *DexTestSuite) TestPlaceLimitOrderGTCWithDustMinAvgPriceFails() {
	s.fundAliceBalances(1, 0)
	s.fundBobBalances(0, 1)
	// GIVEN LP liq at 148.37-148.42 (with dust)
	s.bobDeposits(
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_000, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_001, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_002, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(10), 50_003, 1),
	)
	// THEN alice GTC limitOrder minAvgPrice == limitPrice fails
	limitPrice := math_utils.MustNewPrecDecFromStr("0.006737")
	_, err := s.msgServer.PlaceLimitOrder(s.Ctx, &types.MsgPlaceLimitOrder{
		Creator:             s.alice.String(),
		Receiver:            s.alice.String(),
		TokenIn:             "TokenA",
		TokenOut:            "TokenB",
		LimitSellPrice:      &limitPrice,
		AmountIn:            sdkmath.NewInt(2000),
		OrderType:           types.LimitOrderType_GOOD_TIL_CANCELLED,
		MinAverageSellPrice: &limitPrice,
	})
	s.ErrorIs(err, types.ErrLimitPriceNotSatisfied)
}

func (s *DexTestSuite) TestPlaceLimitOrderGTCWithDustMinAvgPrice() {
	s.fundAliceBalances(1, 0)
	s.fundBobBalances(0, 1)
	// GIVEN LP liq at 148.37-148.42 (with dust)
	s.bobDeposits(
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_000, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_001, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_002, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(10), 50_003, 1),
	)
	// THEN alice IoC limitOrder with lower min avg prices success
	limitPrice := math_utils.MustNewPrecDecFromStr("0.006737")
	minAvgPrice := math_utils.MustNewPrecDecFromStr("0.006728")
	_, err := s.msgServer.PlaceLimitOrder(s.Ctx, &types.MsgPlaceLimitOrder{
		Creator:             s.alice.String(),
		Receiver:            s.alice.String(),
		TokenIn:             "TokenA",
		TokenOut:            "TokenB",
		LimitSellPrice:      &limitPrice,
		AmountIn:            sdkmath.NewInt(2000),
		OrderType:           types.LimitOrderType_GOOD_TIL_CANCELLED,
		MinAverageSellPrice: &minAvgPrice,
	})
	s.NoError(err)
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

	s.assertAliceBalances(110, 40)
	s.assertBobBalances(90, 110)
	s.assertDexBalances(0, 50)
	s.assertCurrentTicks(math.MinInt64, 0)

	s.bobLimitSells("TokenA", 10, 10, types.LimitOrderType_FILL_OR_KILL)

	s.assertAliceBalances(110, 40)
	s.assertBobBalances(80, 120)
	s.assertDexBalances(10, 40)

	s.aliceCancelsLimitSell(trancheKey1)

	s.assertAliceBalances(120, 80)
	s.assertBobBalances(80, 120)
	s.assertDexBalances(0, 0)
}

func (s *DexTestSuite) TestPlaceLimitOrderWithDustHitsTruePriceLimit() {
	s.fundAliceBalances(1, 0)
	s.fundBobBalances(0, 1)
	// GIVEN LP liq at 20001-20004
	s.bobDeposits(
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(180), 20001, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(180), 20002, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 20003, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(10000), 20004, 1),
	)
	// THEN alice IoC limitOrder with limitPrice 20005 fails
	s.assertAliceLimitSellFails(types.ErrLimitPriceNotSatisfied, "TokenA", 20005, 1, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
}

func (s *DexTestSuite) TestPlaceLimitOrderIOCWithDustMinAvgPriceFails() {
	s.fundAliceBalances(1, 0)
	s.fundBobBalances(0, 1)
	// GIVEN LP liq at 148.37-148.42 (with dust)
	s.bobDeposits(
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_000, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_001, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_002, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(10), 50_003, 1),
	)
	// THEN alice IoC limitOrder minAvgPrice == limitPrice fails
	_, err := s.aliceLimitSellsWithMinAvgPrice(
		"TokenA",
		math_utils.MustNewPrecDecFromStr("0.006730"),
		1,
		math_utils.MustNewPrecDecFromStr("0.006730"),
		types.LimitOrderType_IMMEDIATE_OR_CANCEL,
	)
	s.ErrorIs(err, types.ErrLimitPriceNotSatisfied)
}

func (s *DexTestSuite) TestPlaceLimitOrderIOCWithDustMinAvgPrice() {
	s.fundAliceBalances(1, 0)
	s.fundBobBalances(0, 1)
	// GIVEN LP liq at 148.37-148.42 (with dust)
	s.bobDeposits(
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_000, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_001, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 50_002, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(10), 50_003, 1),
	)
	// THEN alice IoC limitOrder with lower min avg prices success
	_, err := s.aliceLimitSellsWithMinAvgPrice(
		"TokenA",
		math_utils.MustNewPrecDecFromStr("0.006730"),
		1,
		math_utils.MustNewPrecDecFromStr("0.006727"),
		types.LimitOrderType_IMMEDIATE_OR_CANCEL,
	)
	s.NoError(err)
}

func (s *DexTestSuite) TestPlaceLimitOrderIOCMinAvgPriceGTSellPriceFails() {
	s.fundAliceBalances(40, 0)
	s.fundBobBalances(0, 40)
	// GIVEN LP liq between taker price .995 and .992
	s.bobDeposits(
		NewDeposit(0, 10, 50, 1),
		NewDeposit(0, 10, 60, 1),
		NewDeposit(0, 10, 70, 1),
		NewDeposit(0, 10, 80, 1),
	)
	// THEN alice places IOC limitOrder with very low MinAveragePrice Fails
	_, err := s.aliceLimitSellsWithMinAvgPrice(
		"TokenA",
		math_utils.MustNewPrecDecFromStr("0.99"),
		40,
		math_utils.MustNewPrecDecFromStr("0.995"),
		types.LimitOrderType_IMMEDIATE_OR_CANCEL,
	)
	s.ErrorIs(err, types.ErrLimitPriceNotSatisfied)
}

func (s *DexTestSuite) TestPlaceLimitOrderIOCMinAvgPriceGTSellPrice() {
	s.fundAliceBalances(40, 0)
	s.fundBobBalances(0, 40)
	// GIVEN LP liq between taker price .995 and .992
	s.bobDeposits(
		NewDeposit(0, 10, 50, 1),
		NewDeposit(0, 10, 60, 1),
		NewDeposit(0, 10, 70, 1),
		NewDeposit(0, 10, 80, 1),
	)
	// THEN alice places IOC limitOrder with an achievable MinAveragePrice
	_, err := s.aliceLimitSellsWithMinAvgPrice(
		"TokenA",
		math_utils.MustNewPrecDecFromStr("0.99"),
		40,
		math_utils.MustNewPrecDecFromStr("0.993"),
		types.LimitOrderType_IMMEDIATE_OR_CANCEL,
	)
	s.NoError(err)
}

func (s *DexTestSuite) TestPlaceLimitOrderWithDust() {
	s.fundAliceBalances(1, 0)
	s.fundBobBalances(0, 1)
	// GIVEN LP liq at 20001-20004
	s.bobDeposits(
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(180), 20001, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(180), 20002, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 20003, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(10000), 20004, 1),
	)
	// WHEN alice submits IoC limitOrder with limitPrice 20006
	s.aliceLimitSells("TokenA", 20006, 1, types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	// THEN all liq is traded through
	s.assertAliceBalancesInt(sdkmath.NewInt(923409), sdkmath.NewInt(10361))
}

func (s *DexTestSuite) TestPlaceLimitOrderTooSmallAfterSwapFails() {
	s.fundAliceBalances(5, 0)
	s.fundBobBalances(0, 2)

	// GIVEN 2 TokenB at tick 0
	s.bobLimitSells("TokenB", 0, 2)
	// WHEN Alice limit Sells at tick 149,149
	// She swap through bobs liquidity leaving her with 3 tokens to place her maker order
	// At price of 0.000000333333333 her expected output for the maker leg will be zero after rounding
	// 3 * 149,149 * 1.0001^-149,149 = ~.9999

	// THEN Alice's order fails
	s.assertAliceLimitSellFails(types.ErrTradeTooSmall, "TokenA", 149_149, 5)
}

func (s *DexTestSuite) TestPlaceLimitOrderWithPrice0To1() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 100)

	// GIVEN
	// Alice place LO at price ~10.0
	trancheKey0 := s.limitSellsWithPrice(s.alice, "TokenA", math_utils.NewPrecDec(10), 10)

	// WHEN bob swaps through all of Alice's LO
	s.bobLimitSells("TokenB", -23078, 100, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
	s.aliceWithdrawsLimitSell(trancheKey0)

	// THEN alice gets out ~100 TOKENB and bob gets ~10 TOKENA
	s.assertAliceBalancesInt(sdkmath.ZeroInt(), sdkmath.NewInt(99_999_977))
	s.assertBobBalancesInt(sdkmath.NewInt(10000000), sdkmath.NewInt(22))
}

func (s *DexTestSuite) TestPlaceLimitOrderWithPrice1To0() {
	s.fundAliceBalances(0, 200)
	s.fundBobBalances(10, 0)
	makerPrice := math_utils.MustNewPrecDecFromStr("0.25")
	takerPrice := math_utils.MustNewPrecDecFromStr("3.99")
	// GIVEN
	// Alice place LO at price ~.25
	trancheKey0 := s.limitSellsWithPrice(s.alice, "TokenB", makerPrice, 200)

	// WHEN bob swaps through Alice's LO
	s.limitSellsWithPrice(s.bob, "TokenA", takerPrice, 10)
	s.aliceWithdrawsLimitSell(trancheKey0)

	// THEN alice gets out ~10 TOKENA and bob gets ~40 TOKENB
	s.assertAliceBalancesInt(sdkmath.NewInt(9999999), sdkmath.ZeroInt())
	s.assertBobBalancesInt(sdkmath.ZeroInt(), sdkmath.NewInt(39997453))
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
	s.assertAliceBalancesInt(sdkmath.NewInt(5), sdkmath.NewInt(1_353_046_854))
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
	s.assertAliceBalancesInt(sdkmath.NewInt(135_3046_854), sdkmath.NewInt(5))
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

	// THEN alice swap ~19 BIGTokenB and gets back 20 BIGTokenA
	s.assertAliceBalancesInt(sdkmath.NewInt(20_000_000), sdkmath.NewInt(31_162_769))
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

	// THEN alice swap ~19 BIGTokenB and gets back 20 BIGTokenA
	s.assertAliceBalancesInt(sdkmath.NewInt(20_000_000), sdkmath.NewInt(31_165_594))
}

func (s *DexTestSuite) TestPlaceLimitOrderFoKHitsTruePriceLimit() {
	s.fundAliceBalances(1, 0)
	s.fundBobBalances(0, 1)
	// GIVEN LP liq at 1-3. With small liq on tick 2
	s.bobDeposits(
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1), 0, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(2000), 1, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(1000), 2, 1),
		NewDepositInt(sdkmath.ZeroInt(), sdkmath.NewInt(2000), 3, 1),
	)
	// WHEN alice submits FoK limitOrder with limitPrice > 20004
	_, err := s.limitSellsInt(s.alice, "TokenA", 6, sdkmath.NewInt(3000), types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	s.ErrorIs(err, types.ErrLimitPriceNotSatisfied)
}

// Immediate Or Cancel LimitOrders ////////////////////////////////////////////////////////////////////

func (s *DexTestSuite) TestPlaceLimitOrderIoCNoLiq() {
	s.fundAliceBalances(10, 0)
	// GIVEN no liquidity
	// Thenalice IoC limitOrder fails
	s.assertAliceLimitSellFails(types.ErrNoLiquidity, "TokenA", 0, 10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
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
	// THEN alice IoC limitOrder for 10 tokenA below current 0To1 price fails
	s.assertAliceLimitSellFails(types.ErrNoLiquidity, "TokenA", -1, 10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
}

func (s *DexTestSuite) TestPlaceLimitOrderIoCTooSmallFails() {
	s.fundAliceBalances(1, 0)
	// WHEN Alice sells at a price where she would get less than 1 Token out
	// 1_000_000 * 1.0001^ 138163 =  0.9999013318

	// THEN Alice's order fails
	s.assertAliceLimitSellFails(types.ErrTradeTooSmall, "TokenA", 138_163, 1, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
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

func (s *DexTestSuite) TestPlaceLimitOrderJITNextBlock() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)

	// GIVEN Alice submits JIT limitOrder for 10 tokenA at tick 0 for block N
	trancheKey := s.aliceLimitSells("TokenA", 0, 10, types.LimitOrderType_JUST_IN_TIME)
	s.assertLimitLiquidityAtTick("TokenA", 0, 10)
	s.assertAliceBalances(0, 0)

	// WHEN we move to block N+1
	s.nextBlockWithTime(time.Now())
	s.beginBlockWithTime(time.Now())

	// THEN there is no liquidity available
	s.assertLimitLiquidityAtTick("TokenA", 0, 0)
	// Alice can withdraw the entirety of the unfilled limitOrder
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(10, 0)
}

func (s *DexTestSuite) TestPlaceLimitOrderJITTooManyFails() {
	s.fundAliceBalances(100, 0)

	// GIVEN Alice places JITS up to the MaxJITPerBlock limit
	for i := 0; i < int(types.DefaultMaxJITsPerBlock); i++ { //nolint:gosec
		s.aliceLimitSells("TokenA", i, 1, types.LimitOrderType_JUST_IN_TIME)
	}
	// WHEN Alive places another JIT order it fails

	s.assertAliceLimitSellFails(types.ErrOverJITPerBlockLimit, "TokenA", 0, 1, types.LimitOrderType_JUST_IN_TIME)

	// WHEN we move to next block alice can place more JITS
	s.nextBlockWithTime(time.Now())
	s.aliceLimitSells("TokenA", 0, 1, types.LimitOrderType_JUST_IN_TIME)
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
	s.beginBlockWithTime(time.Now().AddDate(0, 0, 2))

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
	// instead of moving chain forward by a block (BeginBlock + EndBlock + Commit) it's enough to jut move
	// ctx time forward to simulate passed time.
	newCtx := s.Ctx.WithBlockTime(time.Now().AddDate(0, 0, 2))
	s.Ctx = newCtx

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
		TradePairId:           defaultTradePairID1To0,
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

	_, err := s.msgServer.PlaceLimitOrder(s.Ctx, &types.MsgPlaceLimitOrder{
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

func (s *DexTestSuite) TestPlaceLimitOrderMixedTypes() {
	s.fundAliceBalances(4, 0)
	trancheKey1 := s.aliceLimitSellsGoodTil("TokenA", 0, 1, time.Now())
	trancheKey2 := s.aliceLimitSells("TokenA", 0, 1, types.LimitOrderType_JUST_IN_TIME)
	trancheKey3 := s.aliceLimitSells("TokenA", 0, 1, types.LimitOrderType_GOOD_TIL_CANCELLED)
	trancheKey4 := s.aliceLimitSells("TokenA", 0, 1, types.LimitOrderType_GOOD_TIL_CANCELLED)

	s.NotEqual(trancheKey1, trancheKey2, "GTT and JIT in same tranche")
	s.NotEqual(trancheKey1, trancheKey3, "GTC and GTT in same tranche")
	s.NotEqual(trancheKey2, trancheKey3, "GTC and JIT in same tranche")
	s.Equal(trancheKey4, trancheKey3, "GTCs not combined")
}
