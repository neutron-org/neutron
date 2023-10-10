package keeper_test

func (s *MsgServerTestSuite) TestAutoswapperWithdraws() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	bobDep0 := 10
	bobDep1 := 10
	tickIndex := 200
	fee := 5

	bobSharesMinted := s.calcSharesMinted(int64(tickIndex), int64(bobDep0), int64(bobDep1))

	s.bobDeposits(NewDeposit(bobDep0, bobDep1, tickIndex, fee))
	s.assertBobBalances(40, 40)
	s.assertDexBalances(10, 10)

	// Alice deposits at a different balance ratio
	s.aliceDepositsWithOptions(NewDepositWithOptions(12, 5, tickIndex, fee, DepositOptions{DisableAutoswap: false}))
	s.assertAliceBalances(38, 45)
	s.assertDexBalances(22, 15)

	// Calculated expected amounts out
	autoswapSharesMinted := s.calcAutoswapSharesMinted(int64(tickIndex), uint64(fee), 7, 0, 5, 5, bobSharesMinted.Int64(), bobSharesMinted.Int64())
	// totalShares := autoswapSharesMinted.Add(math.NewInt(20))

	aliceExpectedBalance0, aliceExpectedBalance1, dexExpectedBalance0, dexExpectedBalance1 := s.calcExpectedBalancesAfterWithdrawOnePool(autoswapSharesMinted, s.alice, int64(tickIndex), uint64(fee))

	s.aliceWithdraws(NewWithdrawalInt(autoswapSharesMinted, int64(tickIndex), uint64(fee)))

	s.assertAliceBalances(aliceExpectedBalance0.Int64(), aliceExpectedBalance1.Int64())
	s.assertDexBalances(dexExpectedBalance0.Int64(), dexExpectedBalance1.Int64())
}

func (s *MsgServerTestSuite) TestAutoswapOtherDepositorWithdraws() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	bobDep0 := 10
	bobDep1 := 10
	tickIndex := 150
	fee := 10

	bobSharesMinted := s.calcSharesMinted(int64(tickIndex), int64(bobDep0), int64(bobDep1))

	s.bobDeposits(NewDeposit(bobDep0, bobDep1, tickIndex, fee))
	s.assertBobBalances(40, 40)
	s.assertDexBalances(10, 10)

	// Alice deposits at a different balance ratio
	s.aliceDepositsWithOptions(NewDepositWithOptions(10, 7, tickIndex, fee, DepositOptions{DisableAutoswap: false}))
	s.assertAliceBalances(40, 43)
	s.assertDexBalances(20, 17)

	// Calculated expected amounts out

	bobExpectedBalance0, bobExpectedBalance1, dexExpectedBalance0, dexExpectedBalance1 := s.calcExpectedBalancesAfterWithdrawOnePool(bobSharesMinted, s.bob, int64(tickIndex), uint64(fee))

	s.bobWithdraws(NewWithdrawal(bobSharesMinted.Int64(), int64(tickIndex), uint64(fee)))

	s.assertBobBalances(bobExpectedBalance0.Int64(), bobExpectedBalance1.Int64())
	s.assertDexBalances(dexExpectedBalance0.Int64(), dexExpectedBalance1.Int64())
}

func (s *MsgServerTestSuite) TestAutoswapBothWithdraws() {
	s.fundAliceBalances(50, 50)
	s.fundBobBalances(50, 50)

	// GIVEN
	// create spread around -1, 1
	bobDep0 := 10
	bobDep1 := 10
	tickIndex := 10000
	fee := 5

	bobSharesMinted := s.calcSharesMinted(int64(tickIndex), int64(bobDep0), int64(bobDep1))

	s.bobDeposits(NewDeposit(bobDep0, bobDep1, tickIndex, fee))
	s.assertBobBalances(40, 40)
	s.assertDexBalances(10, 10)

	// Alice deposits at a different balance ratio
	s.aliceDepositsWithOptions(NewDepositWithOptions(10, 5, tickIndex, fee, DepositOptions{DisableAutoswap: false}))
	s.assertAliceBalances(40, 45)
	s.assertDexBalances(20, 15)

	// Calculated expected amounts out
	autoswapSharesMinted := s.calcAutoswapSharesMinted(int64(tickIndex), uint64(fee), 5, 0, 5, 5, bobSharesMinted.Int64(), bobSharesMinted.Int64())
	// totalShares := autoswapSharesMinted.Add(math.NewInt(20))

	bobExpectedBalance0, bobExpectedBalance1, dexExpectedBalance0, dexExpectedBalance1 := s.calcExpectedBalancesAfterWithdrawOnePool(bobSharesMinted, s.bob, int64(tickIndex), uint64(fee))

	s.bobWithdraws(NewWithdrawal(bobSharesMinted.Int64(), int64(tickIndex), uint64(fee)))

	s.assertBobBalances(bobExpectedBalance0.Int64(), bobExpectedBalance1.Int64())
	s.assertDexBalances(dexExpectedBalance0.Int64(), dexExpectedBalance1.Int64())

	aliceExpectedBalance0, aliceExpectedBalance1, dexExpectedBalance0, dexExpectedBalance1 := s.calcExpectedBalancesAfterWithdrawOnePool(autoswapSharesMinted, s.alice, int64(tickIndex), uint64(fee))

	s.aliceWithdraws(NewWithdrawalInt(autoswapSharesMinted, int64(tickIndex), uint64(fee)))

	s.assertAliceBalances(aliceExpectedBalance0.Int64(), aliceExpectedBalance1.Int64())
	s.assertDexBalances(dexExpectedBalance0.Int64(), dexExpectedBalance1.Int64())
}
