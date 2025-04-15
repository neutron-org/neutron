package keeper_test

import (
	"math"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// TODO: In an ideal world, there should be enough lower level testing that the swap tests
// don't need to test both LO and LP. At the level of swap testing these should be indistinguishable.
func (s *DexTestSuite) TestSwap0To1NoLiquidity() {
	// GIVEN no liqudity of token B (deposit only token A and LO of token A)
	s.addDeposit(NewDeposit(10, 0, 0, 1))
	s.placeGTCLimitOrder("TokenA", 1000, 10)

	// WHEN swap 10 of tokenB
	tokenIn, tokenOut := s.swapSuccess("TokenA", "TokenB", 10)

	// THEN swap should do nothing
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(0o_000_000), tokenOut, sdkmath.NewInt(0o_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(1010_000_000), sdkmath.NewInt(0o_000_000))

	s.assertCurr0To1(math.MaxInt64)
}

func (s *DexTestSuite) TestSwap1To0NoLiquidity() {
	// GIVEN no liqudity of token A (deposit only token B and LO of token B)
	s.addDeposit(NewDeposit(0, 10, 0, 1))
	s.placeGTCLimitOrder("TokenB", 1000, 10)

	// WHEN swap 10 of tokenB
	tokenIn, tokenOut := s.swapSuccess("TokenB", "TokenA", 10)

	// THEN swap should do nothing
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(0o_000_000), tokenOut, sdkmath.NewInt(0o_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(0o_000_000), sdkmath.NewInt(1010_000_000))

	s.assertCurr1To0(math.MinInt64)
}

// swaps against LPs only /////////////////////////////////////////////////////

func (s *DexTestSuite) TestSwap0To1PartialFillLP() {
	// GIVEN 10 tokenB LP @ tick 0 fee 1
	s.addDeposit(NewDeposit(0, 10, 0, 1))

	// WHEN swap 20 of tokenA
	tokenIn, tokenOut := s.swapSuccess("TokenA", "TokenB", 20)

	// THEN swap should return ~10 BIGTokenA in and 10 BIGTokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(10_001_000), tokenOut, sdkmath.NewInt(10_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(10_001_000), sdkmath.ZeroInt())

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(-1)
}

func (s *DexTestSuite) TestSwap1To0PartialFillLP() {
	// GIVEN 10 tokenA LP @ tick 0 fee 1
	s.addDeposit(NewDeposit(10, 0, 0, 1))

	// WHEN swap 20 of tokenB
	tokenIn, tokenOut := s.swapSuccess("TokenB", "TokenA", 20)

	// THEN swap should return ~10 BIGTokenB in and 10 BIGTokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(10_001_000), tokenOut, sdkmath.NewInt(10_000_000))
	s.assertTickBalancesInt(sdkmath.ZeroInt(), sdkmath.NewInt(10_001_000))

	s.assertCurr0To1(1)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap0To1FillLP() {
	// GIVEN 100 tokenB LP @ tick 200 fee 5
	s.addDeposit(NewDeposit(0, 100, 200, 5))

	// WHEN swap 100 of tokenA
	tokenIn, tokenOut := s.swapSuccess("TokenA", "TokenB", 100)

	// THEN swap should return 100 BIGTokenA in and ~98 BIGTokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(100000000), tokenOut, sdkmath.NewInt(97_970_970))
	s.assertTickBalancesInt(sdkmath.NewInt(100000000), sdkmath.NewInt(20_29_030))

	s.assertCurr0To1(205)
	s.assertCurr1To0(195)
}

func (s *DexTestSuite) TestSwap1To0FillLP() {
	// GIVEN 100 tokenA LP @ tick -20,000 fee 1
	s.addDeposit(NewDeposit(100, 0, -20_000, 1))

	// WHEN swap 100 of tokenB
	tokenIn, tokenOut := s.swapSuccess("TokenB", "TokenA", 100)

	// THEN swap should return ~99 BIGTokenB in and ~14 BIGTokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	// NOTE: Given rounding for amountOut, amountIn does not use the full maxAmount
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(99_999_998), tokenOut, sdkmath.NewInt(13_533_528))
	s.assertTickBalancesInt(sdkmath.NewInt(86_466_472), sdkmath.NewInt(99_999_998))

	s.assertCurr0To1(-19_999)
	s.assertCurr1To0(-20_001)
}

func (s *DexTestSuite) TestSwap0To1FillLPHighFee() {
	// GIVEN 100 tokenB LP @ tick 20,000 fee 1,000
	s.addDeposit(NewDeposit(0, 100, 20_000, 1_000))

	// WHEN swap 100 of tokenA
	tokenIn, tokenOut := s.swapSuccess("TokenA", "TokenB", 100)

	// THEN swap should return ~99 BIGTokenA in and ~12 BIGTokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(99_999_996), tokenOut, sdkmath.NewInt(12_246_928))
	s.assertTickBalancesInt(sdkmath.NewInt(99_999_996), sdkmath.NewInt(87_753_072))

	s.assertCurr0To1(21_000)
	s.assertCurr1To0(19_000)
}

func (s *DexTestSuite) TestSwap1To0FillLPHighFee() {
	// GIVEN 1000 tokenA LP @ tick 20,000 fee 1000
	s.addDeposit(NewDeposit(1000, 0, 20_000, 1000))

	// WHEN swap 100 of tokenB
	tokenIn, tokenOut := s.swapSuccess("TokenB", "TokenA", 100)

	// THEN swap should return 100 BIGTokenB in and ~668 BIGTokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(100_000_000), tokenOut, sdkmath.NewInt(668_525_935))
	s.assertTickBalancesInt(sdkmath.NewInt(331_474_065), sdkmath.NewInt(100_000_000))

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
	tokenIn, tokenOut := s.swapSuccess("TokenA", "TokenB", 100)

	// THEN swap should return ~40 BIGTokenA in and 300 TokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(40_604_647), tokenOut, sdkmath.NewInt(300_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(40_604_647), sdkmath.ZeroInt())

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
	tokenIn, tokenOut := s.swapSuccess("TokenB", "TokenA", 100)

	// THEN swap should return ~41 BIGTokenB in and 300 BIGTokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(40604647), tokenOut, sdkmath.NewInt(300000000))
	s.assertTickBalancesInt(sdkmath.ZeroInt(), sdkmath.NewInt(40604647))

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
	tokenIn, tokenOut := s.swapSuccess("TokenA", "TokenB", 400)

	// THEN swap should return ~399 BIGTokenA in and 400 BIGTokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(399_180_884), tokenOut, sdkmath.NewInt(400_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(399_180_884), sdkmath.ZeroInt())

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
	tokenIn, tokenOut := s.swapSuccess("TokenB", "TokenA", 400)

	// THEN swap should return 400 BIGTokenB in and ~400 BIGTokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(399_180_884), tokenOut, sdkmath.NewInt(400_000_000))
	s.assertTickBalancesInt(sdkmath.ZeroInt(), sdkmath.NewInt(399_180_884))

	s.assertCurr0To1(21)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap0To1LPMaxAmountUsed() {
	// GIVEN 10 TokenB available
	s.addDeposits(NewDeposit(0, 10, 0, 1))

	// WHEN swap 50 TokenA with maxOut of 5
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 50, 5)

	// THEN swap should return ~5 BIGTokenA in and 5 BIGTokenB out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(5_000_500), tokenOut, sdkmath.NewInt(5_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(5_000_500), sdkmath.NewInt(5_000_000))
}

func (s *DexTestSuite) TestSwap1To0LPMaxAmountUsed() {
	// GIVEN 10 TokenA available
	s.addDeposits(NewDeposit(10, 0, 0, 1))

	// WHEN swap 50 TokenB with maxOut of 5
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 50, 5)

	// THEN swap should return ~5 BIGTokenB in and 5 BIGTokenA out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(5_000_500), tokenOut, sdkmath.NewInt(5_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(5_000_000), sdkmath.NewInt(5_000_500))
}

