package keeper_test

import (
	"math"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/x/dex/types"
)

// TODO: In an ideal world, there should be enough lower level testing that the swap tests
// don't need to test both LO and LP. At the level of swap testing these should be indistinguishable.
func (s *DexTestSuite) TestSwap0To1NoLiquidity() {
	// GIVEN no liqudity of token B (deposit only token A and LO of token A)
	s.addDeposit(NewDeposit(10, 0, 0, 1))
	s.placeGTCLimitOrder("TokenA", 1000, 10)

	// WHEN swap 10 of tokenB
	tokenIn, tokenOut := s.swap("TokenA", "TokenB", 10)

	// THEN swap should do nothing
	s.assertSwapOutput(tokenIn, 0, tokenOut, 0)
	s.assertTickBalances(1010, 0)

	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestSwap1To0NoLiquidity() {
	// GIVEN no liqudity of token A (deposit only token B and LO of token B)
	s.addDeposit(NewDeposit(0, 10, 0, 1))
	s.placeGTCLimitOrder("TokenB", 1000, 10)

	// WHEN swap 10 of tokenB
	tokenIn, tokenOut := s.swap("TokenB", "TokenA", 10)

	// THEN swap should do nothing
	s.assertSwapOutput(tokenIn, 0, tokenOut, 0)
	s.assertTickBalances(0, 1010)

	s.assertCurr1To0(math.MinInt64)
}

// swaps against LPs only /////////////////////////////////////////////////////

func (s *DexTestSuite) TestSwap0To1PartialFillLP() {
	// GIVEN 10 tokenB LP @ tick 0 fee 1
	s.addDeposit(NewDeposit(0, 10, 0, 1))

	// WHEN swap 20 of tokenA
	tokenIn, tokenOut := s.swap("TokenA", "TokenB", 20)

	// THEN swap should return 11 TokenA in and 10 TokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 11, tokenOut, 10)
	s.assertTickBalances(11, 0)

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(-1)
}

func (s *DexTestSuite) TestSwap1To0PartialFillLP() {
	// GIVEN 10 tokenA LP @ tick 0 fee 1
	s.addDeposit(NewDeposit(10, 0, 0, 1))

	// WHEN swap 20 of tokenB
	tokenIn, tokenOut := s.swap("TokenB", "TokenA", 20)

	// THEN swap should return 11 TokenB in and 10 TokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 11, tokenOut, 10)
	s.assertTickBalances(0, 11)

	s.assertCurr0To1(1)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap0To1FillLP() {
	// GIVEN 100 tokenB LP @ tick 200 fee 5
	s.addDeposit(NewDeposit(0, 100, 200, 5))

	// WHEN swap 100 of tokenA
	tokenIn, tokenOut := s.swap("TokenA", "TokenB", 100)

	// THEN swap should return 100 TokenA in and 97 TokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 100, tokenOut, 97)
	s.assertTickBalances(100, 3)

	s.assertCurr0To1(205)
	s.assertCurr1To0(195)
}

func (s *DexTestSuite) TestSwap1To0FillLP() {
	// GIVEN 100 tokenA LP @ tick -20,000 fee 1
	s.addDeposit(NewDeposit(100, 0, -20_000, 1))

	// WHEN swap 100 of tokenB
	tokenIn, tokenOut := s.swap("TokenB", "TokenA", 100)

	// THEN swap should return 97 TokenB in and 13 TokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	// NOTE: Given rounding for amountOut, amountIn does not use the full maxAmount
	s.assertSwapOutput(tokenIn, 97, tokenOut, 13)
	s.assertTickBalances(87, 97)

	s.assertCurr0To1(-19_999)
	s.assertCurr1To0(-20_001)
}

func (s *DexTestSuite) TestSwap0To1FillLPHighFee() {
	// GIVEN 100 tokenB LP @ tick 20,000 fee 1,000
	s.addDeposit(NewDeposit(0, 100, 20_000, 1_000))

	// WHEN swap 100 of tokenA
	tokenIn, tokenOut := s.swap("TokenA", "TokenB", 100)

	// THEN swap should return 98 TokenA in and 12 TokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 98, tokenOut, 12)
	s.assertTickBalances(98, 88)

	s.assertCurr0To1(21_000)
	s.assertCurr1To0(19_000)
}

func (s *DexTestSuite) TestSwap1To0FillLPHighFee() {
	// GIVEN 1000 tokenA LP @ tick 20,000 fee 1000
	s.addDeposit(NewDeposit(1000, 0, 20_000, 1000))

	// WHEN swap 100 of tokenB
	tokenIn, tokenOut := s.swap("TokenB", "TokenA", 100)

	// THEN swap should return 100 TokenB in and 668 TokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 100, tokenOut, 668)
	s.assertTickBalances(332, 100)

	s.assertCurr0To1(21_000)
	s.assertCurr1To0(19_000)
}

func (s *DexTestSuite) TestSwap0To1PartialFillMultipleLP() {
	// GIVEN 300 worth of tokenB LPs
	s.addDeposits(
		NewDeposit(0, 100, -20_000, 1),
		NewDeposit(0, 100, -20_001, 1),
		NewDeposit(0, 100, -20_002, 1),
	)

	// WHEN swap 100 of tokenA
	tokenIn, tokenOut := s.swap("TokenA", "TokenB", 100)

	// THEN swap should return 42 TokenA in and 300 TokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 42, tokenOut, 300)
	s.assertTickBalances(42, 0)

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(-20_001)
}

func (s *DexTestSuite) TestSwap1To0PartialFillMultipleLP() {
	// GIVEN 300 worth of tokenA LPs
	s.addDeposits(
		NewDeposit(100, 0, 20_000, 1),
		NewDeposit(100, 0, 20_001, 1),
		NewDeposit(100, 0, 20_002, 1),
	)

	// WHEN swap 100 of tokenB
	tokenIn, tokenOut := s.swap("TokenB", "TokenA", 100)

	// THEN swap should return 42 TokenB in and 300 TokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 42, tokenOut, 300)
	s.assertTickBalances(0, 42)

	s.assertCurr0To1(20_001)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap0To1FillMultipleLP() {
	// GIVEN 400 worth of tokenB LPs
	s.addDeposits(
		NewDeposit(0, 100, -20, 1),
		NewDeposit(0, 100, -21, 1),
		NewDeposit(0, 100, -22, 1),
		NewDeposit(0, 100, -23, 1),
	)

	// WHEN swap 100 of tokenA
	tokenIn, tokenOut := s.swap("TokenA", "TokenB", 400)

	// THEN swap should return 400 TokenA in and 400 TokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 400, tokenOut, 400)
	s.assertTickBalances(400, 0)

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(-21)
}

func (s *DexTestSuite) TestSwap1To0FillMultipleLP() {
	// GIVEN 400 worth of tokenA LPs
	s.addDeposits(
		NewDeposit(100, 0, 20, 1),
		NewDeposit(100, 0, 21, 1),
		NewDeposit(100, 0, 22, 1),
		NewDeposit(100, 0, 23, 1),
	)

	// WHEN swap 400 of tokenB
	tokenIn, tokenOut := s.swap("TokenB", "TokenA", 400)

	// THEN swap should return 400 TokenB in and 400 TokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 400, tokenOut, 400)
	s.assertTickBalances(0, 400)

	s.assertCurr0To1(21)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap0To1LPMaxAmountUsed() {
	// GIVEN 10 TokenB available
	s.addDeposits(NewDeposit(0, 10, 0, 1))

	// WHEN swap 50 TokenA with maxOut of 5
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 50, 5)

	// THEN swap should return 6 TokenA in and 5 TokenB out
	s.assertSwapOutput(tokenIn, 6, tokenOut, 5)
	s.assertTickBalances(6, 5)
}

func (s *DexTestSuite) TestSwap1To0LPMaxAmountUsed() {
	// GIVEN 10 TokenA available
	s.addDeposits(NewDeposit(10, 0, 0, 1))

	// WHEN swap 50 TokenB with maxOut of 5
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 50, 5)

	// THEN swap should return 6 TokenB in and 5 TokenA out
	s.assertSwapOutput(tokenIn, 6, tokenOut, 5)
	s.assertTickBalances(5, 6)
}

func (s *DexTestSuite) TestSwap0To1LPMaxAmountNotUsed() {
	// GIVEN 10 TokenB available
	s.addDeposits(NewDeposit(0, 10, 0, 1))

	// WHEN swap 8 with maxOut of 15
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 8, 15)

	// THEN swap should return 8 TokenA in and 7 TokenB out
	s.assertSwapOutput(tokenIn, 8, tokenOut, 7)
	s.assertTickBalances(8, 3)
}

func (s *DexTestSuite) TestSwap1To0LPMaxAmountNotUsed() {
	// GIVEN 10 TokenA available
	s.addDeposits(NewDeposit(10, 0, 0, 1))

	// WHEN swap 8 with maxOut of 15
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 8, 15)

	// THEN swap should return 8 TokenB in and 7 TokenA out
	s.assertSwapOutput(tokenIn, 8, tokenOut, 7)
	s.assertTickBalances(3, 8)
}

