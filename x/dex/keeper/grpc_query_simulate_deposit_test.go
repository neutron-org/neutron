package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestSimulateDeposit() {
	req := &types.QuerySimulateDepositRequest{
		Msg: &types.MsgDeposit{
			TokenA:          "TokenA",
			TokenB:          "TokenB",
			AmountsA:        []math.Int{math.OneInt(), math.OneInt()},
			AmountsB:        []math.Int{math.ZeroInt(), math.ZeroInt()},
			TickIndexesAToB: []int64{0, 1},
			Fees:            []uint64{1, 1},
			Options:         []*types.DepositOptions{{}, {}},
		},
	}
	resp, err := s.App.DexKeeper.SimulateDeposit(s.Ctx, req)
	s.NoError(err)
	s.Equal([]math.Int{math.OneInt(), math.OneInt()}, resp.Resp.Reserve0Deposited)
	s.Equal([]math.Int{math.ZeroInt(), math.ZeroInt()}, resp.Resp.Reserve1Deposited)

	expectedShares := sdk.NewCoins(
		sdk.NewCoin("neutron/pool/0", math.OneInt()),
		sdk.NewCoin("neutron/pool/1", math.OneInt()),
	)
	sharesIssued := sdk.NewCoins(resp.Resp.SharesIssued...)
	s.True(sharesIssued.Equal(expectedShares))

	s.assertDexBalances(0, 0)
}

func (s *DexTestSuite) TestSimulateDepositPartialFailure() {
	req := &types.QuerySimulateDepositRequest{
		Msg: &types.MsgDeposit{
			TokenA:          "TokenA",
			TokenB:          "TokenB",
			AmountsA:        []math.Int{math.OneInt(), math.ZeroInt()},
			AmountsB:        []math.Int{math.ZeroInt(), math.OneInt()},
			TickIndexesAToB: []int64{3, 0},
			Fees:            []uint64{1, 1},
			Options:         []*types.DepositOptions{{}, {}},
		},
	}
	resp, err := s.App.DexKeeper.SimulateDeposit(s.Ctx, req)
	s.NoError(err)
	s.Equal(resp.Resp.Reserve0Deposited, []math.Int{math.OneInt(), math.ZeroInt()})
	s.Equal(resp.Resp.Reserve1Deposited, []math.Int{math.ZeroInt(), math.ZeroInt()})
	s.Equal(uint64(1), resp.Resp.FailedDeposits[0].DepositIdx)
	s.Contains(resp.Resp.FailedDeposits[0].Error, types.ErrDepositBehindEnemyLines.Error())

	expectedShares := sdk.NewCoins(
		sdk.NewCoin("neutron/pool/0", math.OneInt()),
	)
	sharesIssued := sdk.NewCoins(resp.Resp.SharesIssued...)
	s.True(sharesIssued.Equal(expectedShares))
}

func (s *DexTestSuite) TestSimulateDepositFails() {
	req := &types.QuerySimulateDepositRequest{
		Msg: &types.MsgDeposit{
			TokenA:          "TokenA",
			TokenB:          "TokenB",
			AmountsA:        []math.Int{math.OneInt(), math.ZeroInt()},
			AmountsB:        []math.Int{math.ZeroInt(), math.OneInt()},
			TickIndexesAToB: []int64{3, 0},
			Fees:            []uint64{1, 1},
			Options:         []*types.DepositOptions{{}, {FailTxOnBel: true}},
		},
	}
	resp, err := s.App.DexKeeper.SimulateDeposit(s.Ctx, req)
	s.Error(err, types.ErrDepositBehindEnemyLines)
	s.Nil(resp)
}
