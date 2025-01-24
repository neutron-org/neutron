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
	// swapAmount = 27.491577 Token0 see pool.go for the math
	// (50 - 27.491577) / ( 27.491577 / 1.0001^2000) = 1 ie. pool ratio is maintained
	// depositValue = depositAmount - (autoswapedAmountAsToken0 * fee)
	//              = 50 - 27.491577 * (1 - 1.0001^-2)
	//              = 49.9945025092
	// SharesIssued = depositValue * existing shares / (existingValue + autoSwapFee)
	//              = 49.9945025092 * 111.069527 / (111.069527 + .005497490762642563860802206452577)
	//              = 49.992027

	s.assertAccountSharesInt(s.bob, 2000, 2, math.NewInt(49992027))
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
	// swapAmount = 22.508423 Token1 see pool.go for the math
	// swapAmountAsToken0 = 27.4915750352
	// (22.508423 * 1.0001^2000) / (50 - 22.508423) = 1 ie. pool ratio is maintained
	// depositValue = depositAmountAsToken0 - (autoswapedAmountAsToken0 * fee)
	//              = 61.06952725039 - 27.4915750352 * (1 - 1.0001^-2)
	//              = 61.06402976002
	// SharesIssued = depositValue * existing shares / (existingValue + autoSwapFee)
	//              = 61.06402976002 * 111.069527 / (111.069527 + 0.00549749037)
	//              = 61061007

	s.assertAccountSharesInt(s.bob, 2000, 2, math.NewInt(61061007))
}

func (s *DexTestSuite) TestAutoswapDoubleSided0To1() {
	s.fundAliceBalances(30, 50)
	s.fundBobBalances(50, 50)

	// GIVEN a pool with double-sided liquidity at ratio 3:5
	s.aliceDeposits(NewDeposit(30, 50, -4000, 10))
	s.assertAccountSharesInt(s.alice, -4000, 10, math.NewInt(63516672))

	// WHEN bob deposits a ratio of 1:1 tokenA and B
	s.bobDeposits(NewDeposit(50, 50, -4000, 10))
	s.assertPoolLiquidity(80, 100, -4000, 10)

	// THEN his deposit is autoswapped
	// He receives 83.5 shares
	// swapAmount = 10.553662 Token0 see pool.go for the math
	// (50 - 10.553662) / (50 + 10.553662 / 1.0001^-4000) = 3/5 ie. pool ratio is maintained
	// depositValue = depositAmountAsToken0 - (autoswapedAmountAsToken0 * fee)
	//              = 83.5166725838 -  10.553662 * (1 - 1.0001^-10)
	//              = 83.506124724
	// SharesIssued = depositValue * existing shares / (existingValue + autoSwapFee)
	//              = 83.506124724 * 63.516672 / (63.516672 + .010547859)
	//              = 83.492258

	s.assertAccountSharesInt(s.bob, -4000, 10, math.NewInt(83492258))
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
	// swapAmount = 14.263300 Token1 see pool.go for the math
	// swapAmountAsToken0 = 9.5611671213
	// depositValue = depositAmountAsToken0 - (autoswapedAmountAsToken0 * fee)
	//              = 83.5166725838 -  9.5611671213 * (1 - 1.0001^-10)
	//              = 83.5071166732
	// SharesIssued = depositValue * existing shares / (existingValue + autoSwapFee)
	//              = 83.5071166732 * 70.110003 / (70.1100035503 + .009555910582)
	//              = 83.495735

	s.assertAccountSharesInt(s.bob, -4000, 10, math.NewInt(83495735))
}
