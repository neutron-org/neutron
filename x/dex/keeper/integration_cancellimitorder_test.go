package keeper_test

import (
	"math"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v4/x/dex/types"
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

	s.aliceCancelsLimitSellFails(trancheKey, types.ErrActiveLimitOrderNotFound)
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

	// Then alice gets back remaining 25 TokenA LO reserves
	s.assertAliceBalances(25, 0)
	s.assertDexBalances(0, 25)
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

	// THEN alice gets back ~41 BIGTokenA (125 * 1/3)
	s.assertAliceBalancesInt(sdkmath.NewInt(41_666_666), sdkmath.ZeroInt())

	// Carol gets back 83 TokenA (125 * 2/3)
	s.assertCarolBalancesInt(sdkmath.NewInt(83_333_333), sdkmath.ZeroInt())
	s.assertDexBalancesInt(sdkmath.OneInt(), sdkmath.NewInt(25_000_000))
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

func (s *DexTestSuite) TestCancelGoodTilAfterExpirationFails() {
	s.fundAliceBalances(50, 0)
	tomorrow := time.Now().AddDate(0, 0, 1)
	// GIVEN alice limit sells 50 TokenA with goodTil date of tommrow
	trancheKey := s.aliceLimitSellsGoodTil("TokenA", 0, 50, tomorrow)
	s.assertLimitLiquidityAtTick("TokenA", 0, 50)
	s.assertNLimitOrderExpiration(1)

	// WHEN expiration date has passed
	s.beginBlockWithTime(time.Now().AddDate(0, 0, 2))

	// THEN alice cancellation fails
	s.aliceCancelsLimitSellFails(trancheKey, types.ErrActiveLimitOrderNotFound)
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

	// THEN alice cancellation fails
	s.aliceCancelsLimitSellFails(trancheKey, types.ErrActiveLimitOrderNotFound)
	s.assertAliceBalances(0, 0)
}
