package keeper_test

import (
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestDepositMultiCompleteFailure() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// liquidity TokenA at tick 0
	s.aliceDeposits(
		NewDeposit(5, 0, 0, 1),
	)

	// WHEN
	// Alice deposits 0 A, 5 B at tick 0
	// THEN
	// second deposit's ratio is different than pool after the first, so amounts will be rounded to 0,0 and tx will fail

	err := types.ErrZeroTrueDeposit
	s.aliceDeposits(NewDeposit(5, 0, 0, 1))
	s.assertAliceDepositFails(
		err,
		NewDeposit(5, 0, 2, 1),
		NewDepositWithOptions(0, 5, 0, 1, types.DepositOptions{DisableAutoswap: true}), // fails
	)
}

func (s *DexTestSuite) TestDepositMultiPartialBELFailureWithFailTx() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// no existing liquidity

	// WHEN
	// alice deposits 5 A, 5 B at tick 0 and 5 A at tick -2
	// THEN
	// second deposit fails BEL check

	err := types.ErrDepositBehindEnemyLines
	s.assertAliceDepositFails(
		err,
		NewDeposit(5, 5, 0, 1),
		NewDepositWithOptions(5, 0, 3, 1, types.DepositOptions{FailTxOnBel: true}),
	)
}

func (s *DexTestSuite) TestDepositMultiPartialBELFailureWithoutFailTx() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// no existing liquidity

	// WHEN
	// alice deposits 5 A, 5 B at tick 0 and 5 A at tick -2
	// THEN
	// second deposit fails BEL check
	// but other deposit succeeds

	resp, err := s.deposits(
		s.alice,
		[]*Deposit{
			NewDeposit(5, 5, 0, 1),
			NewDepositWithOptions(5, 0, 3, 1, types.DepositOptions{FailTxOnBel: false}),
		},
	)

	s.NoError(err)
	s.Assert().EqualValues(1, resp.FailedDeposits[0].DepositIdx)
	s.assertLiquidityAtTick(5, 5, 0, 1)
}

func (s *DexTestSuite) TestDepositMultiSuccess() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// no existing liquidity

	// WHEN
	// alice deposits 5 A, 5 B at tick 0, 5 A at tick -5 and 5 B at tick 5
	s.aliceDeposits(
		NewDeposit(5, 5, 0, 1),
		NewDeposit(5, 0, -6, 1),
		NewDeposit(0, 5, 4, 1),
	)

	// THEN
	// all deposits should go through
	s.assertAliceBalances(40, 40)
	s.assertLiquidityAtTick(5, 5, 0, 1)
	s.assertLiquidityAtTick(5, 0, -6, 1)
	s.assertLiquidityAtTick(0, 5, 4, 1)
	s.assertDexBalances(10, 10)
}
