package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/utils/math"
	"github.com/neutron-org/neutron/x/dex/types"
)

type PoolSetup struct {
	TokenA    string
	TokenB    string
	AmountA   int
	AmountB   int
	TickIndex int
	Fee       int
}

func NewPoolSetup(tokenA, tokenB string, amountA, amountB, tickIndex, fee int) PoolSetup {
	return PoolSetup{
		TokenA:    tokenA,
		TokenB:    tokenB,
		AmountA:   amountA,
		AmountB:   amountB,
		TickIndex: tickIndex,
		Fee:       fee,
	}
}

func (s *DexTestSuite) SetupMultiplePools(poolSetups ...PoolSetup) {
	for _, p := range poolSetups {
		coins := sdk.NewCoins(
			sdk.NewCoin(p.TokenA, math.NewInt(int64(p.AmountA))),
			sdk.NewCoin(p.TokenB, math.NewInt(int64(p.AmountB))),
		)
		s.fundAccountBalancesWithDenom(s.bob, coins)
		pairID := types.PairID{Token0: p.TokenA, Token1: p.TokenB}
		s.deposits(
			s.bob,
			[]*Deposit{NewDeposit(p.AmountA, p.AmountB, p.TickIndex, p.Fee)},
			pairID,
		)
	}
}

func (s *DexTestSuite) TestMultiHopSwapSingleRoute() {
	s.fundAliceBalances(100, 0)

	// GIVEN liquidity in pools A<>B, B<>C, C<>D,
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 100, 0, 1),
	)

	// WHEN alice multihopswaps A<>B => B<>C => C<>D,
	route := [][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}}
	s.aliceMultiHopSwaps(route, 100, math_utils.MustNewPrecDecFromStr("0.9"), false)

	// THEN alice gets out 99 TokenD
	s.assertAccountBalanceWithDenom(s.alice, "TokenA", 0)
	s.assertAccountBalanceWithDenom(s.alice, "TokenD", 97)

	s.assertDexBalanceWithDenom("TokenA", 100)
	s.assertDexBalanceWithDenom("TokenB", 100)
	s.assertDexBalanceWithDenom("TokenC", 100)
	s.assertDexBalanceWithDenom("TokenD", 3)
}

func (s *DexTestSuite) TestMultiHopSwapInsufficientLiquiditySingleRoute() {
	s.fundAliceBalances(100, 0)

	// GIVEN liquidity in pools A<>B, B<>C, C<>D with insufficient liquidity in C<>D
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 50, 0, 1),
	)

	// THEN alice multihopswap fails
	route := [][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}}
	s.aliceMultiHopSwapFails(
		types.ErrInsufficientLiquidity,
		route,
		100,
		math_utils.MustNewPrecDecFromStr("0.9"),
		false,
	)
}

func (s *DexTestSuite) TestMultiHopSwapLimitPriceNotMetSingleRoute() {
	s.fundAliceBalances(100, 0)

	// GIVEN liquidity in pools A<>B, B<>C, C<>D with insufficient liquidity in C<>D
	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 100, 1200, 1),
	)

	// THEN alice multihopswap fails
	route := [][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}}
	s.aliceMultiHopSwapFails(
		types.ErrExitLimitPriceHit,
		route,
		50,
		math_utils.MustNewPrecDecFromStr("0.9"),
		false,
	)
}

func (s *DexTestSuite) TestMultiHopSwapMultiRouteOneGood() {
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

	// WHEN alice multihopswaps with three routes the first two routes fail and the third works
	routes := [][]string{
		{"TokenA", "TokenB", "TokenC", "TokenX"},
		{"TokenA", "TokenB", "TokenD", "TokenX"},
		{"TokenA", "TokenB", "TokenE", "TokenX"},
	}
	s.aliceMultiHopSwaps(routes, 100, math_utils.MustNewPrecDecFromStr("0.9"), false)

	// THEN swap succeeds through route A<>B, B<>E, E<>X
	s.assertAccountBalanceWithDenom(s.alice, "TokenA", 0)
	s.assertAccountBalanceWithDenom(s.alice, "TokenX", 97)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenA", Token1: "TokenB"},
		math.NewInt(100),
		math.NewInt(1),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenE"},
		math.NewInt(99),
		math.NewInt(2),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenE", Token1: "TokenX"},
		math.NewInt(98),
		math.NewInt(3),
		0,
		1,
	)

	// Other pools are unaffected
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenC"},
		math.NewInt(0),
		math.NewInt(100),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenC", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(50),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenC", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(50),
		2200,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenD"},
		math.NewInt(0),
		math.NewInt(100),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenD", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(50),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenD", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(50),
		2200,
		1,
	)
}

func (s *DexTestSuite) TestMultiHopSwapMultiRouteAllFail() {
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
	s.aliceMultiHopSwapFails(
		types.ErrExitLimitPriceHit,
		routes,
		100,
		math_utils.MustNewPrecDecFromStr("0.9"),
		true,
	)

	// and with findFirstRoute

	s.aliceMultiHopSwapFails(
		types.ErrExitLimitPriceHit,
		routes,
		100,
		math_utils.MustNewPrecDecFromStr("0.9"),
		false,
	)
}

func (s *DexTestSuite) TestMultiHopSwapMultiRouteFindBestRoute() {
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
	s.aliceMultiHopSwaps(routes, 100, math_utils.MustNewPrecDecFromStr("0.9"), true)

	// THEN swap succeeds through route A<>B, B<>E, E<>X

	s.assertAccountBalanceWithDenom(s.alice, "TokenA", 0)
	s.assertAccountBalanceWithDenom(s.alice, "TokenX", 132)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenA", Token1: "TokenB"},
		math.NewInt(100),
		math.NewInt(1),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenE"},
		math.NewInt(99),
		math.NewInt(2),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenE", Token1: "TokenX"},
		math.NewInt(98),
		math.NewInt(868),
		-3000,
		1,
	)

	// Other pools are unaffected
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenC"},
		math.NewInt(0),
		math.NewInt(100),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenC", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(1000),
		-1000,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenD"},
		math.NewInt(0),
		math.NewInt(100),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenD", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(1000),
		-2000,
		1,
	)
}

func (s *DexTestSuite) TestMultiHopSwapLongRouteWithCache() {
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
	s.aliceMultiHopSwaps(routes, 100, math_utils.MustNewPrecDecFromStr("0.8"), true)

	// THEN swap succeeds with second route
	s.assertAccountBalanceWithDenom(s.alice, "TokenA", 0)
	s.assertAccountBalanceWithDenom(s.alice, "TokenX", 88)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenM", Token1: "TokenX"},
		math.NewInt(89),
		math.NewInt(12),
		0,
		1,
	)
}

func (s *DexTestSuite) TestMultiHopSwapEventsEmitted() {
	s.fundAliceBalances(100, 0)

	s.SetupMultiplePools(
		NewPoolSetup("TokenA", "TokenB", 0, 100, 0, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
	)

	route := [][]string{{"TokenA", "TokenB", "TokenC"}}
	s.aliceMultiHopSwaps(route, 100, math_utils.MustNewPrecDecFromStr("0.9"), false)

	// 8 tickUpdateEvents are emitted 4x for pool setup 4x for two swaps
	s.AssertNEventValuesEmitted(types.TickUpdateEventKey, 8)
}
