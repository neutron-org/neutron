package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (s *DexTestSuite) TestEstimateMultiHopSwapSingleRoute() {
	s.fundAliceBalances(100, 0)

	// GIVEN liquidity in pools A<>B, B<>C, C<>D,
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 100, 0, 1),
	)

	// WHEN alice multihopswaps A<>B => B<>C => C<>D,
	route := [][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}}
	coinOut := s.aliceEstimatesMultiHopSwap(route, 100, math_utils.MustNewPrecDecFromStr("0.9"), false)

	// THEN alice would get out ~99 BIGTokenD
	s.Assert().Equal(math.NewInt(99970003), coinOut.Amount)
	s.assertAccountBalanceWithDenom(s.alice, "TokenA", 100)
	s.assertAccountBalanceWithDenom(s.alice, "TokenD", 0)

	s.assertDexBalanceWithDenom("TokenA", 0)
	s.assertDexBalanceWithDenom("TokenB", 100)
	s.assertDexBalanceWithDenom("TokenC", 100)
	s.assertDexBalanceWithDenom("TokenD", 100)
}

func (s *DexTestSuite) TestEstimateMultiHopSwapInsufficientLiquiditySingleRoute() {
	s.fundAliceBalances(100, 0)

	// GIVEN liquidity in pools A<>B, B<>C, C<>D with insufficient liquidity in C<>D
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 50, 0, 1),
	)

	// THEN estimate multihopswap fails
	route := [][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}}
	s.aliceEstimatesMultiHopSwapFails(
		types.ErrNoLiquidity,
		route,
		100,
		math_utils.MustNewPrecDecFromStr("0.9"),
		false,
	)
}

func (s *DexTestSuite) TestEstimateMultiHopSwapLimitPriceNotMetSingleRoute() {
	s.fundAliceBalances(100, 0)

	// GIVEN liquidity in pools A<>B, B<>C, C<>D with insufficient liquidity in C<>D
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 100, 1200, 1),
	)

	// THEN estimate multihopswap fails
	route := [][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}}
	s.aliceEstimatesMultiHopSwapFails(
		types.ErrLimitPriceNotSatisfied,
		route,
		50,
		math_utils.MustNewPrecDecFromStr("0.9"),
		false,
	)
}

func (s *DexTestSuite) TestEstimateMultiHopSwapMultiRouteOneGood() {
	s.fundAliceBalances(100, 0)

	// GIVEN viable liquidity in pools A<>B, B<>E, E<>X
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenX", 0, 50, 0, 1),
		NewPoolSetup("TokenC", "TokenX", 0, 50, 2200, 1),
		NewPoolSetup("TokenB", "TokenD", 0, 100, 0, 1),
		NewPoolSetup("TokenD", "TokenX", 0, 50, 0, 1),
		NewPoolSetup("TokenD", "TokenX", 0, 50, 2200, 1),
		NewPoolSetup("TokenB", "TokenE", 0, 100, 0, 1),
		NewPoolSetup("TokenE", "TokenX", 0, 100, 0, 1),
	)

	// WHEN estimate multihopswaps with three routes the first two routes fail and the third works
	routes := [][]string{
		{"TokenA", "TokenB", "TokenC", "TokenX"},
		{"TokenA", "TokenB", "TokenD", "TokenX"},
		{"TokenA", "TokenB", "TokenE", "TokenX"},
	}

	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenA", Token1: "TokenB"},
		0,
		100,
		0,
		1,
	)

	coinOut := s.aliceEstimatesMultiHopSwap(routes, 100, math_utils.MustNewPrecDecFromStr("0.91"), false)

	// THEN swap estimation succeeds through route A<>B, B<>E, E<>X

	s.Assert().Equal(math.NewInt(99970003), coinOut.Amount)

	// pools and accounts are unaffected
	s.assertAccountBalanceWithDenom(s.alice, "TokenA", 100)
	s.assertAccountBalanceWithDenom(s.alice, "TokenX", 0)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenA", Token1: "TokenB"},
		0,
		100,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenE"},
		0,
		100,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenE", Token1: "TokenX"},
		0,
		100,
		0,
		1,
	)

	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenC"},
		0,
		100,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenC", Token1: "TokenX"},
		0,
		50,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenC", Token1: "TokenX"},
		0,
		50,
		2200,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenD"},
		0,
		100,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenD", Token1: "TokenX"},
		0,
		50,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenD", Token1: "TokenX"},
		0,
		50,
		2200,
		1,
	)
}

func (s *DexTestSuite) TestEstimateMultiHopSwapMultiRouteAllFail() {
	s.fundAliceBalances(100, 0)

	// GIVEN liquidity in sufficient liquidity but inadequate prices
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenX", 0, 50, 0, 1),
		NewPoolSetup("TokenC", "TokenX", 0, 50, 2200, 1),
		NewPoolSetup("TokenB", "TokenD", 0, 100, 0, 1),
		NewPoolSetup("TokenD", "TokenX", 0, 50, 0, 1),
		NewPoolSetup("TokenD", "TokenX", 0, 50, 2200, 1),
		NewPoolSetup("TokenB", "TokenE", 0, 50, 0, 1),
		NewPoolSetup("TokenE", "TokenX", 0, 50, 2200, 1),
	)

	// WHEN alice multihopswaps with three routes they all fail
	routes := [][]string{
		{"TokenA", "TokenB", "TokenC", "TokenX"},
		{"TokenA", "TokenB", "TokenD", "TokenX"},
		{"TokenA", "TokenB", "TokenE", "TokenX"},
	}

	// Then fails with findBestRoute
	s.aliceEstimatesMultiHopSwapFails(
		types.ErrLimitPriceNotSatisfied,
		routes,
		100,
		math_utils.MustNewPrecDecFromStr("0.91"),
		true,
	)

	// and with findFirstRoute

	s.aliceEstimatesMultiHopSwapFails(
		types.ErrLimitPriceNotSatisfied,
		routes,
		100,
		math_utils.MustNewPrecDecFromStr("0.91"),
		false,
	)
}