func (s *DexTestSuite) TestSwap0To1LPMaxAmountUsedMultiTick() {
	// GIVEN 50 TokenB available
	s.addDeposits(
		NewDeposit(0, 5, 0, 1),
		NewDeposit(0, 5, 1, 1),
		NewDeposit(0, 5, 2, 1),
		NewDeposit(0, 5, 3, 1),
		NewDeposit(0, 30, 4, 1),
	)

	// WHEN swap 50 TokenA with maxOut of 20
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 50, 20)

	// THEN swap should return 24 TokenA in and 20 TokenB out
	s.assertSwapOutput(tokenIn, 24, tokenOut, 20)
	s.assertTickBalances(24, 30)
}

func (s *DexTestSuite) TestSwap1To0LPMaxAmountUsedMultiTick() {
	// GIVEN 50 TokenA available
	s.addDeposits(
		NewDeposit(5, 0, 0, 1),
		NewDeposit(5, 0, 1, 1),
		NewDeposit(5, 0, 2, 1),
		NewDeposit(5, 0, 3, 1),
		NewDeposit(30, 0, 4, 1),
	)

	// WHEN swap 50 TokenB with maxOut of 20
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 50, 20)

	// THEN swap should return 20 TokenB in and 20 TokenA out
	s.assertSwapOutput(tokenIn, 20, tokenOut, 20)
	s.assertTickBalances(30, 20)
}

