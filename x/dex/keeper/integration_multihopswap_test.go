package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
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
		amountAInt := math.NewInt(int64(p.AmountA)).Mul(denomMultiple)
		amountBInt := math.NewInt(int64(p.AmountB)).Mul(denomMultiple)
		coins := sdk.NewCoins(
			sdk.NewCoin(p.TokenA, amountAInt),
			sdk.NewCoin(p.TokenB, amountBInt),
		)
		s.fundAccountBalancesWithDenom(s.bob, coins)
		pairID := types.PairID{Token0: p.TokenA, Token1: p.TokenB}
		s.depositsSuccess(
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
		NewPoolSetup("TokenA", "TokenB", 0, 100, -1, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, -1, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 100, -1, 1),
	)

	// WHEN alice multihopswaps A<>B => B<>C => C<>D,
	route := [][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}}
	s.aliceMultiHopSwaps(route, 100, math_utils.MustNewPrecDecFromStr("0.9"), false)

	// THEN alice gets out 99 TokenD
	s.assertAccountBalanceWithDenom(s.alice, "TokenA", 0)
	s.assertAccountBalanceWithDenom(s.alice, "TokenD", 100)

	s.assertDexBalanceWithDenom("TokenA", 100)
	s.assertDexBalanceWithDenom("TokenB", 100)
	s.assertDexBalanceWithDenom("TokenC", 100)
	s.assertDexBalanceWithDenom("TokenD", 0)
}

func (s *DexTestSuite) TestMultiHopSwapSingleRouteWithDust() {
	s.fundAliceBalances(200, 0) // 200_000_000 TokenA

	// GIVEN liquidity in pools A<>B, B<>C, C<>D,
	s.SetupMultiplePools(
		// tick 109999 with fee 1 will be traded for A -> B at 110000
		NewPoolSetup("TokenA", "TokenB", 0, 1, 109999, 1), // tick 110000 = 59841.22218557191867154759205905
		NewPoolSetup("TokenB", "TokenC", 0, 1, -1, 1),
		NewPoolSetup("TokenC", "TokenD", 0, 1, -1, 1),
	)

	// WHEN alice multihopswaps A<>B => B<>C => C<>D,
	route := [][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}}
	msg := types.NewMsgMultiHopSwap(
		s.alice.String(),
		s.alice.String(),
		route,
		math.NewInt(int64(60_000)),
		math_utils.MustNewPrecDecFromStr("0.000000013"),
		false,
	) // 60_000A (59841 real) -> 1B -> 1C -> 1D
	_, err := s.msgServer.MultiHopSwap(s.Ctx, msg)
	s.Assert().Nil(err)

	// THEN alice gets out 1 TokenD

	// 200_000_000 - 60_000 (swap in) + 159 (dust)
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenA", math.NewInt(199_940_158)) // alice balance - spent + received dust
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenB", math.NewInt(0))
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenC", math.NewInt(0))
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenD", math.NewInt(1)) // received TokenD after swap

	s.assertDexBalanceWithDenomInt("TokenA", math.NewInt(59_842))
	s.assertDexBalanceWithDenomInt("TokenB", math.NewInt(1_000_000))
	s.assertDexBalanceWithDenomInt("TokenC", math.NewInt(1_000_000))
	s.assertDexBalanceWithDenomInt("TokenD", math.NewInt(999_999))
}

