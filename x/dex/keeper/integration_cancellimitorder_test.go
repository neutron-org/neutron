package keeper_test

import (
	"math"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestCancelEntireLimitOrderAOneExists() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice adds a limit order of A for B and cancels it right away

	trancheKey := s.aliceLimitSells("TokenA", 0, 10)

	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(0)
	s.assertCurr0To1(math.MaxInt64)

	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(50, 50)
	s.assertDexBalances(0, 0)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)

	// Assert that the LimitOrderTrancheUser has been deleted
	_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.Assert().False(found)
}

func (s *DexTestSuite) TestCancelEntireLimitOrderBOneExists() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice adds a limit order of B for A and cancels it right away

	trancheKey := s.aliceLimitSells("TokenB", 0, 10)

	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(0)

	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(50, 50)
	s.assertDexBalances(0, 0)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(math.MaxInt64)

	// Tranche is deleted
	tranche, _, found := s.App.DexKeeper.FindLimitOrderTranche(
		s.Ctx,
		&types.LimitOrderTrancheKey{
			TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
			TickIndexTakerToMaker: 0,
			TrancheKey:            trancheKey,
		},
	)
	s.Nil(tranche)
	s.False(found)
}

func (s *DexTestSuite) TestCancelHigherEntireLimitOrderATwoExistDiffTicksSameDirection() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice adds two limit orders from A to B and removes the one at the higher tick (0)

	trancheKey := s.aliceLimitSells("TokenA", 0, 10)
	s.aliceLimitSells("TokenA", -1, 10)

	s.assertAliceBalances(30, 50)
	s.assertDexBalances(20, 0)
	s.assertCurr1To0(0)
	s.assertCurr0To1(math.MaxInt64)

	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestCancelLowerEntireLimitOrderATwoExistDiffTicksSameDirection() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice adds two limit orders from A to B and removes the one at the lower tick (-1)

	s.aliceLimitSells("TokenA", 0, 10)
	trancheKey := s.aliceLimitSells("TokenA", -1, 10)

	s.assertAliceBalances(30, 50)
	s.assertDexBalances(20, 0)
	s.assertCurr1To0(0)
	s.assertCurr0To1(math.MaxInt64)

	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertCurr1To0(0)
	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestCancelLowerEntireLimitOrderATwoExistDiffTicksDiffDirection() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice adds one limit orders from A to B and one from B to A and removes the one from A to B

	trancheKey := s.aliceLimitSells("TokenA", 0, 10)
	s.aliceLimitSells("TokenB", 1, 10)

	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(0)
	s.assertCurr0To1(1)

	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestCancelHigherEntireLimitOrderBTwoExistDiffTicksSameDirection() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice adds two limit orders from B to A and removes the one at tick 0

	trancheKey := s.aliceLimitSells("TokenB", 0, 10)
	s.aliceLimitSells("TokenB", -1, 10)

	s.assertAliceBalances(50, 30)
	s.assertDexBalances(0, 20)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(-1)

	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(-1)
}

func (s *DexTestSuite) TestCancelLowerEntireLimitOrderBTwoExistDiffTicksSameDirection() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice adds two limit orders from B to A and removes the one at tick 0

	s.aliceLimitSells("TokenB", 0, 10)
	trancheKey := s.aliceLimitSells("TokenB", -1, 10)

	s.assertAliceBalances(50, 30)
	s.assertDexBalances(0, 20)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(-1)

	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertCurr1To0(math.MinInt64)
	s.assertCurr0To1(0)
}

func (s *DexTestSuite) TestCancelTwiceFails() {
	s.fundAliceBalances(50, 50)
	// CASE
	// Alice tries to cancel the same limit order twice

	trancheKey := s.aliceLimitSells("TokenB", 0, 10)

	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)

	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(50, 50)
	s.assertDexBalances(0, 0)

	s.aliceCancelsLimitSellFails(trancheKey, types.ErrValidLimitOrderTrancheNotFound)
}