func (s *DexTestSuite) TestSwap0To1LPMaxAmountNotUsedMultiTick() {
	// GIVEN 50 TokenB available
	s.addDeposits(
		NewDeposit(0, 5, 0, 1),
		NewDeposit(0, 5, 1, 1),
		NewDeposit(0, 5, 2, 1),
		NewDeposit(0, 5, 3, 1),
		NewDeposit(0, 30, 4, 1),
	)

	// WHEN swap 19 TokenA with maxOut of 20
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 19, 20)

	// THEN swap should return 19 TokenA in and 15 TokenB out
	s.assertSwapOutput(tokenIn, 18, tokenOut, 15)
	s.assertTickBalances(18, 35)
}

// swaps against LOs only /////////////////////////////////////////////////////

func (s *DexTestSuite) TestSwap0To1PartialFillLO() {
	// GIVEN 10 tokenB LO @ tick 1,000
	s.placeGTCLimitOrder("TokenB", 10, 1_000)

	// WHEN swap 20 of tokenA
	tokenIn, tokenOut := s.swap("TokenA", "TokenB", 20)

	// THEN swap should return 12 TokenA in and 10 TokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 12, tokenOut, 10)
	s.assertTickBalances(12, 0)
}

func (s *DexTestSuite) TestSwap1To0PartialFillLO() {
	// GIVEN 10 tokenA LO @ tick -1,000
	s.placeGTCLimitOrder("TokenA", 10, -1_000)

	// WHEN swap 20 of tokenB
	tokenIn, tokenOut := s.swap("TokenB", "TokenA", 20)

	// THEN swap should return 12 TokenB in and 10 TokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 12, tokenOut, 10)
	s.assertTickBalances(0, 12)

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap0To1FillLO() {
	// GIVEN 100 tokenB LO @ tick 10,000
	s.placeGTCLimitOrder("TokenB", 100, 10_000)

	// WHEN swap 100 of tokenA
	tokenIn, tokenOut := s.swap("TokenA", "TokenB", 100)

	// THEN swap should return 98 TokenA in and 36 TokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 98, tokenOut, 36)
	s.assertTickBalances(98, 64)

	s.assertCurr0To1(10_000)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap1To0FillLO() {
	// GIVEN 100 tokenA LO @ tick 10,000
	s.placeGTCLimitOrder("TokenA", 100, -10_000)

	// WHEN swap 10 of tokenB
	tokenIn, tokenOut := s.swap("TokenB", "TokenA", 10)

	// THEN swap should return 9 TokenB in and 3 TokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 9, tokenOut, 3)
	s.assertTickBalances(97, 9)

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(-10_000)
}

func (s *DexTestSuite) TestSwap0To1FillMultipleLO() {
	// GIVEN 300 tokenB across multiple LOs
	s.placeGTCLimitOrder("TokenB", 100, 1_000)
	s.placeGTCLimitOrder("TokenB", 100, 1_001)
	s.placeGTCLimitOrder("TokenB", 100, 1_002)

	// WHEN swap 300 of tokenA
	tokenIn, tokenOut := s.swap("TokenA", "TokenB", 300)

	// THEN swap should return 300 TokenA in and 270 TokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 300, tokenOut, 270)
	s.assertTickBalances(300, 30)

	s.assertCurr0To1(1_002)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap1To0FillMultipleLO() {
	// GIVEN 300 tokenA across multiple LOs
	s.placeGTCLimitOrder("TokenA", 100, -1_000)
	s.placeGTCLimitOrder("TokenA", 100, -1_001)
	s.placeGTCLimitOrder("TokenA", 100, -1_002)

	// WHEN swap 300 of tokenB
	tokenIn, tokenOut := s.swap("TokenB", "TokenA", 300)

	// THEN swap should return 300 TokenB in and 270 TokenB out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutput(tokenIn, 300, tokenOut, 270)
	s.assertTickBalances(30, 300)

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(-1_002)
}

func (s *DexTestSuite) TestSwap0To1LOMaxAmountUsed() {
	// GIVEN 10 TokenB available
	s.placeGTCLimitOrder("TokenB", 10, 1)

	// WHEN swap 50 TokenA with maxOut of 5
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 50, 5)

	// THEN swap should return 6 TokenA in and 5 TokenB out
	s.assertSwapOutput(tokenIn, 6, tokenOut, 5)
	s.assertTickBalances(6, 5)
}

func (s *DexTestSuite) TestSwap1To0LOMaxAmountUsed() {
	// GIVEN 10 TokenA available
	s.placeGTCLimitOrder("TokenA", 10, 0)

	// WHEN swap 50 TokenB with maxOut of 5
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 50, 5)

	// THEN swap should return 5 TokenB in and 5 TokenA out
	s.assertSwapOutput(tokenIn, 5, tokenOut, 5)
	s.assertTickBalances(5, 5)
}

func (s *DexTestSuite) TestSwap0To1LOMaxAmountNotUsed() {
	// GIVEN 10 TokenB available
	s.placeGTCLimitOrder("TokenB", 10, 1)

	// WHEN swap 8 with maxOut of 15
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 8, 15)

	// THEN swap should return 8 TokenA in and 7 TokenB out
	s.assertSwapOutput(tokenIn, 8, tokenOut, 7)
	s.assertTickBalances(8, 3)
}