// same test for receiving dust, but this time should receive multiple dust tokens
func (s *DexTestSuite) TestMultiHopSwapSingleRouteWithManyDustTokens() {
	s.fundAliceBalances(2_000, 0) // 2_000_000_000 TokenA

	// GIVEN liquidity in pools A<>B, B<>C, C<>D,
	s.SetupMultiplePools(
		// tick 109999 with fee 1 will be traded for A -> B at 110000
		NewPoolSetup("TokenA", "TokenB", 0, 1, 109_999, 1), // tick 110000 = 59841.22218557191867154759205905
		NewPoolSetup("TokenB", "TokenC", 0, 1, 59_999, 1),  // tick 60000 = 403.307791072
		NewPoolSetup("TokenC", "TokenD", 0, 1, -1, 1),
	)

	// WHEN alice multihopswaps A<>B => B<>C => C<>D,
	route := [][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}}
	msg := types.NewMsgMultiHopSwap(
		s.alice.String(),
		s.alice.String(),
		route,
		math.NewInt(int64(600_000_000)),
		math_utils.MustNewPrecDecFromStr("0.00000000013"),
		false,
	) // 600_000_000A (599968093 real, 31907 dust) -> 10_026B (9679 real, 347 dust) -> 24C -> 24D
	_, err := s.msgServer.MultiHopSwap(s.Ctx, msg)
	s.Assert().Nil(err)

	// THEN alice gets out 1 TokenD

	// 2_000_000_000 - 600_000_000 (swap in) + 1_468_096 (dust)
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenA", math.NewInt(1_400_031_906)) // alice balance - spent + received dust
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenB", math.NewInt(346))
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenC", math.NewInt(0))
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenD", math.NewInt(24)) // received TokenD after swap

	s.assertDexBalanceWithDenomInt("TokenA", math.NewInt(600_000_000-31906)) // send - dust
	s.assertDexBalanceWithDenomInt("TokenB", math.NewInt(1_000_000-346))     // dex_balance - dust
	s.assertDexBalanceWithDenomInt("TokenC", math.NewInt(1_000_000))         // dex_balance
	s.assertDexBalanceWithDenomInt("TokenD", math.NewInt(1_000_000-24))      // dex_balance - swap_output
}