func (s *DexTestSuite) TestSwap0To1LPMaxAmountNotUsed() {
	// GIVEN 10 TokenB available
	s.addDeposits(NewDeposit(0, 10, 0, 1))

	// WHEN swap 8 with maxOut of 15
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 8, 15)

	// THEN swap should return 8 BIGTokenA in and ~8 BIGTokenB out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(8_000_000), tokenOut, sdkmath.NewInt(7_999_200))
	s.assertTickBalancesInt(sdkmath.NewInt(8_000_000), sdkmath.NewInt(20_00_800))
}

func (s *DexTestSuite) TestSwap1To0LPMaxAmountNotUsed() {
	// GIVEN 10 TokenA available
	s.addDeposits(NewDeposit(10, 0, 0, 1))

	// WHEN swap 8 with maxOut of 15
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 8, 15)

	// THEN swap should return 8 BIGTokenB in and ~8 BIGTokenA out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(8_000_000), tokenOut, sdkmath.NewInt(7_999_200))
	s.assertTickBalancesInt(sdkmath.NewInt(2000800), sdkmath.NewInt(8_000_000))
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

	// THEN swap should return ~20 BIGTokenA in and 20 BIGTokenB out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(20_005_003), tokenOut, sdkmath.NewInt(20_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(20_005_003), sdkmath.NewInt(30_000_000))
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

	// THEN swap should return ~ 20 BIGTokenB in and 20 BIGTokenA out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(19_994_002), tokenOut, sdkmath.NewInt(20_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(30_000_000), sdkmath.NewInt(19_994_002))
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

	// THEN swap should return 19 BIGTokenA in and 19 BIGTokenB out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(19_000_000), tokenOut, sdkmath.NewInt(18_995_399))
	s.assertTickBalancesInt(sdkmath.NewInt(19_000_000), sdkmath.NewInt(31_004_601))
}

