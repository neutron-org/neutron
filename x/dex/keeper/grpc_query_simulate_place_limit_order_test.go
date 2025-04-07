package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestSimulatePlaceLimitOrderWithTick() {
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
	)

	req := &types.QuerySimulatePlaceLimitOrderRequest{
		Msg: &types.MsgPlaceLimitOrder{
			TokenIn:          "TokenA",
			TokenOut:         "TokenB",
			TickIndexInToOut: 5,
			AmountIn:         math.NewInt(20_000_000),
			OrderType:        types.LimitOrderType_FILL_OR_KILL,
		},
	}

	resp, err := s.App.DexKeeper.SimulatePlaceLimitOrder(s.Ctx, req)
	s.NoError(err)

	s.Equal(sdk.NewCoin("TokenA", math.NewInt(20_000_000)), resp.Resp.CoinIn)
	s.Equal(sdk.NewCoin("TokenA", math.NewInt(20_000_000)), resp.Resp.TakerCoinIn)
	s.Equal(sdk.NewCoin("TokenB", math.NewInt(19_998_000)), resp.Resp.TakerCoinOut)

	s.assertDexBalances(0, 100)
}

func (s *DexTestSuite) TestSimulatePlaceLimitOrderWithPrice() {
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 10, 0, 1),
		NewPoolSetup("TokenA", "TokenB", 0, 10, 100, 1),
	)

	price := math_utils.MustNewPrecDecFromStr("0.995")

	req := &types.QuerySimulatePlaceLimitOrderRequest{
		Msg: &types.MsgPlaceLimitOrder{
			TokenIn:        "TokenA",
			TokenOut:       "TokenB",
			LimitSellPrice: &price,
			AmountIn:       math.NewInt(20_000_000),
			OrderType:      types.LimitOrderType_GOOD_TIL_CANCELLED,
		},
	}

	resp, err := s.App.DexKeeper.SimulatePlaceLimitOrder(s.Ctx, req)
	s.NoError(err)

	s.Equal(sdk.NewCoin("TokenA", math.NewInt(20_000_000)), resp.Resp.CoinIn)
	s.Equal(sdk.NewCoin("TokenA", math.NewInt(10_001_000)), resp.Resp.TakerCoinIn)
	s.Equal(sdk.NewCoin("TokenB", math.NewInt(10_000_000)), resp.Resp.TakerCoinOut)

	s.assertDexBalances(0, 20)
}

func (s *DexTestSuite) TestSimulatePlaceLimitOrderFails() {
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
	)

	price := math_utils.MustNewPrecDecFromStr("1.1")

	req := &types.QuerySimulatePlaceLimitOrderRequest{
		Msg: &types.MsgPlaceLimitOrder{
			TokenIn:        "TokenA",
			TokenOut:       "TokenB",
			LimitSellPrice: &price,
			AmountIn:       math.NewInt(20_000_000),
			OrderType:      types.LimitOrderType_FILL_OR_KILL,
		},
	}

	resp, err := s.App.DexKeeper.SimulatePlaceLimitOrder(s.Ctx, req)
	s.ErrorIs(err, types.ErrFoKLimitOrderNotFilled)
	s.Nil(resp)

	s.assertDexBalances(0, 100)
}