func (s *DexTestSuite) TestSwap1To0LOMaxAmountNotUsed() {
	// GIVEN 10 TokenA available
	s.placeGTCLimitOrder("TokenA", 10, 1)

	// WHEN swap 8 with maxOut of 15
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 8, 15)

	// THEN swap should return 8 TokenB in and 8 TokenA out
	s.assertSwapOutput(tokenIn, 8, tokenOut, 8)
	s.assertTickBalances(2, 8)
}

func (s *DexTestSuite) TestSwap0To1LOMaxAmountUsedMultiTick() {
	// GIVEN 50 TokenB available
	s.placeGTCLimitOrder("TokenB", 5, 0)
	s.placeGTCLimitOrder("TokenB", 5, 1)
	s.placeGTCLimitOrder("TokenB", 5, 2)
	s.placeGTCLimitOrder("TokenB", 5, 3)
	s.placeGTCLimitOrder("TokenB", 30, 4)

	// WHEN swap 50 TokenA with maxOut of 20
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 50, 20)

	// THEN swap should return 23 TokenA in and 20 TokenB out
	s.assertSwapOutput(tokenIn, 23, tokenOut, 20)
	s.assertTickBalances(23, 30)
}

func (s *DexTestSuite) TestSwap1To0LOMaxAmountUsedMultiTick() {
	// GIVEN 50 TokenA available
	s.placeGTCLimitOrder("TokenA", 5, 0)
	s.placeGTCLimitOrder("TokenA", 5, 1)
	s.placeGTCLimitOrder("TokenA", 5, 2)
	s.placeGTCLimitOrder("TokenA", 5, 3)
	s.placeGTCLimitOrder("TokenA", 30, 4)

	// WHEN swap 50 TokenB with maxOut of 20
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 50, 20)

	// THEN swap should return 20 TokenB in and 20 TokenA out
	s.assertSwapOutput(tokenIn, 20, tokenOut, 20)
	s.assertTickBalances(30, 20)
}