func (s *DexTestSuite) TestCancelPartiallyFilled() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(0, 50)

	// GIVEN alice limit sells 50 TokenA
	trancheKey := s.aliceLimitSells("TokenA", 0, 50)
	// Bob swaps 25 TokenB for TokenA
	s.bobLimitSells("TokenB", -10, 25, types.LimitOrderType_FILL_OR_KILL)

	s.assertDexBalances(25, 25)
	s.assertAliceBalances(0, 0)

	// WHEN alice cancels her limit order
	s.aliceCancelsLimitSell(trancheKey)

	// Then alice gets back remaining 25 TokenA LO reserves & 25 TokenB taker tokens
	s.assertAliceBalances(25, 25)
	s.assertDexBalances(0, 0)

	// Assert that the LimitOrderTrancheUser has been deleted
	_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.Assert().False(found)
}

func (s *DexTestSuite) TestCancelWithdrawThenCancel() {
	s.fundAliceBalances(10, 0)
	s.fundBobBalances(0, 20)

	// GIVEN alice limit sells 10 TokenA
	trancheKey := s.aliceLimitSells("TokenA", -6931, 10)
	// Bob swaps some TokenB for 5 TokenA
	s.bobLimitSellsWithMaxOut("TokenB", 7000, 20, 5)

	// WHEN alice withdraws
	s.aliceWithdrawsLimitSell(trancheKey)
	s.assertAliceBalancesInt(sdkmath.ZeroInt(), sdkmath.NewInt(9999181))

	// THEN Alice cancel still works
	s.aliceCancelsLimitSell(trancheKey)
	s.assertAliceBalancesInt(sdkmath.NewInt(4999999), sdkmath.NewInt(9999181))
}

func (s *DexTestSuite) TestCancelPartiallyFilledWithdrawFails() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(0, 10)

	// GIVEN alice limit sells 50 TokenA
	trancheKey := s.aliceLimitSells("TokenA", 2000, 50)
	// Bob swaps 10 TokenB for TokenA
	s.bobLimitSells("TokenB", -2001, 10, types.LimitOrderType_FILL_OR_KILL)

	s.assertDexBalancesInt(sdkmath.NewInt(37786095), sdkmath.NewInt(10000000))
	s.assertAliceBalances(0, 0)

	// WHEN alice cancels her limit order
	s.aliceCancelsLimitSell(trancheKey)

	// Then alice gets back remaining ~37 BIGTokenA LO reserves & 10 BIGTokenB taker tokens
	s.assertAliceBalancesInt(sdkmath.NewInt(37786094), sdkmath.NewInt(9999999))
	s.assertDexBalancesInt(sdkmath.OneInt(), sdkmath.OneInt())

	// Assert that the LimitOrderTrancheUser has been deleted
	_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.Assert().False(found)

	s.aliceWithdrawLimitSellFails(types.ErrValidLimitOrderTrancheNotFound, trancheKey)
}

func (s *DexTestSuite) TestCancelPartiallyFilledMultiUser() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(0, 50)
	s.fundCarolBalances(100, 0)

	// GIVEN alice limit sells 50 TokenA; carol sells 100 tokenA
	trancheKey := s.aliceLimitSells("TokenA", 0, 50)
	s.carolLimitSells("TokenA", 0, 100)
	// Bob swaps 25 TokenB for TokenA
	s.bobLimitSells("TokenB", -10, 25, types.LimitOrderType_FILL_OR_KILL)

	s.assertLimitLiquidityAtTick("TokenA", 0, 125)
	s.assertDexBalances(125, 25)
	s.assertAliceBalances(0, 0)

	// WHEN alice and carol cancel their limit orders
	s.aliceCancelsLimitSell(trancheKey)
	s.carolCancelsLimitSell(trancheKey)

	// THEN alice gets back ~41 BIGTokenA (125 * 1/3) & ~8.3 BIGTokenB Taker tokens (25 * 1/3)
	s.assertAliceBalancesInt(sdkmath.NewInt(41_666_666), sdkmath.NewInt(8333333))

	// Carol gets back 83 TokenA (125 * 2/3) & ~16.6 BIGTokenB Taker tokens (25 * 2/3)
	s.assertCarolBalancesInt(sdkmath.NewInt(83_333_333), sdkmath.NewInt(16666667))
	s.assertDexBalancesInt(sdkmath.OneInt(), sdkmath.ZeroInt())

	// Assert that the LimitOrderTrancheUsers has been deleted
	_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.Assert().False(found)
	_, found = s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.carol.String(), trancheKey)
	s.Assert().False(found)
}

