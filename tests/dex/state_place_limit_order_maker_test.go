package dex_state_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
	"strconv"
	"testing"
	"time"
)

// ExistingTokenAHolders
const (
	NoneLO               = "NoneLO"
	CreatorLO            = "CreatorLO"
	OneOtherLO           = "OneOtherLO"
	OneOtherAndCreatorLO = "OneOtherAndCreatorLO"
)

// tests autoswap, BehindEnemyLineLessLimit and BehindEnemyLineGreaterLimit only makes sense when ExistingTokenAHolders==NoneLO
// BehindEnemyLine
const (
	BehindEnemyLineNo           = "BehindEnemyLinesNo"           // no opposite liquidity to trigger swap
	BehindEnemyLineLessLimit    = "BehindEnemyLinesLessLimit"    // not enough opposite liquidity to swap maker lo fully
	BehindEnemyLineGreaterLimit = "BehindEnemyLinesGreaterLimit" // enough liquidity to swap make fully
)

const (
	DefaultSellPrice     = "2"
	DefaultBuyPriceTaker = "0.4" // 1/(DefaultSellPrice+0.5) immediately trade over DefaultSellPrice maker order
)

const MakerAmountIn = 1_000_000

type placeLimitOrderMakerTestParams struct {
	// State Conditions
	SharedParams
	ExistingLOLiquidityDistribution LiquidityDistribution
	ExistingTokenAHolders           string
	BehindEnemyLine                 string
	PreexistingTraded               bool
	// Message Variants
	OrderType int32 // JIT, GTT, GTC
}

func (p placeLimitOrderMakerTestParams) printTestInfo(t *testing.T) {
	t.Logf(`
		Existing ExistingTokenAHolders: %s
		BehindEnemyLine: %v
		Pre-existing Traded: %v
		Order Type: %v`,
		p.ExistingTokenAHolders,
		p.BehindEnemyLine,
		p.PreexistingTraded,
		dextypes.LimitOrderType_name[p.OrderType],
	)
}

func parseLOLiquidityDistribution(existingShareHolders, behindEnemyLine string, pairID *dextypes.PairID) LiquidityDistribution {
	tokenA := pairID.Token0
	tokenB := pairID.Token1
	switch {
	case existingShareHolders == NoneLO && behindEnemyLine == BehindEnemyLineLessLimit:
		// half "taker" deposit. We buy all of it by placing opposite maker order
		return LiquidityDistribution{
			TokenA: sdk.NewCoin(tokenA, math.ZeroInt()),
			TokenB: sdk.NewCoin(tokenB, BaseTokenAmountInt.QuoRaw(2)),
		}
	case existingShareHolders == NoneLO && behindEnemyLine == BehindEnemyLineGreaterLimit:
		// double "taker" deposit. We spend whole limit to partially consume LO.
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.ZeroInt()), TokenB: sdk.NewCoin(tokenB, math.NewInt(4).Mul(BaseTokenAmountInt))}
	default:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin(tokenB, math.NewInt(1).Mul(BaseTokenAmountInt))}
	}
}

func removeRedundantPlaceLOMakerTests(tcs []placeLimitOrderMakerTestParams) []placeLimitOrderMakerTestParams {
	// here we remove impossible cases such two side LO at the "same" (1/-1, 2/-2) ticks
	result := make([]placeLimitOrderMakerTestParams, 0)
	for _, tc := range tcs {
		if tc.ExistingTokenAHolders != NoneLO && tc.BehindEnemyLine != BehindEnemyLineNo {
			continue
		}
		if tc.PreexistingTraded && tc.ExistingTokenAHolders != OneOtherLO {
			// PreexistingTraded only make sense in case `tc.ExistingTokenAHolders == OneOtherLO`
			continue
		}
		result = append(result, tc)
	}
	return result
}

