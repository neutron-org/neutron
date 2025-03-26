package keeper_test

import (
	"cosmossdk.io/math"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestDepositDoubleSidedInSpreadCurrTickAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -5, 5
	s.aliceDeposits(NewDeposit(10, 10, 0, 5))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-5)
	s.assertCurr0To1(5)

	// WHEN
	// deposit in spread (10 of A,B at tick 0 fee 1)
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(30, 30)
	s.assertDexBalances(20, 20)

	// THEN
	// assert currentTick1To0, currTick0to1 moved
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestDepositDoubleSidedAroundSpreadCurrTickNotAdjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	s.aliceDeposits(NewDeposit(10, 10, 0, 1))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)

	// WHEN
	// deposit around spread (10 of A,B at tick 0 fee 3)
	s.aliceDeposits(NewDeposit(10, 10, 0, 3))
	s.assertAliceBalances(30, 30)
	s.assertDexBalances(20, 20)

	// THEN
	// assert CurrentTick0To1, CurrentTick1To0 unchanged
	s.assertCurr1To0(-1)
	s.assertCurr0To1(1)
}

func (s *DexTestSuite) TestDepositDoubleSidedHalfInSpreadCurrTick0To1Adjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -5, 5
	s.aliceDeposits(NewDeposit(10, 10, 0, 5))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-5)
	s.assertCurr0To1(5)

	// WHEN
	// deposit half in spread (10 of A,B at tick 1 fee 5)
	s.aliceDeposits(NewDeposit(10, 10, 1, 5))
	s.assertAliceBalances(30, 30)
	s.assertDexBalances(20, 20)

	// THEN
	// assert CurrTick1to0 unchanged, CurrTick0to1 adjusted
	s.assertCurr1To0(-4)
	s.assertCurr0To1(5)
}

func (s *DexTestSuite) TestDepositDoubleSidedHalfInSpreadCurrTick1To0Adjusted() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// create spread around -5, 5
	s.aliceDeposits(NewDeposit(10, 10, 0, 5))
	s.assertAliceBalances(40, 40)
	s.assertDexBalances(10, 10)
	s.assertCurr1To0(-5)
	s.assertCurr0To1(5)

	// WHEN
	// deposit half in spread (10 of A,B at tick -1 fee 5)
	s.aliceDeposits(NewDeposit(10, 10, -1, 5))
	s.assertAliceBalances(30, 30)
	s.assertDexBalances(20, 20)

	// THEN
	// assert CurrTick0to1 unchanged, CurrTick1to0 adjusted
	s.assertCurr1To0(-5)
	s.assertCurr0To1(4)
}

func (s *DexTestSuite) TestDepositDoubleSidedFirstSharesMintedTotal() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// empty pool
	s.assertPoolShares(0, 1, 0)
	s.assertPoolLiquidity(0, 0, 0, 1)

	// WHEN
	// depositing 10, 5 at tick 0 fee 1
	s.aliceDeposits(NewDeposit(10, 5, 0, 1))

	// THEN
	// 15 shares are minted and are the total
	s.assertPoolShares(0, 1, 15)
}

func (s *DexTestSuite) TestDepositDoubleSidedFirstSharesMintedUser() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// empty pool
	s.assertAliceShares(0, 1, 0)
	s.assertPoolLiquidity(0, 0, 0, 1)
	s.assertAliceShares(0, 1, 0)

	// WHEN
	// depositing 10, 5 at tick 0 fee 1
	s.aliceDeposits(NewDeposit(10, 5, 0, 1))

	// THEN
	// 15 shares are minted for alice
	s.assertAliceShares(0, 1, 15)
}

