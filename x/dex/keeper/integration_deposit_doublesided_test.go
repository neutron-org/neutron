package keeper_test

import (
	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/x/dex/types"
)

func (s *MsgServerTestSuite) TestDepositDoubleSidedInSpreadCurrTickAdjusted() {
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

func (s *MsgServerTestSuite) TestDepositDoubleSidedAroundSpreadCurrTickNotAdjusted() {
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

func (s *MsgServerTestSuite) TestDepositDoubleSidedHalfInSpreadCurrTick0To1Adjusted() {
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

func (s *MsgServerTestSuite) TestDepositDoubleSidedHalfInSpreadCurrTick1To0Adjusted() {
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

func (s *MsgServerTestSuite) TestDepositDoubleSidedCreatingArbBelow() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// deposit 10 of token A at tick 0 fee 1
	s.aliceDeposits(NewDeposit(10, 0, 0, 1))
	s.assertAliceBalances(40, 50)
	s.assertDexBalances(10, 0)
	s.assertPoolLiquidity(10, 0, 0, 1)

	// WHEN
	// depositing below enemy lines at tick -5
	// THEN
	// deposit should not fail with BEL error, balances and liquidity should not change at deposited tick

	s.aliceDeposits(NewDeposit(10, 10, -5, 1))

	// buying liquidity behind enemy lines doesn't break anything
	s.bobLimitSells("TokenA", 0, 10, types.LimitOrderType_FILL_OR_KILL)
}

func (s *MsgServerTestSuite) TestDepositDoubleSidedCreatingArbAbove() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// deposit 10 of token A at tick 0 fee 1
	s.aliceDeposits(NewDeposit(0, 10, 0, 1))
	s.assertAliceBalances(50, 40)
	s.assertDexBalances(0, 10)
	s.assertPoolLiquidity(0, 10, 0, 1)

	// WHEN
	// depositing above enemy lines at tick 5
	// THEN
	// deposit should not fail with BEL error, balances and liquidity should not change at deposited tick

	s.aliceDeposits(NewDeposit(10, 10, 5, 1))

	// buying liquidity behind enemy lines doesn't break anything
	s.bobLimitSells("TokenB", 0, 10, types.LimitOrderType_FILL_OR_KILL)
}

func (s *MsgServerTestSuite) TestDepositDoubleSidedFirstSharesMintedTotal() {
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

func (s *MsgServerTestSuite) TestDepositDoubleSidedFirstSharesMintedUser() {
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

func (s *MsgServerTestSuite) TestDepositDoubleSidedExistingSharesMintedTotal() {
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

func (s *MsgServerTestSuite) TestDepositDoubleSidedExistingSharesMintedUser() {
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

func (s *MsgServerTestSuite) TestDepositValueAccural() {
	s.fundAliceBalances(100, 0)
	s.fundBobBalances(1000, 1000)
	s.fundCarolBalances(100, 1)

	// Alice deposits 100TokenA @ tick0 => 100 shares
	s.aliceDeposits(NewDeposit(100, 0, 0, 10))
	s.assertAliceShares(0, 10, 100)
	s.assertLiquidityAtTick(math.NewInt(100), math.ZeroInt(), 0, 10)

	// Lots of trade activity => ~200 ExistingValueToken0

	for i := 0; i < 100; i++ {
		liquidityA, liquidityB := s.getLiquidityAtTick(0, 10)
		if i%2 == 0 {
			s.bobLimitSells("TokenB", -10, int(liquidityA.Int64())+10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
		} else {
			s.bobLimitSells("TokenA", 10, int(liquidityB.Int64())+10, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
		}
	}
	s.assertLiquidityAtTick(math.NewInt(200), math.NewInt(0), 0, 10)
	s.assertDexBalances(200, 0)

	// Carol deposits 100TokenA @tick0
	s.carolDeposits(NewDeposit(100, 1, 0, 10))
	s.assertCarolShares(0, 10, 50)

	s.aliceWithdraws(NewWithdrawal(100, 0, 10))
	s.assertAliceBalances(200, 0)

	s.carolWithdraws(NewWithdrawal(50, 0, 10))
	s.assertCarolBalances(100, 1)
}