func (s *DexTestSuite) TestCancelPartiallyFilledMultiUser2() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(50, 0)
	s.fundCarolBalances(0, 40)

	// // GIVEN alice and bob each limit sells 50 TokenA
	trancheKey := s.aliceLimitSells("TokenA", 2000, 50)
	s.bobLimitSells("TokenA", 2000, 50)
	// carol swaps 20 TokenB for TokenA
	s.carolLimitSells("TokenB", -2001, 20, types.LimitOrderType_FILL_OR_KILL)

	// WHEN alice cancels her limit order
	s.aliceCancelsLimitSell(trancheKey)

	// THEN alice gets back remaining ~38 BIGTokenA LO reserves & 10 BIGTokenB taker tokens
	s.assertAliceBalancesInt(sdkmath.NewInt(37786094), sdkmath.NewInt(9999999))
	s.assertDexBalancesInt(sdkmath.NewInt(37786096), sdkmath.NewInt(10000001))

	// THEN carol swap through more of the limitorder
	s.carolLimitSells("TokenB", -2001, 20, types.LimitOrderType_FILL_OR_KILL)

	// And bob can withdraw his portion
	s.bobWithdrawsLimitSell(trancheKey)
	s.assertBobBalancesInt(sdkmath.ZeroInt(), sdkmath.NewInt(30000000))
}

func (s *DexTestSuite) TestCancelFirstMultiCancel() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(50, 0)
	s.fundCarolBalances(0, 40)

	// // GIVEN alice and bob each limit sells 50 TokenA
	trancheKey := s.aliceLimitSells("TokenA", 0, 50)
	s.bobLimitSells("TokenA", 0, 50)
	s.bobCancelsLimitSell(trancheKey)
	// carol swaps 10 TokenB for TokenA
	s.carolLimitSells("TokenB", -1, 10, types.LimitOrderType_FILL_OR_KILL)

	// WHEN alice cancels her limit order
	s.aliceCancelsLimitSell(trancheKey)

	// THEN alice gets back remaining 40 tokenA  10 TokenB taker tokens
	s.assertAliceBalances(40, 10)
}

func (s *DexTestSuite) TestCancelFirstMultiWithdraw() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(50, 0)
	s.fundCarolBalances(0, 40)

	// // GIVEN alice and bob each limit sells 50 TokenA
	trancheKey := s.aliceLimitSells("TokenA", 0, 50)
	s.bobLimitSells("TokenA", 0, 50)
	s.bobCancelsLimitSell(trancheKey)
	// carol swaps 10 TokenB for TokenA
	s.carolLimitSells("TokenB", -1, 10, types.LimitOrderType_FILL_OR_KILL)

	// WHEN alice withdraws her limit order
	s.aliceWithdrawsLimitSell(trancheKey)

	// THEN alice gets 10 TokenB
	s.assertAliceBalances(0, 10)
}

