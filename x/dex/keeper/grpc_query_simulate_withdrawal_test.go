package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestSimulateWithdrawal() {
	s.fundAliceBalances(20, 0)

	s.aliceDeposits(NewDeposit(10, 0, 0, 1), NewDeposit(10, 0, 1, 1))

	req := &types.QuerySimulateWithdrawalRequest{
		Msg: &types.MsgWithdrawal{
			Creator:         s.alice.String(),
			Receiver:        s.alice.String(),
			TokenA:          "TokenA",
			TokenB:          "TokenB",
			SharesToRemove:  []math.Int{math.NewInt(5), math.NewInt(9)},
			TickIndexesAToB: []int64{0, 1},
			Fees:            []uint64{1, 1},
		},
	}
	resp, err := s.App.DexKeeper.SimulateWithdrawal(s.Ctx, req)
	s.NoError(err)

	s.Equal(math.NewInt(14), resp.Resp.Reserve0Withdrawn)
	s.Equal(math.ZeroInt(), resp.Resp.Reserve1Withdrawn)

	expectedSharesBurned := sdk.NewCoins(
		sdk.NewCoin("neutron/pool/0", math.NewInt(5)),
		sdk.NewCoin("neutron/pool/1", math.NewInt(9)),
	)
	sharesBurned := sdk.NewCoins(resp.Resp.SharesBurned...)
	s.True(sharesBurned.Equal(expectedSharesBurned))

	// Dex Balances Unchanged

	s.assertDexBalances(20, 0)
}

func (s *DexTestSuite) TestSimulateWithdrawalFails() {
	s.fundAliceBalances(20, 0)

	s.aliceDeposits(NewDeposit(10, 0, 0, 1), NewDeposit(10, 0, 1, 1))

	req := &types.QuerySimulateWithdrawalRequest{
		Msg: &types.MsgWithdrawal{
			Creator:         s.alice.String(),
			Receiver:        s.alice.String(),
			TokenA:          "TokenA",
			TokenB:          "TokenB",
			SharesToRemove:  []math.Int{math.NewInt(5), math.NewInt(200_000_000)},
			TickIndexesAToB: []int64{0, 1},
			Fees:            []uint64{1, 1},
		},
	}
	resp, err := s.App.DexKeeper.SimulateWithdrawal(s.Ctx, req)
	s.Error(err, types.ErrInsufficientShares)
	s.Nil(resp)
}