// swaps against LOs only /////////////////////////////////////////////////////

func (s *DexTestSuite) TestSwap0To1PartialFillLO() {
	// GIVEN 10 tokenB LO @ tick 1,000
	s.placeGTCLimitOrder("TokenB", 10, 1_000)

	// WHEN swap 20 of tokenA
	tokenIn, tokenOut := s.swapSuccess("TokenA", "TokenB", 20)

	// THEN swap should return ~11 BIGTokenA in and 10 BIGTokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(11_051_654), tokenOut, sdkmath.NewInt(10_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(11_051_654), sdkmath.ZeroInt())
}

func (s *DexTestSuite) TestSwap1To0PartialFillLO() {
	// GIVEN 10 tokenA LO @ tick -1,000
	s.placeGTCLimitOrder("TokenA", 10, -1_000)

	// WHEN swap 20 of tokenB
	tokenIn, tokenOut := s.swapSuccess("TokenB", "TokenA", 20)

	// THEN swap should return ~11 BIGTokenB in and 10 BIGTokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(11_051_654), tokenOut, sdkmath.NewInt(10_000_000))
	s.assertTickBalancesInt(sdkmath.ZeroInt(), sdkmath.NewInt(11_051_654))

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap0To1FillLO() {
	// GIVEN 100 tokenB LO @ tick 10,000
	s.placeGTCLimitOrder("TokenB", 100, 10_000)

	// WHEN swap 100 of tokenA
	tokenIn, tokenOut := s.swapSuccess("TokenA", "TokenB", 100)

	// THEN swap should return ~99 BIGTokenA in and ~37 BIGTokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(99_999_999), tokenOut, sdkmath.NewInt(36_789_783))
	s.assertTickBalancesInt(sdkmath.NewInt(99_999_999), sdkmath.NewInt(63_210_217))

	s.assertCurr0To1(10_000)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap1To0FillLO() {
	// GIVEN 100 tokenA LO @ tick 10,000
	s.placeGTCLimitOrder("TokenA", 100, -10_000)

	// WHEN swap 10 of tokenB
	tokenIn, tokenOut := s.swapSuccess("TokenB", "TokenA", 10)

	// THEN swap should return 10 BIGTokenB in and ~4 BIGTokenA out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(10_000_000), tokenOut, sdkmath.NewInt(3_678_978))
	s.assertTickBalancesInt(sdkmath.NewInt(96_321_022), sdkmath.NewInt(10_000_000))

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(-10_000)
}

func (s *DexTestSuite) TestSwap0To1FillMultipleLO() {
	// GIVEN 300 tokenB across multiple LOs
	s.placeGTCLimitOrder("TokenB", 100, 1_000)
	s.placeGTCLimitOrder("TokenB", 100, 1_001)
	s.placeGTCLimitOrder("TokenB", 100, 1_002)

	// WHEN swap 300 of tokenA
	tokenIn, tokenOut := s.swapSuccess("TokenA", "TokenB", 300)

	// THEN swap should return 300 BIGTokenA in and ~271 BIGTokenB out
	s.Assert().Equal("TokenA", tokenIn.Denom)
	s.Assert().Equal("TokenB", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(300_000_000), tokenOut, sdkmath.NewInt(271_428_295))
	s.assertTickBalancesInt(sdkmath.NewInt(300_000_000), sdkmath.NewInt(28_571_705))

	s.assertCurr0To1(1_002)
	s.assertCurr1To0(math.MinInt64)
}

func (s *DexTestSuite) TestSwap1To0FillMultipleLO() {
	// GIVEN 300 tokenA across multiple LOs
	s.placeGTCLimitOrder("TokenA", 100, -1_000)
	s.placeGTCLimitOrder("TokenA", 100, -1_001)
	s.placeGTCLimitOrder("TokenA", 100, -1_002)

	// WHEN swap 300 of tokenB
	tokenIn, tokenOut := s.swapSuccess("TokenB", "TokenA", 300)

	// THEN swap should return 300 BIGTokenB in and ~271 BIGTokenB out
	s.Assert().Equal("TokenB", tokenIn.Denom)
	s.Assert().Equal("TokenA", tokenOut.Denom)
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(300_000_000), tokenOut, sdkmath.NewInt(271_428_295))
	s.assertTickBalancesInt(sdkmath.NewInt(28_571_705), sdkmath.NewInt(300_000_000))

	s.assertCurr0To1(math.MaxInt64)
	s.assertCurr1To0(-1_002)
}

func (s *DexTestSuite) TestSwap0To1LOMaxAmountUsed() {
	// GIVEN 10 TokenB available
	s.placeGTCLimitOrder("TokenB", 10, 1)

	// WHEN swap 50 TokenA with maxOut of 5
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 50, 5)

	// THEN swap should return ~5 BIGTokenA in and 5 BIGTokenB out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(5_000_500), tokenOut, sdkmath.NewInt(5_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(5_000_500), sdkmath.NewInt(5_000_000))
}