func (s *DexTestSuite) TestSwap0To1LOMaxAmountNotUsedMultiTick() {
	// GIVEN 50 TokenB available
	s.placeGTCLimitOrder("TokenB", 5, 0)
	s.placeGTCLimitOrder("TokenB", 5, 1)
	s.placeGTCLimitOrder("TokenB", 5, 2)
	s.placeGTCLimitOrder("TokenB", 5, 3)
	s.placeGTCLimitOrder("TokenB", 30, 4)

	// WHEN swap 19 TokenA with maxOut of 20
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 19, 20)

	// THEN swap should return 19 TokenA in and 16 TokenB out
	s.assertSwapOutput(tokenIn, 19, tokenOut, 16)
	s.assertTickBalances(19, 34)
}

// Swap LO and LP  ////////////////////////////////////////////////////////////

func (s *DexTestSuite) TestSwapExhaustsLOAndLP() {
	s.placeGTCLimitOrder("TokenB", 10, 0)

	s.addDeposits(NewDeposit(0, 10, 0, 1))

	s.swapWithMaxOut("TokenA", "TokenB", 19, 20)

	// There should be total of 6 tick updates
	// (limitOrder, 2x deposit,  2x swap LP, swap LO)
	s.AssertNEventValuesEmitted(types.TickUpdateEventKey, 6)
}

// Test helpers ///////////////////////////////////////////////////////////////

func (s *DexTestSuite) addDeposit(deposit *Deposit) {
	pool, err := s.App.DexKeeper.GetOrInitPool(s.Ctx, defaultPairID, deposit.TickIndex, deposit.Fee)
	s.Assert().NoError(err)
	pool.LowerTick0.ReservesMakerDenom = pool.LowerTick0.ReservesMakerDenom.Add(deposit.AmountA)
	pool.UpperTick1.ReservesMakerDenom = pool.UpperTick1.ReservesMakerDenom.Add(deposit.AmountB)
	s.App.DexKeeper.SetPool(s.Ctx, pool)
}

func (s *DexTestSuite) addDeposits(deposits ...*Deposit) {
	for _, deposit := range deposits {
		s.addDeposit(deposit)
	}
}

func (s *DexTestSuite) placeGTCLimitOrder(
	makerDenom string,
	amountIn int64,
	tickIndex int64,
) {
	tradePairID := defaultPairID.MustTradePairIDFromMaker(makerDenom)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(tickIndex)
	tranche, err := s.App.DexKeeper.GetOrInitPlaceTranche(
		s.Ctx,
		tradePairID,
		tickIndexTakerToMaker,
		nil,
		types.LimitOrderType_GOOD_TIL_CANCELLED,
	)
	s.Assert().NoError(err)
	tranche.PlaceMakerLimitOrder(sdkmath.NewInt(amountIn))
	s.App.DexKeeper.SaveTranche(s.Ctx, tranche)
}