func (s *DexTestSuite) TestCancelMultiAfterFilled() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(50, 0)
	s.fundCarolBalances(0, 100)

	// GIVEN alice and bob each limit sells 50 TokenA
	trancheKey := s.aliceLimitSells("TokenA", 0, 50)
	s.bobLimitSells("TokenA", 0, 50)

	// carol swaps through the tranche
	s.carolLimitSells("TokenB", -1, 100, types.LimitOrderType_IMMEDIATE_OR_CANCEL)

	// WHEN alice and bob cancel their limit order
	s.aliceCancelsLimitSell(trancheKey)
	s.assertAliceBalances(0, 50)
	s.bobCancelsLimitSell(trancheKey)
	s.assertBobBalances(0, 50)

	// THEN they get back the expected profit
	s.assertAliceBalances(0, 50)
	s.assertBobBalances(0, 50)

	// AND tranche and trancheUsers are deleted

	s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	_, _, found := s.App.DexKeeper.FindLimitOrderTranche(s.Ctx, &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenB", "TokenA"),
		TickIndexTakerToMaker: 0,
		TrancheKey:            trancheKey,
	})
	s.False(found)
	_, found = s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	s.Assert().False(found)
	_, found = s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.carol.String(), trancheKey)
	s.Assert().False(found)
}

func (s *DexTestSuite) TestCancelGoodTil() {
	s.fundAliceBalances(50, 0)
	tomorrow := time.Now().AddDate(0, 0, 1)
	// GIVEN alice limit sells 50 TokenA with goodTil date of tommrow
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", 0, 50, tomorrow)
	s.assertLimitLiquidityAtTick("TokenA", 0, 50)
	s.assertNLimitOrderExpiration(1)

	// WHEN alice cancels the limit order
	s.aliceCancelsLimitSell(trancheKey)
	// THEN there is no liquidity left
	s.assertLimitLiquidityAtTick("TokenA", 0, 0)
	// and the LimitOrderExpiration has been removed
	s.assertNLimitOrderExpiration(0)
}

func (s *DexTestSuite) TestCancelGoodTilAfterExpiration() {
	s.fundAliceBalances(50, 0)
	tomorrow := time.Now().AddDate(0, 0, 1)
	// GIVEN alice limit sells 50 TokenA with goodTil date of tommrow
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", 0, 50, tomorrow)
	s.assertLimitLiquidityAtTick("TokenA", 0, 50)
	s.assertNLimitOrderExpiration(1)

	// WHEN expiration date has passed
	s.beginBlockWithTime(time.Now().AddDate(0, 0, 2))

	// THEN alice cancellation succeeds
	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(50, 0)

	// TrancheUser and Tranche are removed
	s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	_, _, found := s.App.DexKeeper.FindLimitOrderTranche(s.Ctx, &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenB", "TokenA"),
		TickIndexTakerToMaker: 0,
		TrancheKey:            trancheKey,
	})
	s.False(found)
}

func (s *DexTestSuite) TestCancelJITSameBlock() {
	s.fundAliceBalances(50, 0)
	// GIVEN alice limit sells 50 TokenA as JIT
	trancheKey := s.aliceLimitSells("TokenA", 0, 50, types.LimitOrderType_JUST_IN_TIME)
	s.assertLimitLiquidityAtTick("TokenA", 0, 50)
	s.assertNLimitOrderExpiration(1)

	// WHEN alice cancels the limit order
	s.aliceCancelsLimitSell(trancheKey)
	// THEN there is no liquidity left
	s.assertLimitLiquidityAtTick("TokenA", 0, 0)
	// and the LimitOrderExpiration has been removed
	s.assertNLimitOrderExpiration(0)
}

func (s *DexTestSuite) TestCancelJITNextBlock() {
	s.fundAliceBalances(50, 0)
	// GIVEN alice limit sells 50 TokenA as JIT
	trancheKey := s.aliceLimitSells("TokenA", 0, 50, types.LimitOrderType_JUST_IN_TIME)
	s.assertLimitLiquidityAtTick("TokenA", 0, 50)
	s.assertNLimitOrderExpiration(1)

	// WHEN we move to block N+1
	s.nextBlockWithTime(time.Now())
	s.beginBlockWithTime(time.Now())

	// THEN alice cancellation succeeds
	s.aliceCancelsLimitSell(trancheKey)

	s.assertAliceBalances(50, 0)

	// TrancheUser and Tranche are removed
	s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.alice.String(), trancheKey)
	_, _, found := s.App.DexKeeper.FindLimitOrderTranche(s.Ctx, &types.LimitOrderTrancheKey{
		TradePairId:           types.MustNewTradePairID("TokenB", "TokenA"),
		TickIndexTakerToMaker: 0,
		TrancheKey:            trancheKey,
	})
	s.False(found)
}