func (s *DexTestSuite) TestMultiHopSwapSingleRouteWithManyDustTokens2() {
	s.fundAliceBalances(200, 0) // 200_000_000 TokenA

	// GIVEN liquidity in pools A<>B, B<>C, C<>D,
	s.SetupMultiplePools(
		// tick 109999 with fee 1 will be traded for A -> B at 110000
		NewPoolSetup("TokenA", "TokenB", 0, 1, 109999, 1), // tick 110000 =>  59,960,905 TokenA In; 1,0002 TokenB out
		NewPoolSetup("TokenB", "TokenC", 0, 1, -60001, 1), // tick -60000 => 1,002 TokenB In;  404,114 TokenC out
		NewPoolSetup("TokenC", "TokenD", 0, 1, 69999, 1),  // tick 70000 =>  403,419 TokenC In; 368 TokenD out
	)

	// WHEN alice multihopswaps A<>B => B<>C => C<>D,
	msg := types.NewMsgMultiHopSwap(
		s.alice.String(),
		s.alice.String(),
		[][]string{{"TokenA", "TokenB", "TokenC", "TokenD"}},
		math.NewInt(int64(60_000_000)),
		math_utils.MustNewPrecDecFromStr("0.0000013"),
		false,
	)
	_, err := s.msgServer.MultiHopSwap(s.Ctx, msg)

	s.Assert().Nil(err)

	s.assertAccountBalanceWithDenomInt(s.alice, "TokenA", math.NewInt(140_039_095))
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenB", math.NewInt(0))
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenC", math.NewInt(694))
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenD", math.NewInt(368))

	s.assertDexBalanceWithDenomInt("TokenA", math.NewInt(60_000_000-39_095))
	s.assertDexBalanceWithDenomInt("TokenB", math.NewInt(1_000_000))
	s.assertDexBalanceWithDenomInt("TokenC", math.NewInt(1_000_000-694))
	s.assertDexBalanceWithDenomInt("TokenD", math.NewInt(1_000_000-368))
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
		types.ErrNoLiquidity,
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
		types.ErrLimitPriceNotSatisfied,
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
		NewPoolSetup("TokenA", "TokenB", 0, 100, -1, 1),
		NewPoolSetup("TokenB", "TokenC", 0, 100, 0, 1),
		NewPoolSetup("TokenC", "TokenX", 0, 50, 0, 1),
		NewPoolSetup("TokenC", "TokenX", 0, 50, 2200, 1),
		NewPoolSetup("TokenB", "TokenD", 0, 100, 0, 1),
		NewPoolSetup("TokenD", "TokenX", 0, 50, 0, 1),
		NewPoolSetup("TokenD", "TokenX", 0, 50, 2200, 1),
		NewPoolSetup("TokenB", "TokenE", 0, 100, -1, 1),
		NewPoolSetup("TokenE", "TokenX", 0, 100, -1, 1),
	)

	// WHEN alice multihopswaps with three routes the first two routes fail and the third works
	routes := [][]string{
		{"TokenA", "TokenB", "TokenC", "TokenX"},
		{"TokenA", "TokenB", "TokenD", "TokenX"},
		{"TokenA", "TokenB", "TokenE", "TokenX"},
	}
	s.aliceMultiHopSwaps(routes, 100, math_utils.MustNewPrecDecFromStr("0.91"), false)

	// THEN swap succeeds through route A<>B, B<>E, E<>X
	s.assertAccountBalanceWithDenom(s.alice, "TokenA", 0)
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenX", math.NewInt(100_000_000))
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenA", Token1: "TokenB"},
		100,
		0,
		-1,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenB", Token1: "TokenE"},
		100,
		0,
		-1,
		1,
	)
	s.assertLiquidityAtTickWithDenom(
		&types.PairID{Token0: "TokenE", Token1: "TokenX"},
		100,
		0,
		-1,
		1,
	)

	// Other pools are unaffected
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenB", Token1: "TokenC"},
		math.NewInt(0),
		math.NewInt(100_000_000),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenC", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(50_000_000),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenC", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(50_000_000),
		2200,
		1,
	)
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenB", Token1: "TokenD"},
		math.NewInt(0),
		math.NewInt(100_000_000),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenD", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(50_000_000),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenD", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(50_000_000),
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
		types.ErrLimitPriceNotSatisfied,
		routes,
		100,
		math_utils.MustNewPrecDecFromStr("0.91"),
		true,
	)

	// and with findFirstRoute

	s.aliceMultiHopSwapFails(
		types.ErrLimitPriceNotSatisfied,
		routes,
		100,
		math_utils.MustNewPrecDecFromStr("0.91"),
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

	s.assertAccountBalanceWithDenomInt(s.alice, "TokenA", math.NewInt(1)) // dust left
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenX", math.NewInt(134_943_366))
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenA", Token1: "TokenB"},
		math.NewInt(99_999_999),
		math.NewInt(10000),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenB", Token1: "TokenE"},
		math.NewInt(99_990_000),
		math.NewInt(19_999),
		0,
		1,
	)

	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenE", Token1: "TokenX"},
		math.NewInt(99_980_001),
		math.NewInt(865_056_634),
		-3000,
		1,
	)

	// Other pools are unaffected
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenB", Token1: "TokenC"},
		math.NewInt(0),
		math.NewInt(100_000_000),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenC", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(1_000_000_000),
		-1000,
		1,
	)
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenB", Token1: "TokenD"},
		math.NewInt(0),
		math.NewInt(100_000_000),
		0,
		1,
	)
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenD", Token1: "TokenX"},
		math.NewInt(0),
		math.NewInt(1_000_000_000),
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
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenA", math.NewInt(1)) // dust left
	s.assertAccountBalanceWithDenomInt(s.alice, "TokenX", math.NewInt(99_880_066))
	s.assertLiquidityAtTickWithDenomInt(
		&types.PairID{Token0: "TokenM", Token1: "TokenX"},
		math.NewInt(99_890_055),
		math.NewInt(119_934),
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
