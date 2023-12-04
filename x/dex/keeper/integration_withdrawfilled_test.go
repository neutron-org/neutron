package keeper_test

import (
	"math"

	"github.com/neutron-org/neutron/v2/x/dex/types"
)

func (s *DexTestSuite) TestWithdrawFilledSimpleFull() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)
	// CASE
	// Alice places a limit order of A for B
	// Bob swaps from B to A
	// Alice withdraws the limit order

	trancheKey := s.aliceLimitSells("TokenA", 0, 10)

	s.assertAliceBalances(40, 50)
	s.assertBobBalances(50, 50)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(0)
	s.assertCurr0To1(math.MaxInt64)

	s.bobLimitSells("TokenB", -10, 10, types.LimitOrderType_FILL_OR_KILL)

	s.assertAliceBalances(40, 50)
	s.assertBobBalances(60, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)

	s.aliceWithdrawsLimitSell(trancheKey)

	s.assertAliceBalances(40, 60)
	s.assertBobBalances(60, 40)
	s.assertDexBalances(0, 0)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)

	// Assert that the LimitOrderTrancheUser has been deleted
	_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.Assert().False(found)
}

func (s *DexTestSuite) TestWithdrawFilledPartial() {
	s.fundAliceBalances(100, 100)
	s.fundBobBalances(100, 100)

	// GIVEN
	// alice limit sells 50 B at tick 0
	trancheKey := s.aliceLimitSells("TokenB", 0, 50)
	s.assertAliceLimitLiquidityAtTick("TokenB", 50, 0)
	// bob market sells 10 A
	s.bobLimitSells("TokenA", 10, 10, types.LimitOrderType_FILL_OR_KILL)
	// alice has 10 A filled
	s.assertAliceLimitFilledAtTickAtIndex("TokenB", 10, 0, trancheKey)
	// balances are 50, 100 for alice and 90, 100 for bob
	s.assertAliceBalances(100, 50)
	s.assertBobBalances(90, 110)

	// WHEN
	// alice withdraws filled limit order proceeds from tick 0 tranche 0
	s.aliceWithdrawsLimitSell(trancheKey)

	// THEN
	// limit order has been partially filled
	s.assertAliceLimitLiquidityAtTick("TokenB", 40, 0)
	// the filled reserved have been withdrawn from
	s.assertAliceLimitFilledAtTickAtIndex("TokenB", 0, 0, trancheKey)
	// balances are 110, 100 for alice and 90, 100 for bob
	s.assertAliceBalances(110, 50)
	s.assertBobBalances(90, 110)

	// the LimitOrderTrancheUser still exists
	_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.Assert().True(found)
}

func (s *DexTestSuite) TestWithdrawFilledTwiceFullSameDirection() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)
	// CASE
	// Alice places a limit order of A for B
	// Bob swaps through
	// Alice withdraws the limit order and places a new one
	// Bob swaps through again
	// Alice withdraws the limit order

	trancheKey0 := s.aliceLimitSells("TokenA", 0, 10)

	s.assertAliceBalances(40, 50)
	s.assertBobBalances(50, 50)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(0)
	s.assertCurr0To1(math.MaxInt64)

	s.bobLimitSells("TokenB", -10, 10, types.LimitOrderType_FILL_OR_KILL)

	s.assertAliceBalances(40, 50)
	s.assertBobBalances(60, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)

	s.aliceWithdrawsLimitSell(trancheKey0)
	trancheKey1 := s.aliceLimitSells("TokenA", 0, 10)

	s.assertAliceBalances(30, 60)
	s.assertBobBalances(60, 40)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(0)
	s.assertCurr0To1(math.MaxInt64)

	s.bobLimitSells("TokenB", -10, 10, types.LimitOrderType_FILL_OR_KILL)

	s.assertAliceBalances(30, 60)
	s.assertBobBalances(70, 30)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)

	s.aliceWithdrawsLimitSell(trancheKey1)

	s.assertAliceBalances(30, 70)
	s.assertBobBalances(70, 30)
	s.assertDexBalances(0, 0)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestWithdrawFilledTwiceFullDifferentDirection() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)
	// CASE
	// Alice places a limit order of A for B
	// Bob swaps through
	// Alice withdraws the limit order and places a new one
	// Bob swaps through again
	// Alice withdraws the limit order

	trancheKeyA := s.aliceLimitSells("TokenA", 0, 10)

	s.assertAliceBalances(40, 50)
	s.assertBobBalances(50, 50)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(0)
	s.assertCurr0To1(math.MaxInt64)

	s.bobLimitSells("TokenB", 0, 10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	s.assertAliceBalances(40, 50)
	s.assertBobBalances(60, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)

	s.aliceWithdrawsLimitSell(trancheKeyA)
	trancheKeyB := s.aliceLimitSells("TokenB", 0, 10)

	s.assertAliceBalances(40, 50)
	s.assertBobBalances(60, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(0)

	s.bobLimitSells("TokenA", 10, 10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	s.assertAliceBalances(40, 50)
	s.assertBobBalances(50, 50)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)

	s.aliceWithdrawsLimitSell(trancheKeyB)

	s.assertAliceBalances(50, 50)
	s.assertBobBalances(50, 50)
	s.assertDexBalances(0, 0)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestWithdrawFilledEmptyFilled() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// alice places limit order selling A for B at tick 0
	trancheKey := s.aliceLimitSells("TokenA", 0, 10)

	// WHEN
	// order is unfilled, i.e. trachne.filled = 0
	// THEN

	err := types.ErrWithdrawEmptyLimitOrder
	s.aliceWithdrawLimitSellFails(err, trancheKey)
}

func (s *DexTestSuite) TestWithdrawFilledNoExistingOrderByUser() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// only alice has an existing order placed
	trancheKey := s.aliceLimitSells("TokenA", 0, 10)

	// WHEN
	// bob tries to withdraw filled from tick 0 tranche 0
	// THEN

	err := types.ErrValidLimitOrderTrancheNotFound
	s.bobWithdrawLimitSellFails(err, trancheKey)
}

func (s *DexTestSuite) TestWithdrawFilledOtherUserOrder() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// only alice has a single existing order placed
	trancheKey := s.aliceLimitSells("TokenA", 0, 10)

	// WHEN
	// bob tries to withdraw with allice's TrancheKey
	// THEN

	err := types.ErrValidLimitOrderTrancheNotFound
	s.bobWithdrawLimitSellFails(err, trancheKey)
}