func (s *DexTestSuite) TestDepositDoubleSidedExistingSharesMintedTotal() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// tick 0 fee 1 has existing liquidity of 10 tokenA and 5 tokenB, shares are 15
	s.aliceDeposits(NewDeposit(10, 5, 0, 1))
	s.assertPoolShares(0, 1, 15)
	s.assertPoolLiquidity(10, 5, 0, 1)

	// WHEN
	// depositing 10, 5 at tick 0 fee 1
	s.aliceDeposits(NewDeposit(10, 5, 0, 1))

	// THEN
	// 15 more shares are minted and the total is 30
	s.assertPoolShares(0, 1, 30)
}

func (s *DexTestSuite) TestDepositDoubleSidedExistingSharesMintedUser() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// tick 0 fee 1 has existing liquidity of 10 tokenA and 5 tokenB, shares are 15
	s.aliceDeposits(NewDeposit(10, 5, 0, 1))
	s.assertAliceShares(0, 1, 15)
	s.assertPoolShares(0, 1, 15)
	s.assertPoolLiquidity(10, 5, 0, 1)

	// WHEN
	// alice deposits 6, 3 at tick 0 fee 1
	s.aliceDeposits(NewDeposit(6, 3, 0, 1))

	// THEN
	// 9 more shares are minted for alice for a total of 24
	s.assertAliceShares(0, 1, 24)
}

func (s *DexTestSuite) TestDepositValueAccural() {
	s.fundAliceBalances(100, 0)
	s.fundBobBalances(1000, 1000)
	s.fundCarolBalances(100, 1)

	// Alice deposits 100TokenA @ tick0 => 100 shares
	s.aliceDeposits(NewDeposit(100, 0, 0, 10))
	s.assertAliceShares(0, 10, 100)
	s.assertLiquidityAtTick(100, 0, 0, 10)

	// Lots of trade activity => ~110 ExistingValueToken0

	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			s.bobLimitSells("TokenB", -11, 1000, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
		} else {
			s.bobLimitSells("TokenA", 11, 1000, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
		}
	}
	s.assertLiquidityAtTickInt(math.NewInt(110516593), math.ZeroInt(), 0, 10)
	s.assertDexBalancesInt(math.NewInt(110516593), math.ZeroInt())

	s.assertLiquidityAtTickInt(math.NewInt(110516593), math.ZeroInt(), 0, 10)
	// Carol deposits 100TokenA @tick0
	s.carolDeposits(NewDeposit(100, 0, 0, 10))
	s.assertLiquidityAtTickInt(math.NewInt(210516593), math.ZeroInt(), 0, 10)
	s.assertAccountSharesInt(s.carol, 0, 10, math.NewInt(90484150))

	// Alice gets back 100% of the accrued trade value
	s.aliceWithdraws(NewWithdrawal(100, 0, 10))
	s.assertAliceBalancesInt(math.NewInt(110516593), math.NewInt(0))
	// AND carol get's back only what she put in
	s.carolWithdraws(NewWithdrawalInt(math.NewInt(90484150), 0, 10))
	s.assertCarolBalances(100, 1)
}

func (s *DexTestSuite) TestDepositToken0BELWithSwapPartial() {
	s.fundAliceBalances(50, 10)
	s.fundBobBalances(0, 30)

	// GIVEN TokenB liquidity at tick 2002-2004
	s.bobDeposits(NewDeposit(0, 10, 2001, 1),
		NewDeposit(0, 10, 2002, 1),
		NewDeposit(0, 10, 2003, 1),
	)
	// WHEN alice deposits TokenA at tick -2005 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(50, 10, 2006, 1, types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN some of alice's TokenA is swapped and she deposits ~13TokenA & 40TokenB
	// A = 50 - 30 * 1.0001^~2002 = 13.3
	// B = 30(from swap) + 10 (initial deposit) = 40
	// SharesIssued = 13.3 + 40 * 1.0001^2006 = 62

	s.Equal(math.NewInt(13347289), resp.Reserve0Deposited[0])
	s.Equal(math.NewInt(40000000), resp.Reserve1Deposited[0])
	s.Equal(math.NewInt(62232231), resp.SharesIssued[0].Amount)
	s.assertAliceBalances(0, 0)

	s.assertLiquidityAtTickInt(math.NewInt(13347289), math.NewInt(40000000), 2006, 1)
}

