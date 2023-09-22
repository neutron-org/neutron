package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/dex/types"
)

func (s *MsgServerTestSuite) TestDepositMultiCompleteFailure() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// no existing liquidity

	// WHEN
	// alice deposits 0 A, 5 B at tick 0 fee 0 and 5 A, 0 B at tick 0 fee 0
	// THEN
	// second deposit's ratio is different than pool after the first, so amounts will be rounded to 0,0 and tx will fail

	err := types.ErrZeroTrueDeposit
	s.assertAliceDepositFails(
		err,
		NewDeposit(5, 0, 0, 1),
		NewDeposit(0, 5, 0, 1),
	)
}

func (s *MsgServerTestSuite) TestDepositMultiSuccess() {
	s.fundAliceBalances(50, 50)

	// GIVEN
	// no existing liquidity

	// WHEN
	// alice deposits 5 A, 5 B at tick 0 fee 0 and then 10 A, 10 B at tick 5 fee 0
	s.aliceDeposits(
		NewDeposit(5, 5, 0, 1),
		NewDeposit(10, 10, 0, 1),
	)

	// THEN
	// both deposits should go through
	s.assertAliceBalances(35, 35)
	s.assertLiquidityAtTick(sdk.NewInt(15), sdk.NewInt(15), 0, 1)
	s.assertDexBalances(15, 15)
}