func (s *DexTestSuite) TestWithdrawThenCancel() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(50, 0)
	s.fundCarolBalances(0, 40)

	// // GIVEN alice and bob each limit sells 50 TokenA
	trancheKey := s.aliceLimitSells("TokenA", 0, 50)
	s.bobLimitSells("TokenA", 0, 50)

	s.carolLimitSells("TokenB", -1, 10, types.LimitOrderType_FILL_OR_KILL)

	// WHEN alice withdraws and  cancels her limit order
	s.aliceWithdrawsLimitSell(trancheKey)
	s.aliceCancelsLimitSell(trancheKey)
	s.assertAliceBalances(45, 5)

	s.bobWithdrawsLimitSell(trancheKey)
	s.assertBobBalances(0, 5)
	s.bobCancelsLimitSell(trancheKey)
	s.assertBobBalances(45, 5)
}

func (s *DexTestSuite) TestWithdrawThenCancel2() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(50, 0)
	s.fundCarolBalances(0, 40)

	// // GIVEN alice and bob each limit sells 50 TokenA
	trancheKey := s.aliceLimitSells("TokenA", 0, 50)
	s.bobLimitSells("TokenA", 0, 50)

	s.carolLimitSells("TokenB", -1, 10, types.LimitOrderType_FILL_OR_KILL)

	// WHEN alice withdraws and  cancels her limit order
	s.aliceWithdrawsLimitSell(trancheKey)
	s.aliceCancelsLimitSell(trancheKey)
	s.assertAliceBalances(45, 5)

	s.bobCancelsLimitSell(trancheKey)
	s.assertBobBalances(45, 5)
}

func (s *DexTestSuite) TestWithdrawThenCancelLowTick() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(50, 0)
	s.fundCarolBalances(0, 40)

	// // GIVEN alice and bob each limit sells 50 TokenA
	trancheKey := s.aliceLimitSells("TokenA", 20000, 50)
	s.bobLimitSells("TokenA", 20000, 50)

	s.carolLimitSells("TokenB", -20001, 10, types.LimitOrderType_FILL_OR_KILL)

	// WHEN alice withdraws and  cancels her limit order
	s.aliceWithdrawsLimitSell(trancheKey)
	s.aliceCancelsLimitSell(trancheKey)
	s.assertAliceBalancesInt(sdkmath.NewInt(13058413), sdkmath.NewInt(4999999))

	s.bobWithdrawsLimitSell(trancheKey)
	s.assertBobBalancesInt(sdkmath.ZeroInt(), sdkmath.NewInt(4999999))
	s.bobCancelsLimitSell(trancheKey)
	s.assertBobBalancesInt(sdkmath.NewInt(13058413), sdkmath.NewInt(4999999))
}

func (s *DexTestSuite) TestCancelExpiringLimitOrderWithDust() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(0, 50)

	// GIVEN alice places a GTC limit order
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", -66671, 1, time.Now().Add(time.Second))

	// AND bob trades through alice's order
	s.bobLimitSells("TokenB", -66672, 50, types.LimitOrderType_FILL_OR_KILL)

	// WHEN alice cancels her limit order
	s.aliceCancelsLimitSell(trancheKey)

	// THEN A small amount of tokenA remains in the tranche
	s.assertLimitLiquidityAtTickInt("TokenA", -66671, sdkmath.NewInt(1))

	// The tranche is purged after the expiration time
	s.App.DexKeeper.PurgeExpiredLimitOrders(s.Ctx, time.Now().Add(time.Second*3))

	// THEN the tranche is deleted
	s.assertLimitLiquidityAtTick("TokenA", -66671, 0)
}
