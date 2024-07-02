package keeper_test

import (
	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/v4/x/dex/types"
)

func (s *DexTestSuite) TestEstimatePlaceLimitOrderGTC() {
	// GIVEN liquidity  A<>B
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 1, 0, 1),
		NewPoolSetup("TokenA", "TokenB", 0, 1, 1, 1),
	)

	// WHEN estimate GTC Limit order selling "TokenA"
	resp, err := s.App.DexKeeper.EstimatePlaceLimitOrder(s.Ctx, &types.QueryEstimatePlaceLimitOrderRequest{
		TokenIn:          "TokenA",
		TokenOut:         "TokenB",
		TickIndexInToOut: 4,
		AmountIn:         math.NewInt(3_000_000),
		OrderType:        types.LimitOrderType_GOOD_TIL_CANCELLED,
	})
	s.NoError(err)

	// Then estimate is 3 BIG TokenA in with 2 BIG TokenB out from swap
	s.True(math.NewInt(3_000_000).Equal(resp.TotalInCoin.Amount), "Got %v", resp.TotalInCoin.Amount)
	s.Equal("TokenA", resp.TotalInCoin.Denom)

	s.True(math.NewInt(2_000_301).Equal(resp.SwapInCoin.Amount), "Got %v", resp.SwapInCoin.Amount)
	s.Equal("TokenA", resp.SwapInCoin.Denom)

	s.True(math.NewInt(2_000_000).Equal(resp.SwapOutCoin.Amount), "Got %v", resp.SwapOutCoin.Amount)
	s.Equal("TokenB", resp.SwapOutCoin.Denom)

	// AND state is not altered
	s.assertDexBalanceWithDenom("TokenA", 0)
	s.assertDexBalanceWithDenom("TokenB", 2)

	// No events are emitted
	s.AssertEventValueNotEmitted(types.TickUpdateEventKey, "Expected no events")

	// Subsequent transactions use the original BankKeeper
	// ie. The simulation bankkeeper is not retained giving users unlimited funds
	s.assertBobLimitSellFails(sdkerrors.ErrInsufficientFunds, "TokenA", -400_000, 100_000_000)
}

func (s *DexTestSuite) TestEstimatePlaceLimitOrderFoK() {
	// GIVEN liquidity TokenB
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 1, 0, 1),
		NewPoolSetup("TokenA", "TokenB", 0, 1, 1, 1),
	)

	// WHEN estimate FoK Limit order selling "TokenA"
	resp, err := s.App.DexKeeper.EstimatePlaceLimitOrder(s.Ctx, &types.QueryEstimatePlaceLimitOrderRequest{
		TokenIn:          "TokenA",
		TokenOut:         "TokenB",
		TickIndexInToOut: 4,
		AmountIn:         math.NewInt(1_500_000),
		OrderType:        types.LimitOrderType_FILL_OR_KILL,
	})
	s.NoError(err)

	// Then estimate is 1.5 BIG TokenA in with 1.5 BIG TokenB out from swap
	s.True(math.NewInt(1_500_000).Equal(resp.TotalInCoin.Amount), "Got %v", resp.TotalInCoin.Amount)
	s.Equal("TokenA", resp.TotalInCoin.Denom)

	s.True(math.NewInt(1_500_000).Equal(resp.SwapInCoin.Amount), "Got %v", resp.SwapInCoin.Amount)
	s.Equal("TokenA", resp.SwapInCoin.Denom)

	s.True(math.NewInt(1_499_800).Equal(resp.SwapOutCoin.Amount), "Got %v", resp.SwapOutCoin.Amount)
	s.Equal("TokenB", resp.SwapOutCoin.Denom)

	// AND state is not altered
	s.assertDexBalanceWithDenom("TokenA", 0)
	s.assertDexBalanceWithDenom("TokenB", 2)

	// No events are emitted
	s.AssertEventValueNotEmitted(types.TickUpdateEventKey, "Expected no events")

	// Subsequent transactions use the original BankKeeper
	s.assertBobLimitSellFails(sdkerrors.ErrInsufficientFunds, "TokenA", -400_000, 100_000_000)
}

func (s *DexTestSuite) TestEstimatePlaceLimitOrderFoKFails() {
	// GIVEN no liquidity

	// WHEN estimate placeLimitOrder
	resp, err := s.App.DexKeeper.EstimatePlaceLimitOrder(s.Ctx, &types.QueryEstimatePlaceLimitOrderRequest{
		TokenIn:          "TokenA",
		TokenOut:         "TokenB",
		TickIndexInToOut: 4,
		AmountIn:         math.NewInt(1_500_000),
		OrderType:        types.LimitOrderType_IMMEDIATE_OR_CANCEL,
	})

	// THEN error is returned
	s.ErrorIs(err, types.ErrLimitPriceNotSatisfied)
	s.Nil(resp)

	// Subsequent transactions use the original BankKeeper
	s.assertBobLimitSellFails(sdkerrors.ErrInsufficientFunds, "TokenA", -400_000, 100_000_000)
}