func hydratePlaceLOMakerTestCase(params map[string]string, pairID *dextypes.PairID) placeLimitOrderMakerTestParams {
	liquidityDistribution := parseLOLiquidityDistribution(params["ExistingTokenAHolders"], params["BehindEnemyLines"], pairID)
	return placeLimitOrderMakerTestParams{
		ExistingLOLiquidityDistribution: liquidityDistribution,
		ExistingTokenAHolders:           params["ExistingTokenAHolders"],
		BehindEnemyLine:                 params["BehindEnemyLines"],
		PreexistingTraded:               parseBool(params["PreexistingTraded"]),
		OrderType:                       dextypes.LimitOrderType_value[params["OrderType"]],
	}
}

func hydrateAllPlaceLOMakerTestCases(paramsList []map[string]string) []placeLimitOrderMakerTestParams {
	allTCs := make([]placeLimitOrderMakerTestParams, 0)
	for i, paramsRaw := range paramsList {
		pairID := generatePairID(i)
		tc := hydratePlaceLOMakerTestCase(paramsRaw, pairID)
		tc.PairID = pairID
		allTCs = append(allTCs, tc)
	}

	return removeRedundantPlaceLOMakerTests(allTCs)
}

func (s *DexStateTestSuite) setupLoState(params placeLimitOrderMakerTestParams) (trancheKey string) {
	liquidityDistr := params.ExistingLOLiquidityDistribution
	coins := sdk.NewCoins(liquidityDistr.TokenA, liquidityDistr.TokenB)
	switch params.BehindEnemyLine {
	case BehindEnemyLineNo:
		switch params.ExistingTokenAHolders {
		case OneOtherLO:
			s.FundAcc(s.alice, coins)
			res := s.makePlaceLOSuccess(s.alice, params.ExistingLOLiquidityDistribution.TokenA, params.ExistingLOLiquidityDistribution.TokenB.Denom, DefaultSellPrice, dextypes.LimitOrderType_GOOD_TIL_CANCELLED, nil)
			trancheKey = res.TrancheKey

		case OneOtherAndCreatorLO:
			s.FundAcc(s.alice, coins)
			res := s.makePlaceLOSuccess(s.alice, params.ExistingLOLiquidityDistribution.TokenA, params.ExistingLOLiquidityDistribution.TokenB.Denom, DefaultSellPrice, dextypes.LimitOrderType_GOOD_TIL_CANCELLED, nil)
			trancheKey = res.TrancheKey

			s.FundAcc(s.creator, coins)
			s.makePlaceLOSuccess(s.creator, params.ExistingLOLiquidityDistribution.TokenA, params.ExistingLOLiquidityDistribution.TokenB.Denom, DefaultSellPrice, dextypes.LimitOrderType_GOOD_TIL_CANCELLED, nil)

		case CreatorLO:
			s.FundAcc(s.creator, coins)
			res := s.makePlaceLOSuccess(s.creator, params.ExistingLOLiquidityDistribution.TokenA, params.ExistingLOLiquidityDistribution.TokenB.Denom, DefaultSellPrice, dextypes.LimitOrderType_GOOD_TIL_CANCELLED, nil)
			trancheKey = res.TrancheKey
		}

		if params.PreexistingTraded {
			s.FundAcc(s.bob, coins)
			// bob trades over the tranche
			InTokenBAmount := sdk.NewCoin(params.ExistingLOLiquidityDistribution.TokenB.Denom, BaseTokenAmountInt.QuoRaw(2))
			s.makePlaceLOSuccess(s.alice, InTokenBAmount, params.ExistingLOLiquidityDistribution.TokenA.Denom, DefaultBuyPriceTaker, dextypes.LimitOrderType_GOOD_TIL_CANCELLED, nil)
		}
	case BehindEnemyLineLessLimit:
		s.FundAcc(s.alice, coins)
		s.makePlaceLOSuccess(s.alice, params.ExistingLOLiquidityDistribution.TokenB, params.ExistingLOLiquidityDistribution.TokenA.Denom, DefaultBuyPriceTaker, dextypes.LimitOrderType_GOOD_TIL_CANCELLED, nil)
	case BehindEnemyLineGreaterLimit:
		s.FundAcc(s.alice, coins)
		s.makePlaceLOSuccess(s.alice, params.ExistingLOLiquidityDistribution.TokenB, params.ExistingLOLiquidityDistribution.TokenA.Denom, DefaultBuyPriceTaker, dextypes.LimitOrderType_GOOD_TIL_CANCELLED, nil)
	}
	return trancheKey
}

