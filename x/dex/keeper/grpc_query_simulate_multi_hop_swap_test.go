package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestSimulateMultiHopSwapSingleRoute() {
	// GIVEN liquidity in pools A<>B, B<>C, C<>D,
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 100, 0, 1),
	)

	// WHEN alice multihopswaps A<>B => B<>C => C<>D,
	route := &types.MultiHopRoute{Hops: []string{"TokenA", "TokenB", "TokenC", "TokenD"}}
	req := &types.QuerySimulateMultiHopSwapRequest{
		Msg: &types.MsgMultiHopSwap{
			Routes:         []*types.MultiHopRoute{route},
			AmountIn:       math.NewInt(100_000_000),
			ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			PickBestRoute:  false,
		},
	}
	resp, err := s.App.DexKeeper.SimulateMultiHopSwap(s.Ctx, req)
	s.NoError(err)

	// THEN alice would get out ~99 BIGTokenD
	expectedOutCoin := sdk.NewCoin("TokenD", math.NewInt(99970003))
	s.Assert().True(resp.Resp.CoinOut.Equal(expectedOutCoin))
	s.Assert().Equal(route, resp.Resp.Route)
	dust := sdk.NewCoins(resp.Resp.Dust...)
	expectedDust := sdk.NewCoin("TokenA", math.OneInt())
	s.Assert().True(dust.Equal(sdk.NewCoins(expectedDust)))

	// Nothing changes on the dex
	s.assertDexBalanceWithDenom("TokenA", 0)
	s.assertDexBalanceWithDenom("TokenB", 100)
	s.assertDexBalanceWithDenom("TokenC", 100)
	s.assertDexBalanceWithDenom("TokenD", 100)
}

func (s *DexTestSuite) TestSimulateMultiHopSwapMultiRoute() {
	// GIVEN liquidity in pools A<>B, B<>C, C<>D,
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenD", 0, 150, -1000, 1),
	)

	// WHEN alice multihopswaps A<>B => B<>C => C<>D,
	route1 := &types.MultiHopRoute{Hops: []string{"TokenA", "TokenB", "TokenC", "TokenD"}}
	route2 := &types.MultiHopRoute{Hops: []string{"TokenA", "TokenB", "TokenD"}}
	req := &types.QuerySimulateMultiHopSwapRequest{
		Msg: &types.MsgMultiHopSwap{
			Routes:         []*types.MultiHopRoute{route1, route2},
			AmountIn:       math.NewInt(100_000_000),
			ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			PickBestRoute:  true,
		},
	}
	resp, err := s.App.DexKeeper.SimulateMultiHopSwap(s.Ctx, req)
	s.NoError(err)

	// THEN alice would get out ~110 BIGTokenD
	expectedOutCoin := sdk.NewCoin("TokenD", math.NewInt(110494438))
	s.Assert().True(resp.Resp.CoinOut.Equal(expectedOutCoin))
	s.Assert().Equal(route2, resp.Resp.Route)
	dust := sdk.NewCoins(resp.Resp.Dust...)
	expectedDust := sdk.NewCoin("TokenA", math.OneInt())
	s.Assert().True(dust.Equal(sdk.NewCoins(expectedDust)))

	// Nothing changes on the dex
	s.assertDexBalanceWithDenom("TokenA", 0)
	s.assertDexBalanceWithDenom("TokenB", 100)
	s.assertDexBalanceWithDenom("TokenC", 100)
	s.assertDexBalanceWithDenom("TokenD", 250)
}

func (s *DexTestSuite) TestSimulateMultiHopSwapFails() {
	// GIVEN liquidity in pools A<>B, B<>C, C<>D,
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 100, 0, 1),
	)

	// WHEN alice multihopswaps A<>B => B<>C => C<>D with a high limit price,
	route := &types.MultiHopRoute{Hops: []string{"TokenA", "TokenB", "TokenC", "TokenD"}}
	req := &types.QuerySimulateMultiHopSwapRequest{
		Msg: &types.MsgMultiHopSwap{
			Routes:         []*types.MultiHopRoute{route},
			AmountIn:       math.NewInt(100_000_000),
			ExitLimitPrice: math_utils.MustNewPrecDecFromStr("2"),
			PickBestRoute:  false,
		},
	}
	// THEN her request fails
	resp, err := s.App.DexKeeper.SimulateMultiHopSwap(s.Ctx, req)
	s.Error(err, types.ErrLimitPriceNotSatisfied)
	s.Nil(resp)
}