func (s *DexTestSuite) TestSwap1To0LOMaxAmountUsed() {
	// GIVEN 10 TokenA available
	s.placeGTCLimitOrder("TokenA", 10, 0)

	// WHEN swap 50 TokenB with maxOut of 5
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 50, 5)

	// THEN swap should return 5 TokenB in and 5 TokenA out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(5_000_000), tokenOut, sdkmath.NewInt(5_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(5_000_000), sdkmath.NewInt(5_000_000))
}

func (s *DexTestSuite) TestSwap0To1LOMaxAmountNotUsed() {
	// GIVEN 10 TokenB available
	s.placeGTCLimitOrder("TokenB", 10, 1)

	// WHEN swap 8 with maxOut of 15
	tokenIn, tokenOut := s.swapWithMaxOut("TokenA", "TokenB", 8, 15)

	// THEN swap should return 8 BIGTokenA in and ~8 BIGTokenB out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(8_000_000), tokenOut, sdkmath.NewInt(7_999_200))
	s.assertTickBalancesInt(sdkmath.NewInt(8_000_000), sdkmath.NewInt(2_000_800))
}

func (s *DexTestSuite) TestSwap1To0LOMaxAmountNotUsed() {
	// GIVEN 10 TokenA available
	s.placeGTCLimitOrder("TokenA", 10, 1)

	// WHEN swap 8 with maxOut of 15
	tokenIn, tokenOut := s.swapWithMaxOut("TokenB", "TokenA", 8, 15)

	// THEN swap should return 8 BIGTokenB in and ~8 BIGTokenA out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(8_000_000), tokenOut, sdkmath.NewInt(8_000_799))
	s.assertTickBalancesInt(sdkmath.NewInt(1_999_201), sdkmath.NewInt(8_000_000))
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

	// THEN swap should return ~20 BIGTokenA in and 20 TokenB out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(20_003_002), tokenOut, sdkmath.NewInt(20_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(20_003_002), sdkmath.NewInt(30_000_000))
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

	// THEN swap should return ~20 BIGTokenB in and 20 BIGTokenA out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(19_992_002), tokenOut, sdkmath.NewInt(20_000_000))
	s.assertTickBalancesInt(sdkmath.NewInt(30_000_000), sdkmath.NewInt(19_992_002))
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

	// THEN swap should return 19 BIGTokenA in and ~19 BIGTokenB out
	s.assertSwapOutputInt(tokenIn, sdkmath.NewInt(19_000_000), tokenOut, sdkmath.NewInt(18_997_299))
	s.assertTickBalancesInt(sdkmath.NewInt(19_000_000), sdkmath.NewInt(31_002_701))
}