// assertLiquidity checks the amount of tokens at dex balance exactly equals the amount in all tranches (active + inactive)
// TODO: add AMM pools to check
func (s *DexStateTestSuite) assertLiquidity(id dextypes.PairID) {
	TokenAInReserves := math.ZeroInt()
	TokenBInReserves := math.ZeroInt()

	// Active tranches A -> B
	tranches, err := s.App.DexKeeper.LimitOrderTrancheAll(s.Ctx, &dextypes.QueryAllLimitOrderTrancheRequest{
		PairId:     id.CanonicalString(),
		TokenIn:    id.Token0,
		Pagination: nil,
	})
	s.Require().NoError(err)
	for _, t := range tranches.LimitOrderTranche {
		TokenAInReserves = TokenAInReserves.Add(t.ReservesMakerDenom)
		TokenBInReserves = TokenBInReserves.Add(t.ReservesTakerDenom)
	}

	// Active tranches B -> A
	tranches, err = s.App.DexKeeper.LimitOrderTrancheAll(s.Ctx, &dextypes.QueryAllLimitOrderTrancheRequest{
		PairId:     id.CanonicalString(),
		TokenIn:    id.Token1,
		Pagination: nil,
	})
	s.Require().NoError(err)
	for _, t := range tranches.LimitOrderTranche {
		TokenAInReserves = TokenAInReserves.Add(t.ReservesTakerDenom)
		TokenBInReserves = TokenBInReserves.Add(t.ReservesMakerDenom)
	}

	// Inactive tranches (expired or filled)
	// TODO: since it's impossible to filter tranches against a specific pair in a request, pagination request may be needed in some cases. Add pagination
	inactiveTranches, err := s.App.DexKeeper.InactiveLimitOrderTrancheAll(s.Ctx, &dextypes.QueryAllInactiveLimitOrderTrancheRequest{
		Pagination: nil,
	})
	s.Require().NoError(err)
	for _, t := range inactiveTranches.InactiveLimitOrderTranche {
		// A -> B
		if t.Key.TradePairId.MakerDenom == id.Token0 || t.Key.TradePairId.TakerDenom == id.Token1 {
			TokenAInReserves = TokenAInReserves.Add(t.ReservesMakerDenom)
			TokenBInReserves = TokenBInReserves.Add(t.ReservesTakerDenom)
		}
		// B -> A
		if t.Key.TradePairId.MakerDenom == id.Token1 || t.Key.TradePairId.TakerDenom == id.Token0 {
			TokenAInReserves = TokenAInReserves.Add(t.ReservesTakerDenom)
			TokenBInReserves = TokenBInReserves.Add(t.ReservesMakerDenom)
		}

	}

	s.assertDexBalance(id.Token0, TokenAInReserves)
	s.assertDexBalance(id.Token1, TokenBInReserves)

}

// We assume, if there is a TokenB tranche in dex module, it's always BEL.
func (s *DexStateTestSuite) expectedInOutTokensAmount(tokenA sdk.Coin, denomOut string) (amountOut math.Int) {
	pair := dextypes.MustNewPairID(tokenA.Denom, denomOut)
	amountOut = math.ZeroInt()
	// Active tranches B -> A
	tranches, err := s.App.DexKeeper.LimitOrderTrancheAll(s.Ctx, &dextypes.QueryAllLimitOrderTrancheRequest{
		PairId:     pair.CanonicalString(),
		TokenIn:    pair.Token1,
		Pagination: nil,
	})
	s.Require().NoError(err)
	reserveA := tokenA.Amount

	for _, t := range tranches.LimitOrderTranche {
		// users tokenA denom = tranche TakerDenom
		// t.ReservesMakerDenom - reserve TokenB we are going to get
		// t.Price() - price taker -> maker => 1/t.Price() - maker -> taker
		// maxSwap - max amount of tokenA (ReservesTakerDenom) tranche can consume us by changing ReservesMakerDenom -> ReservesTakerDenom
		maxSwap := math_utils.NewPrecDecFromInt(t.ReservesMakerDenom).Quo(t.Price()).TruncateInt()
		// we can swap full our tranche
		if maxSwap.GTE(reserveA) {
			// expected to get tokenB = tokenA*
			amountOut = amountOut.Add(math_utils.NewPrecDecFromInt(reserveA).Mul(t.Price()).TruncateInt())
			reserveA = math.ZeroInt()
			break
		}
		reserveA = reserveA.Sub(maxSwap)
		amountOut = amountOut.Add(t.ReservesMakerDenom)
	}
	return amountOut
}

