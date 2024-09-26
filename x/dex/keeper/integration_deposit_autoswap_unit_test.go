package keeper_test

import (
	"cosmossdk.io/math"
)

func (s *DexTestSuite) TestAutoswapSingleSided0To1() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 0)

	// GIVEN a pool with double-sided liquidity
	s.aliceDeposits(NewDeposit(50, 50, 2000, 2))
	s.assertAccountSharesInt(s.alice, 2000, 2, math.NewInt(111069527))

	// WHEN bob deposits only TokenA
	s.bobDeposits(NewDeposit(50, 0, 2000, 2))
	s.assertPoolLiquidity(100, 50, 2000, 2)

	// THEN his deposit is autoswapped
	// He receives 49.985501 shares
	// depositValue = depositAmount - (autoswapedAmountAsToken0 * fee)
	//              = 50 - 50 * (1 - 1.0001^-2)
	//              = 49.9900014998
	// SharesIssued = depositValue * existing shares / (existingValue + autoSwapFee)
	//              = 49.9900014998 * 111.069527 / (111.069527 + 0.0099985002)
	//              = 49.985501

	s.assertAccountSharesInt(s.bob, 2000, 2, math.NewInt(49985501))
}

func (s *DexTestSuite) TestAutoswapSingleSided1To0() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(0, 50)

	// GIVEN a pool with double-sided liquidity
	s.aliceDeposits(NewDeposit(50, 50, 2000, 2))
	s.assertAccountSharesInt(s.alice, 2000, 2, math.NewInt(111069527))

	// WHEN bob deposits only TokenB
	s.bobDeposits(NewDeposit(0, 50, 2000, 2))
	s.assertPoolLiquidity(50, 100, 2000, 2)

	// THEN his deposit is autoswapped
	// He receives 61.0 shares
	// depositAmountAsToken0 = 50 * 1.0001^2000 = 61.06952725039
	// depositValue = depositAmountAsToken0 - (autoswapedAmountAsToken0 * fee)
	//              = 61.06952725039 - 61.06952725039 * (1 - 1.0001^-2)
	//              = 61.05731517678
	// SharesIssued = depositValue * existing shares / (existingValue + autoSwapFee)
	//              = 61.05731517678 * 111.069527 / (111.069527 + 0.01221207361)
	//              = 61.050602

	s.assertAccountSharesInt(s.bob, 2000, 2, math.NewInt(61050602))
}

func (s *DexTestSuite) TestAutoswapDoubleSided0To1() {
	s.fundAliceBalances(30, 50)
	s.fundBobBalances(50, 50)

	// GIVEN a pool with double-sided liquidity
	s.aliceDeposits(NewDeposit(30, 50, -4000, 10))
	s.assertAccountSharesInt(s.alice, -4000, 10, math.NewInt(63516672))

	// WHEN bob deposits a ratio of 1:1 tokenA and B
	s.bobDeposits(NewDeposit(50, 50, -4000, 10))
	s.assertPoolLiquidity(80, 100, -4000, 10)

	// THEN his deposit is autoswapped
	// He receives 83.5 shares
	// depositValue = depositAmountAsToken0 - (autoswapedAmountAsToken0 * fee)
	//              = 83.5166725838 -  20 * (1 - 1.0001^-10)
	//              = 83.4966835794
	// SharesIssued = depositValue * existing shares / (existingValue + autoSwapFee)
	//              = 83.4966835794 * 63.516672 / (63.5166725838 + 0.0199890044)
	//              = 83.470414

	s.assertAccountSharesInt(s.bob, -4000, 10, math.NewInt(83470414))
}

func (s *DexTestSuite) TestAutoswapDoubleSided1To0() {
	s.fundAliceBalances(50, 30)
	s.fundBobBalances(50, 50)

	// GIVEN a pool with double-sided liquidity
	s.aliceDeposits(NewDeposit(50, 30, -4000, 10))
	s.assertAccountSharesInt(s.alice, -4000, 10, math.NewInt(70110003))

	// WHEN bob deposits a ratio of 1:1 tokenA and B
	s.bobDeposits(NewDeposit(50, 50, -4000, 10))
	s.assertPoolLiquidity(100, 80, -4000, 10)

	// THEN his deposit is autoswapped
	// He receives 83.5 shares
	// depositValue = depositAmountAsToken0 - (autoswapedAmountAsToken0 * fee)
	//              = 83.5166725838 -  13.4066690335 * (1 - 1.0001^-10)
	//              = 83.5032732855
	// SharesIssued = depositValue * existing shares / (existingValue + autoSwapFee)
	//              = 83.5032732855 * 70.110003 / (70.1100035503 + 0.01339929831)
	//              = 83.487316

	s.assertAccountSharesInt(s.bob, -4000, 10, math.NewInt(83487316))
}
