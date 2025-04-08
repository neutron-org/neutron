package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestSimulateWithdrawFilledLimitOrder() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(0, 20)

	trancheKey := s.aliceLimitSells("TokenA", 0, 50)

	s.bobLimitSells("TokenB", -10, 20, types.LimitOrderType_FILL_OR_KILL)

	req := &types.QuerySimulateWithdrawFilledLimitOrderRequest{
		Msg: &types.MsgWithdrawFilledLimitOrder{
			Creator:    s.alice.String(),
			TrancheKey: trancheKey,
		},
	}

	resp, err := s.App.DexKeeper.SimulateWithdrawFilledLimitOrder(s.Ctx, req)
	s.NoError(err)

	s.Equal(sdk.NewCoin("TokenB", math.NewInt(20_000_000)), resp.Resp.TakerCoinOut)
	s.Equal(sdk.NewCoin("TokenA", math.ZeroInt()), resp.Resp.MakerCoinOut)

	s.assertDexBalances(30, 20)
}

func (s *DexTestSuite) TestSimulateWithdrawFilledLimitOrderFails() {
	s.fundAliceBalances(50, 0)
	s.fundBobBalances(0, 20)

	trancheKey := s.aliceLimitSells("TokenA", 0, 50)

	s.bobLimitSells("TokenB", -10, 20, types.LimitOrderType_FILL_OR_KILL)

	req := &types.QuerySimulateWithdrawFilledLimitOrderRequest{
		Msg: &types.MsgWithdrawFilledLimitOrder{
			Creator:    s.bob.String(),
			TrancheKey: trancheKey,
		},
	}

	resp, err := s.App.DexKeeper.SimulateWithdrawFilledLimitOrder(s.Ctx, req)
	s.ErrorIs(err, types.ErrValidLimitOrderTrancheNotFound)
	s.Nil(resp)
}