func (s *DexStateTestSuite) assertExpectedTrancheKey(initialKey, msgKey string, params placeLimitOrderMakerTestParams) {
	// we expect initialKey != msgKey
	if params.ExistingTokenAHolders == NoneLO || params.PreexistingTraded || params.OrderType == int32(dextypes.LimitOrderType_GOOD_TIL_TIME) || params.OrderType == int32(dextypes.LimitOrderType_JUST_IN_TIME) {
		s.NotEqual(initialKey, msgKey)
		return
	}

	//otherwise they are equal
	s.Equal(initialKey, msgKey)
}

func TestPlaceLimitOrderMaker(t *testing.T) {
	testParams := []testParams{
		{field: "ExistingTokenAHolders", states: []string{NoneLO, CreatorLO, OneOtherLO, OneOtherAndCreatorLO}},
		{field: "BehindEnemyLines", states: []string{BehindEnemyLineNo, BehindEnemyLineLessLimit, BehindEnemyLineGreaterLimit}},
		{field: "PreexistingTraded", states: []string{True, False}},
		{field: "OrderType", states: []string{
			dextypes.LimitOrderType_name[int32(dextypes.LimitOrderType_GOOD_TIL_CANCELLED)],
			dextypes.LimitOrderType_name[int32(dextypes.LimitOrderType_GOOD_TIL_TIME)],
			dextypes.LimitOrderType_name[int32(dextypes.LimitOrderType_JUST_IN_TIME)],
		}},
	}
	testCasesRaw := generatePermutations(testParams)
	testCases := hydrateAllPlaceLOMakerTestCases(testCasesRaw)

	s := new(DexStateTestSuite)
	s.SetT(t)
	s.SetupTest()
	totalExpectedToSwap := math.ZeroInt()

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s.SetT(t)
			tc.printTestInfo(t)

			initialTrancheKey := s.setupLoState(tc)
			s.fundCreatorBalanceDefault(tc.PairID)
			//

			amountIn := sdk.NewCoin(tc.PairID.Token0, math.NewInt(MakerAmountIn))
			var expTime *time.Time
			if tc.OrderType == int32(dextypes.LimitOrderType_GOOD_TIL_TIME) {
				// any time is valid for tests
				t := time.Now()
				expTime = &t
			}
			expectedSwapTakerDenom := s.expectedInOutTokensAmount(amountIn, tc.PairID.Token1)
			totalExpectedToSwap = totalExpectedToSwap.Add(expectedSwapTakerDenom)
			resp, err := s.makePlaceLO(s.creator, amountIn, tc.PairID.Token1, DefaultSellPrice, dextypes.LimitOrderType(tc.OrderType), expTime)
			s.Require().NoError(err)

			// 1. generic liquidity check assertion
			s.assertLiquidity(*tc.PairID)
			// 2. BEL assertion
			s.intsApproxEqual("", expectedSwapTakerDenom, resp.TakerCoinOut.Amount, 1)
			// 3. TrancheKey assertion
			s.assertExpectedTrancheKey(initialTrancheKey, resp.TrancheKey, tc)

		})
	}
	s.SetT(t)
	// sanity check: at least one `expectedSwapTakerDenom` > 0
	s.True(totalExpectedToSwap.GT(math.ZeroInt()))
}
