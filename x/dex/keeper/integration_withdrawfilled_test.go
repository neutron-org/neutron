package keeper_test

import (
	"math"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v8/x/dex/types"
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

func (s *DexTestSuite) TestWithdrawUnfilledCancelled() {
	s.fundAliceBalances(1, 0)

	// GIVEN Alice places a limit order of A and then cancels it
	trancheKey := s.aliceLimitSells("TokenA", 0, 1)
	s.aliceCancelsLimitSell(trancheKey)

	// THEN she withdraws it fails
	s.aliceWithdrawLimitSellFails(types.ErrValidLimitOrderTrancheNotFound, trancheKey)
}

func (s *DexTestSuite) TestWithdrawPartiallyFilledCancelled() {
	s.fundAliceBalances(2, 0)
	s.fundBobBalances(0, 1)

	// GIVEN Alice places a limit order of A and then cancels it
	trancheKey := s.aliceLimitSells("TokenA", 0, 2)
	// WHEN Bob trades through half of Alice's limit order
	s.bobLimitSells("TokenB", -1, 1)

	// AND alice cancels the remainder
	s.aliceCancelsLimitSell(trancheKey)
	s.assertAliceBalances(1, 1)

	// AND her LimitOrderTrancheUser is removed
	_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.False(found, "Alice's LimitOrderTrancheUser not removed")
}

func (s *DexTestSuite) TestWithdrawUnfilledGTTFilledCancelled() {
	s.fundAliceBalances(1, 0)

	// GIVEN Alice places an expiring limit order of A
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", 0, 1, time.Now())

	// WHEN it is purged
	s.App.DexKeeper.PurgeExpiredLimitOrders(s.Ctx, time.Now())

	// THEN she can withdraw the amount she put in
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(1, 0)

	// AND her LimitOrderTrancheUser is removed
	_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.False(found, "Alice's LimitOrderTrancheUser not removed")
}

func (s *DexTestSuite) TestWithdrawPartiallyGTTFilledCancelled() {
	s.fundAliceBalances(5, 0)
	s.fundBobBalances(0, 750)

	// GIVEN Alice places an expiring limit order of A
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", -56990, 5, time.Now())

	// AND bob trades through half of it
	s.bobLimitSells("TokenB", -60000, 750)

	// WHEN it is purged
	s.App.DexKeeper.PurgeExpiredLimitOrders(s.Ctx, time.Now())

	// THEN she can withdraw the unused portion and the tokenOut
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalancesInt(sdkmath.NewInt(2487299), sdkmath.NewInt(749999999))

	// AND her LimitOrderTrancheUser is removed
	_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.False(found, "Alice's LimitOrderTrancheUser not removed")
}

func (s *DexTestSuite) TestWithdrawInactive() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)

	// GIVEN Alice places an expiring limit order of A
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", 0, 10, time.Now())

	// Bob trades through half of it
	s.bobLimitSells("TokenB", -1, 5)

	// Alice withdraws the profits
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(0, 5)

	// bob swap through more
	s.bobLimitSells("TokenB", -1, 4)

	// WHEN it is purged
	s.App.DexKeeper.PurgeExpiredLimitOrders(s.Ctx, time.Now())

	// THEN alice can withdraw the expected amount
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(1, 9)
}

func (s *DexTestSuite) TestWrongSharesProtectionWithdraw() {
	s.fundAliceBalances(1, 0)
	s.fundBobBalances(0, 1)

	trancheKey := s.aliceLimitSells("TokenA", 0, 1)
	s.bobLimitSells("TokenB", -1, 1)

	trancheUser, _ := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)

	// This should be impossible, but we still want to check that if there is a bug with the share calculation the user cannot withdraw more than the balance of the tranche
	trancheUser.SharesOwned = sdkmath.NewInt(1_000_000_000_000_000_000)

	s.App.DexKeeper.SetLimitOrderTrancheUser(s.Ctx, trancheUser)

	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalances(0, 1)
}