func (s *DexTestSuite) TestEstimateMultiHopSwapMultiRouteFindBestRoute() {
	s.fundAliceBalances(100, 0)

	// GIVEN viable liquidity in pools but with a best route through E<>X
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenX", 0, 1000, -1000, 1),
		NewPoolSetup("TokenB", "TokenD", 0, 100, 0, 1),
		NewPoolSetup("TokenD", "TokenX", 0, 1000, -2000, 1),
		NewPoolSetup("TokenB", "TokenE", 0, 100, 0, 1),
		NewPoolSetup("TokenE", "TokenX", 0, 1000, -3000, 1),
	)

	// WHEN alice multihopswaps with three routes
	routes := [][]string{
		{"TokenA", "TokenB", "TokenC", "TokenX"},
		{"TokenA", "TokenB", "TokenD", "TokenX"},
		{"TokenA", "TokenB", "TokenE", "TokenX"},
	}
	coinOut := s.aliceEstimatesMultiHopSwap(routes, 100, math_utils.MustNewPrecDecFromStr("0.9"), true)

	// THEN swap succeeds through route A<>B, B<>E, E<>X

	// pools are unaffected
	s.Assert().Equal(sdk.NewCoin("TokenX", math.NewInt(134943366)), coinOut)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenA", Token1: "TokenB"},
		0,
		100,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenE"},
		0,
		100,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenE", Token1: "TokenX"},
		0,
		1000,
		-3000,
		1,
	)

	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenC"},
		0,
		100,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenC", Token1: "TokenX"},
		0,
		1000,
		-1000,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenD"},
		0,
		100,
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenD", Token1: "TokenX"},
		0,
		1000,
		-2000,
		1,
	)
}

func (s *DexTestSuite) TestEstimateMultiHopSwapLongRouteWithCache() {
	s.fundAliceBalances(100, 0)

	// GIVEN viable route from A->B->C...->L but last leg to X only possible through K->M->X
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 100, 0, 1),
		NewPoolSetup("TokenD", "TokenE", 0, 100, 0, 1),
		NewPoolSetup("TokenE", "TokenF", 0, 100, 0, 1),
		NewPoolSetup("TokenF", "TokenG", 0, 100, 0, 1),
		NewPoolSetup("TokenG", "TokenH", 0, 100, 0, 1),
		NewPoolSetup("TokenH", "TokenI", 0, 100, 0, 1),
		NewPoolSetup("TokenI", "TokenJ", 0, 100, 0, 1),
		NewPoolSetup("TokenJ", "TokenK", 0, 100, 0, 1),
		NewPoolSetup("TokenK", "TokenL", 0, 100, 0, 1),
		NewPoolSetup("TokenL", "TokenX", 0, 50, 0, 1),
		NewPoolSetup("TokenL", "TokenX", 0, 50, 100, 1),

		NewPoolSetup("TokenK", "TokenM", 0, 100, 0, 1),
		NewPoolSetup("TokenM", "TokenX", 0, 100, 0, 1),
	)

	// WHEN alice multihopswaps with two overlapping routes with only the last leg different
	routes := [][]string{
		{
			"TokenA", "TokenB", "TokenC", "TokenD", "TokenE", "TokenF",
			"TokenG", "TokenH", "TokenI", "TokenJ", "TokenK", "TokenL", "TokenX",
		},
		{
			"TokenA", "TokenB", "TokenC", "TokenD", "TokenE", "TokenF",
			"TokenG", "TokenH", "TokenI", "TokenJ", "TokenK", "TokenM", "TokenX",
		},
	}
	coinOut := s.aliceEstimatesMultiHopSwap(routes, 100, math_utils.MustNewPrecDecFromStr("0.8"), true)

	// THEN swap succeeds with second route
	s.Assert().Equal(coinOut, sdk.NewCoin("TokenX", math.NewInt(99880066)))
	s.assertAccountBalanceWithDenom(s.alice, "TokenA", 100)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenM", Token1: "TokenX"},
		0,
		100,
		0,
		1,
	)
}

func (s *DexTestSuite) TestEstimateMultiHopSwapEventsEmitted() {
	s.fundAliceBalances(100, 0)

	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
	)

	route := [][]string{{"TokenA", "TokenB", "TokenC"}}
	_ = s.aliceEstimatesMultiHopSwap(route, 100, math_utils.MustNewPrecDecFromStr("0.9"), false)

	// 8 tickUpdateEvents are emitted 4x for pool setup 4x for two swaps
	s.AssertEventValueNotEmitted(types.TickUpdateEventKey, "Expected no events")
}
