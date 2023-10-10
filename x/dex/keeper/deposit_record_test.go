package keeper_test

import (
	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/x/dex/types"
)

func (s *MsgServerTestSuite) TestGetAllDeposits() {
	s.fundAliceBalances(20, 20)
	// GIVEN Alice Deposits 3 positions and withdraws the first
	s.aliceDeposits(
		&Deposit{
			AmountA:   math.NewInt(1),
			AmountB:   math.NewInt(0),
			TickIndex: -50,
			Fee:       1,
		},
		&Deposit{
			AmountA:   math.NewInt(5),
			AmountB:   math.NewInt(5),
			TickIndex: 0,
			Fee:       1,
		},
		&Deposit{
			AmountA:   math.NewInt(0),
			AmountB:   math.NewInt(10),
			TickIndex: 2,
			Fee:       1,
		},
	)
	s.aliceWithdraws(&Withdrawal{
		TickIndex: -50,
		Fee:       1,
		Shares:    math.NewInt(1),
	},
	)

	// THEN GetAllDeposits returns the two remaining LP positions
	depositList := s.app.DexKeeper.GetAllDepositsForAddress(s.ctx, s.alice)
	s.Assert().Equal(2, len(depositList))
	s.Assert().Equal(&types.DepositRecord{
		PairID:          defaultPairID,
		SharesOwned:     math.NewInt(10),
		CenterTickIndex: 0,
		LowerTickIndex:  -1,
		UpperTickIndex:  1,
		Fee:             1,
	},
		depositList[0],
	)
	s.Assert().Equal(&types.DepositRecord{
		PairID:          defaultPairID,
		SharesOwned:     math.NewInt(10),
		CenterTickIndex: 2,
		LowerTickIndex:  1,
		UpperTickIndex:  3,
		Fee:             1,
	},
		depositList[1],
	)
}