func (s *DexTestSuite) swap(
	tokenIn string,
	tokenOut string,
	maxAmountIn int64,
) (coinIn, coinOut sdk.Coin) {
	tradePairID, err := types.NewTradePairID(tokenIn, tokenOut)
	s.Assert().NoError(err)
	coinIn, coinOut, _ = s.App.DexKeeper.Swap(
		s.Ctx,
		tradePairID,
		sdkmath.NewInt(maxAmountIn),
		nil,
		nil,
	)
	return coinIn, coinOut
}

func (s *DexTestSuite) swapWithMaxOut(
	tokenIn string,
	tokenOut string,
	maxAmountIn int64,
	maxAmountOut int64,
) (coinIn, coinOut sdk.Coin) {
	tradePairID := types.MustNewTradePairID(tokenIn, tokenOut)
	maxAmountOutInt := sdkmath.NewInt(maxAmountOut)
	coinIn, coinOut, _ = s.App.DexKeeper.Swap(
		s.Ctx,
		tradePairID,
		sdkmath.NewInt(maxAmountIn),
		&maxAmountOutInt,
		nil,
	)

	return coinIn, coinOut
}

func (s *DexTestSuite) assertSwapOutput(
	actualIn sdk.Coin,
	expectedIn int64,
	actualOut sdk.Coin,
	expectedOut int64,
) {
	amtIn := actualIn.Amount
	amtOut := actualOut.Amount

	s.Assert().
		True(amtIn.Equal(sdkmath.NewInt(expectedIn)), "Expected amountIn %d != %s", expectedIn, amtIn)
	s.Assert().
		True(amtOut.Equal(sdkmath.NewInt(expectedOut)), "Expected amountOut %d != %s", expectedOut, amtOut)
}

func (s *DexTestSuite) assertTickBalances(expectedABalance, expectedBBalance int64) {
	// NOTE: We can't just check the actual DEX bank balances since we are testing swap
	// before any transfers take place. Instead we have to sum up the total amount of coins
	// at each tick
	expectedAInt := sdkmath.NewInt(expectedABalance)
	expectedBInt := sdkmath.NewInt(expectedBBalance)
	allCoins := sdk.Coins{}
	ticks := s.App.DexKeeper.GetAllTickLiquidity(s.Ctx)
	inactiveLOs := s.App.DexKeeper.GetAllInactiveLimitOrderTranche(s.Ctx)

	for _, tick := range ticks {
		switch liquidity := tick.Liquidity.(type) {
		case *types.TickLiquidity_LimitOrderTranche:
			tokenIn := liquidity.LimitOrderTranche.Key.TradePairID.MakerDenom
			amountIn := liquidity.LimitOrderTranche.ReservesMakerDenom
			allCoins = allCoins.Add(sdk.NewCoin(tokenIn, amountIn))

			tokenOut := liquidity.LimitOrderTranche.Key.TradePairID.TakerDenom
			amountOut := liquidity.LimitOrderTranche.ReservesTakerDenom
			allCoins = allCoins.Add(sdk.NewCoin(tokenOut, amountOut))

		case *types.TickLiquidity_PoolReserves:
			tokenIn := liquidity.PoolReserves.Key.TradePairID.MakerDenom
			reserves := liquidity.PoolReserves.ReservesMakerDenom
			allCoins = allCoins.Add(sdk.NewCoin(tokenIn, reserves))
		}
	}

	for _, lo := range inactiveLOs {
		tokenOut := lo.Key.TradePairID.TakerDenom
		amountOut := lo.ReservesTakerDenom
		allCoins = allCoins.Add(sdk.NewCoin(tokenOut, amountOut))
	}

	actualA := allCoins.AmountOf("TokenA")
	actualB := allCoins.AmountOf("TokenB")

	s.Assert().
		True(actualA.Equal(expectedAInt), "TokenA: expected %s != actual %s", expectedAInt, actualA)
	s.Assert().
		True(actualB.Equal(expectedBInt), "TokenB: expected %s != actual %s", expectedBInt, actualB)
}