// Swap LO and LP  ////////////////////////////////////////////////////////////

func (s *DexTestSuite) TestSwapExhaustsLOAndLP() {
	s.placeGTCLimitOrder("TokenB", 10, 0)

	s.addDeposits(NewDeposit(0, 10, 0, 1))

	s.swapWithMaxOut("TokenA", "TokenB", 19, 20)

	// There should be total of 6 tick updates
	// (limitOrder, 2x deposit,  2x swap LP, swap LO)
	s.AssertNEventValuesEmitted(types.TickUpdateEventKey, 6)

	tickUpdates := s.GetAllMatchingEvents(types.TickUpdateEventKey)

	// LimitOrder TickUpdate has correct SwapMetadatrra
	loTickUpdate := tickUpdates[3]
	loSwapIn, _ := loTickUpdate.GetAttribute(types.AttributeSwapAmountIn)
	loSwapOut, _ := loTickUpdate.GetAttribute(types.AttributeSwapAmountOut)
	s.Equal("10000000", loSwapIn.Value)
	s.Equal("10000000", loSwapOut.Value)

	// LP TickUpdate has correct SwapMetadata
	lpTickUpdate := tickUpdates[5]
	lpSwapIn, _ := lpTickUpdate.GetAttribute(types.AttributeSwapAmountIn)
	lpSwapOut, _ := lpTickUpdate.GetAttribute(types.AttributeSwapAmountOut)
	s.Equal("9000000", lpSwapIn.Value)
	s.Equal("8999100", lpSwapOut.Value)

	// opposite LP TickUpdate has no SwapMetadata
	lpTickUpdateOppositeTick := tickUpdates[4]
	_, found := lpTickUpdateOppositeTick.GetAttribute(types.AttributeSwapAmountIn)
	s.False(found)
	_, found = lpTickUpdateOppositeTick.GetAttribute(types.AttributeSwapAmountOut)
	s.False(found)
}

// Test helpers ///////////////////////////////////////////////////////////////

func (s *DexTestSuite) addDeposit(deposit *Deposit) {
	pool, err := s.App.DexKeeper.GetOrInitPool(s.Ctx, defaultPairID, deposit.TickIndex, deposit.Fee)
	s.Assert().NoError(err)
	pool.LowerTick0.ReservesMakerDenom = pool.LowerTick0.ReservesMakerDenom.Add(deposit.AmountA)
	pool.UpperTick1.ReservesMakerDenom = pool.UpperTick1.ReservesMakerDenom.Add(deposit.AmountB)
	s.App.DexKeeper.UpdatePool(s.Ctx, pool)
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
	tranche.PlaceMakerLimitOrder(sdkmath.NewInt(amountIn).Mul(denomMultiple))
	s.App.DexKeeper.UpdateTranche(s.Ctx, tranche)
}

func (s *DexTestSuite) swapInt(
	tokenIn string,
	tokenOut string,
	maxAmountIn sdkmath.Int,
) (coinIn, coinOut sdk.Coin, filled bool, err error) {
	tradePairID, err := types.NewTradePairID(tokenIn, tokenOut)
	s.Assert().NoError(err)
	return s.App.DexKeeper.Swap(
		s.Ctx,
		tradePairID,
		maxAmountIn,
		nil,
		nil,
	)
}

func (s *DexTestSuite) swapSuccess(
	tokenIn string,
	tokenOut string,
	maxAmountIn int64,
) (coinIn, coinOut sdk.Coin) {
	coinIn, coinOut, _, err := s.swapInt(tokenIn, tokenOut, sdkmath.NewInt(maxAmountIn).Mul(denomMultiple))
	s.Assert().NoError(err)
	return coinIn, coinOut
}

