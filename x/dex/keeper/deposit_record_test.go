package keeper_test

import (
	"cosmossdk.io/math"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestGetAllDeposits() {
	s.fundAliceBalances(20, 20)
	// GIVEN Alice Deposits 3 positions and withdraws the first
	s.aliceDeposits(
		NewDeposit(1, 0, -50, 1),
		NewDeposit(5, 5, 0, 1),
		NewDeposit(0, 10, 2, 1),
	)
	s.aliceWithdraws(NewWithdrawal(1, -50, 1))

	// THEN GetAllDeposits returns the two remaining LP positions
	depositList := s.App.DexKeeper.GetAllDepositsForAddress(s.Ctx, s.alice)
	s.Assert().Equal(2, len(depositList))
	s.Assert().Equal(&types.DepositRecord{
		PairId:          defaultPairID,
		SharesOwned:     math.NewInt(10_000_000),
		CenterTickIndex: 0,
		LowerTickIndex:  -1,
		UpperTickIndex:  1,
		Fee:             1,
	},
		depositList[0],
	)
	s.Assert().Equal(&types.DepositRecord{
		PairId:          defaultPairID,
		SharesOwned:     math.NewInt(10_002_000),
		CenterTickIndex: 2,
		LowerTickIndex:  1,
		UpperTickIndex:  3,
		Fee:             1,
	},
		depositList[1],
	)
}