func (s *DexTestSuite) TestDepositToken0BELWithSwapAll() {
	s.fundAliceBalances(20, 10)
	s.fundBobBalances(0, 30)

	// GIVEN TokenB liquidity at tick 10,001
	s.bobDeposits(NewDeposit(0, 10, 10000, 1))
	// WHEN alice deposits TokenA at tick -10,002 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(20, 10, 10003, 1,
			types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true, SwapOnDepositSlopToleranceBps: 10},
		),
	)

	// THEN all of alice's token0 is swapped with 2 coins not swapped due to rounding
	// and she deposits 0TokenA & ~17.3TokenB
	// B = 10 + 20 / 1.0001^100001 = 17.3
	// SharesIssued = 17.3 * 1.0001^10003 = 47

	s.True(resp.Reserve0Deposited[0].IsZero())
	s.Equal(math.NewInt(17357220), resp.Reserve1Deposited[0])
	s.Equal(math.NewInt(47193612), resp.SharesIssued[0].Amount)
	s.assertAliceBalancesInt(math.NewInt(2), math.ZeroInt())

	s.assertLiquidityAtTickInt(math.ZeroInt(), math.NewInt(17357220), 10003, 1)
}

func (s *DexTestSuite) TestDepositToken1BELWithSwapPartial() {
	s.fundAliceBalances(10, 50)
	s.fundBobBalances(30, 30)

	// GIVEN TokenA liquidity at tick -2000 to -2002
	s.bobDeposits(NewDeposit(10, 0, 2001, 1),
		NewDeposit(10, 0, 2002, 1),
		NewDeposit(10, 0, 2003, 1),
	)
	// WHEN alice deposits TokenB at tick 1998 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(10, 50, 1998, 1, types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true}),
	)

	// THEN some of alice's tokenB is swapped and she deposits 40TokenA & ~25.4TokenB
	//
	// A = 30(from swap) + 10 (initial deposit) = 40
	// B = 50 - 30 * 1.0001^~-2001 = 25.4
	// SharesIssued = 40 + 25.4 * 1.0001^1998 = ~71

	s.Equal(math.NewInt(40000000), resp.Reserve0Deposited[0])
	s.Equal(math.NewInt(25440286), resp.Reserve1Deposited[0])
	s.Equal(math.NewInt(71066311), resp.SharesIssued[0].Amount)
	s.assertAliceBalances(0, 0)

	s.assertLiquidityAtTickInt(math.NewInt(40000000), math.NewInt(25440286), 1998, 1)
}

func (s *DexTestSuite) TestDepositToken1BELWithSwapAll() {
	s.fundAliceBalances(10, 20)
	s.fundBobBalances(30, 0)

	// GIVEN TokenA liquidity at tick 10,002
	s.bobDeposits(NewDeposit(10, 0, -10001, 1))
	// WHEN alice deposits TokenB at tick -10,003 (BEL)
	resp := s.aliceDeposits(
		NewDepositWithOptions(10, 20, -10004, 1,
			types.DepositOptions{FailTxOnBel: true, SwapOnDeposit: true, SwapOnDepositSlopToleranceBps: 10},
		),
	)

	// THEN all of alice's tokenB is swapped
	// and she deposits ~17.3TokenA and 0TokenB
	// A = 10 + 20 / 1.0001^10002 = 17.3
	// SharesIssued = 17.3

	s.Equal(math.NewInt(17356485), resp.Reserve0Deposited[0])
	s.True(resp.Reserve1Deposited[0].IsZero())
	s.Equal(math.NewInt(17356485), resp.SharesIssued[0].Amount)
	s.assertAliceBalancesInt(math.NewInt(0), math.ZeroInt())

	s.assertLiquidityAtTickInt(math.NewInt(17356485), math.ZeroInt(), -10004, 1)
}