func (s *DexTestSuite) swapWithMaxOut(
	tokenIn string,
	tokenOut string,
	maxAmountIn int64,
	maxAmountOut int64,
) (coinIn, coinOut sdk.Coin) {
	tradePairID := types.MustNewTradePairID(tokenIn, tokenOut)
	maxAmountOutInt := sdkmath.NewInt(maxAmountOut).Mul(denomMultiple)
	coinIn, coinOut, _, err := s.App.DexKeeper.Swap(
		s.Ctx,
		tradePairID,
		sdkmath.NewInt(maxAmountIn).Mul(denomMultiple),
		&maxAmountOutInt,
		nil,
	)
	s.Assert().NoError(err)

	return coinIn, coinOut
}

func (s *DexTestSuite) assertSwapOutputInt(
	actualIn sdk.Coin,
	expectedIn sdkmath.Int,
	actualOut sdk.Coin,
	expectedOut sdkmath.Int,
) {
	amtIn := actualIn.Amount
	amtOut := actualOut.Amount

	s.Assert().
		True(amtIn.Equal(expectedIn), "Expected amountIn %s != %s", expectedIn, amtIn)
	s.Assert().
		True(amtOut.Equal(expectedOut), "Expected amountOut %s != %s", expectedOut, amtOut)
}

//nolint:unused
func (s *DexTestSuite) assertSwapOutput(
	actualIn sdk.Coin,
	expectedIn int64,
	actualOut sdk.Coin,
	expectedOut int64,
) {
	expectedInInt := sdkmath.NewInt(expectedIn).Mul(denomMultiple)
	expectedOutInt := sdkmath.NewInt(expectedOut).Mul(denomMultiple)
	s.assertSwapOutputInt(actualIn, expectedInInt, actualOut, expectedOutInt)
}

func (s *DexTestSuite) assertTickBalancesInt(expectedABalance, expectedBBalance sdkmath.Int) {
	// NOTE: We can't just check the actual DEX bank balances since we are testing swap
	// before any transfers take place. Instead we have to sum up the total amount of coins
	// at each tick
	allCoins := sdk.Coins{}
	ticks := s.App.DexKeeper.GetAllTickLiquidity(s.Ctx)
	inactiveLOs := s.App.DexKeeper.GetAllInactiveLimitOrderTranche(s.Ctx)

	for _, tick := range ticks {
		switch liquidity := tick.Liquidity.(type) {
		case *types.TickLiquidity_LimitOrderTranche:
			tokenIn := liquidity.LimitOrderTranche.Key.TradePairId.MakerDenom
			amountIn := liquidity.LimitOrderTranche.ReservesMakerDenom
			allCoins = allCoins.Add(sdk.NewCoin(tokenIn, amountIn))

			tokenOut := liquidity.LimitOrderTranche.Key.TradePairId.TakerDenom
			amountOut := liquidity.LimitOrderTranche.ReservesTakerDenom
			allCoins = allCoins.Add(sdk.NewCoin(tokenOut, amountOut))

		case *types.TickLiquidity_PoolReserves:
			tokenIn := liquidity.PoolReserves.Key.TradePairId.MakerDenom
			reserves := liquidity.PoolReserves.ReservesMakerDenom
			allCoins = allCoins.Add(sdk.NewCoin(tokenIn, reserves))
		}
	}

	for _, lo := range inactiveLOs {
		tokenOut := lo.Key.TradePairId.TakerDenom
		amountOut := lo.ReservesTakerDenom
		allCoins = allCoins.Add(sdk.NewCoin(tokenOut, amountOut))
	}

	actualA := allCoins.AmountOf("TokenA")
	actualB := allCoins.AmountOf("TokenB")

	s.Assert().
		True(actualA.Equal(expectedABalance), "TokenA: expected %s != actual %s", expectedABalance, actualA)
	s.Assert().
		True(actualB.Equal(expectedBBalance), "TokenB: expected %s != actual %s", expectedBBalance, actualB)
}

//nolint:unused
func (s *DexTestSuite) assertTickBalances(expectedABalance, expectedBBalance int64) {
	expectedAInt := sdkmath.NewInt(expectedABalance).Mul(denomMultiple)
	expectedBInt := sdkmath.NewInt(expectedBBalance).Mul(denomMultiple)
	s.assertTickBalancesInt(expectedAInt, expectedBInt)
}
