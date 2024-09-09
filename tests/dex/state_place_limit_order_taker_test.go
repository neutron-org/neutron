package dex_state_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
)

// LiquidityType
const (
	LO   = "LO"
	LP   = "LP"
	LOLP = "LOLP"
)

// LimitPrice
const (
	LOWSELLPRICE  = "LOWSELLPRICE"
	AVGSELLPRICE  = "AVGSELLPRICE"
	HIGHSELLPRICE = "HIGHSELLPRICE"
)

var (
	DefaultPriceDelta = math_utils.NewPrecDecWithPrec(1, 1) // 0.1
	DefaultStartPrice = math_utils.NewPrecDecWithPrec(2, 0) // 2.0
)

type placeLimitOrderTakerTestParams struct {
	PairID *dextypes.PairID

	// State Conditions
	LiquidityType     string
	TicksDistribution []int64

	// Message Variants
	OrderType    int32 // FillOrKill or ImmediateOrCancel
	AmountIn     sdk.Coin
	LimitPrice   math_utils.PrecDec
	MaxAmountOut *math.Int
}

func (p placeLimitOrderTakerTestParams) printTestInfo(t *testing.T) {
	t.Logf(`
        LiquidityType: %s
		TicksDistribution: %v
		OrderType: %v
		AmountIn: %v
		LimitPrice: %v,
		MaxAmountOut: %v`,
		p.LiquidityType,
		p.TicksDistribution,
		dextypes.LimitOrderType_name[p.OrderType],
		p.AmountIn,
		p.LimitPrice,
		p.MaxAmountOut,
	)
}

func hydratePlaceLOTakerTestCase(params map[string]string, pairID *dextypes.PairID) placeLimitOrderTakerTestParams {
	ticks, err := strconv.Atoi(params["TicksDistribution"])
	if err != nil {
		panic(err)
	}
	amountInShare, err := strconv.Atoi(params["AmountIn"])
	if err != nil {
		panic(err)
	}
	// average sell price is defined by loop over the ticks in `setupLoTakerState`
	// and ~ ((2+0.1*(ticksAmount-1))+2)/2 = 2+0.05*(ticksAmount-1)
	// to buy 100% we want to put ~ BaseTokenAmountInt*(2+0.05*(ticksAmount-1)) as amountIn
	avgPrice := DefaultStartPrice.Add(
		DefaultPriceDelta.QuoInt64(2).MulInt64(int64(ticks - 1)),
	)
	amountIn := avgPrice.MulInt(BaseTokenAmountInt).MulInt64(int64(amountInShare)).QuoInt64(100).TruncateInt()

	maxOutShare, err := strconv.Atoi(params["MaxAmountOut"])
	if err != nil {
		panic(err)
	}

	var maxAmountOut *math.Int
	if maxOutShare > 0 {
		maxAmountOut = &math.Int{}
		*maxAmountOut = BaseTokenAmountInt.MulRaw(int64(maxOutShare)).QuoRaw(100)
	}

	LimitPrice := DefaultStartPrice // LOWSELLPRICE
	switch params["LimitPrice"] {
	case AVGSELLPRICE:
		LimitPrice = avgPrice
	case HIGHSELLPRICE:
		// 2 * max price
		LimitPrice = DefaultStartPrice.Add(DefaultPriceDelta.MulInt64(int64(ticks)).MulInt64Mut(2))
	}
	return placeLimitOrderTakerTestParams{
		LiquidityType:     params["LiquidityType"],
		TicksDistribution: generateTicks(ticks),
		OrderType:         dextypes.LimitOrderType_value[params["OrderType"]],
		AmountIn:          sdk.NewCoin(pairID.Token1, amountIn),
		MaxAmountOut:      maxAmountOut,
		LimitPrice:        math_utils.OnePrecDec().Quo(LimitPrice.Add(DefaultPriceDelta)),
	}
}

func hydrateAllPlaceLOTakerTestCases(paramsList []map[string]string) []placeLimitOrderTakerTestParams {
	allTCs := make([]placeLimitOrderTakerTestParams, 0)
	for i, paramsRaw := range paramsList {
		pairID := generatePairID(i)
		tc := hydratePlaceLOTakerTestCase(paramsRaw, pairID)
		tc.PairID = pairID
		allTCs = append(allTCs, tc)
	}

	return allTCs
}

func generateTicks(ticksAmount int) []int64 {
	ticks := make([]int64, 0, ticksAmount)
	for i := 0; i < ticksAmount; i++ {
		tick, err := dextypes.CalcTickIndexFromPrice(DefaultStartPrice.Add(DefaultPriceDelta.MulInt64(int64(i))))
		if err != nil {
			panic(err)
		}
		ticks = append(ticks, tick)
	}
	return ticks
}

func (s *DexStateTestSuite) setupLoTakerState(params placeLimitOrderTakerTestParams) {
	if params.LiquidityType == None {
		return
	}
	coins := sdk.NewCoins(sdk.NewCoin(params.PairID.Token0, BaseTokenAmountInt), sdk.NewCoin(params.PairID.Token1, BaseTokenAmountInt))
	s.FundAcc(s.alice, coins)
	// BaseTokenAmountInt is full liquidity
	tickLiquidity := BaseTokenAmountInt.QuoRaw(int64(len(params.TicksDistribution)))
	if params.LiquidityType == LOLP {
		tickLiquidity = tickLiquidity.QuoRaw(2)
	}
	for _, tick := range params.TicksDistribution {
		// hit both if LOLP
		if strings.Contains(params.LiquidityType, LO) {
			price := dextypes.MustCalcPrice(tick)
			amountIn := sdk.NewCoin(params.PairID.Token0, tickLiquidity)
			s.makePlaceLOSuccess(s.alice, amountIn, params.PairID.Token1, price.String(), dextypes.LimitOrderType_GOOD_TIL_CANCELLED, nil)
		}
		if strings.Contains(params.LiquidityType, LP) {
			liduidity := LiquidityDistribution{
				TokenA: sdk.NewCoin(params.PairID.Token0, tickLiquidity),
				TokenB: sdk.NewCoin(params.PairID.Token1, math.ZeroInt()),
			}
			// tick+DefaultFee to put liquidity the same tick as LO
			_, err := s.makeDeposit(s.alice, liduidity, DefaultFee, tick+DefaultFee, true)
			s.NoError(err)
		}
	}
}

func MaxAmountAOut(params placeLimitOrderTakerTestParams) math.Int {
	// liquidity equally distributed over the ticks with a delta `DefaultPriceDelta` starting from `DefaultStartPrice`
	// we find amount of ticks (tickOffset) are being covered by limitPrice
	// and that is the max liquidity we can swap in

	if params.LiquidityType == None {
		return math.ZeroInt()
	}
	// see `setupLoTakerState`
	tickLiquidity := BaseTokenAmountInt.QuoRaw(int64(len(params.TicksDistribution)))

	tickOffset := math_utils.OnePrecDec().Quo(params.LimitPrice).Sub(DefaultStartPrice).Quo(DefaultPriceDelta).Ceil().TruncateInt()
	liquidity := tickLiquidity.Mul(tickOffset)

	return math.MinInt(liquidity, BaseTokenAmountInt)
}

func ExpectedInOut(params placeLimitOrderTakerTestParams) (math.Int, math.Int) {
	if params.LiquidityType == None {
		return math.ZeroInt(), math.ZeroInt()
	}
	LimitTick, err := dextypes.CalcTickIndexFromPrice(math_utils.OnePrecDec().Quo(params.LimitPrice))
	if err != nil {
		panic(err)
	}
	tickLiquidity := BaseTokenAmountInt.QuoRaw(int64(len(params.TicksDistribution)))
	TotalIn := math.ZeroInt()
	TotalOut := math.ZeroInt()
	for _, tick := range params.TicksDistribution {
		if LimitTick > tick {
			break
		}
		toOut := tickLiquidity
		if params.MaxAmountOut != nil {
			toOut = math.MinInt(toOut, params.MaxAmountOut.Sub(TotalOut))
		}

		toIn := dextypes.MustCalcPrice(tick).MulInt(toOut).Ceil().TruncateInt()
		if toIn.GT(params.AmountIn.Amount.Sub(TotalIn)) {
			toIn = params.AmountIn.Amount.Sub(TotalIn)
			toOut = dextypes.MustCalcPrice(-1 * tick).MulInt(toIn).Ceil().TruncateInt()
		}
		TotalIn = TotalIn.Add(
			toIn,
		)
		TotalOut = TotalOut.Add(toOut)
	}
	return TotalIn, TotalOut
}

func (s *DexStateTestSuite) handleTakerErrors(params placeLimitOrderTakerTestParams, err error) {
	if params.OrderType == int32(dextypes.LimitOrderType_FILL_OR_KILL) {
		maxIn, _ := ExpectedInOut(params)
		if maxIn.LT(params.AmountIn.Amount) {
			if errors.Is(err, dextypes.ErrFoKLimitOrderNotFilled) {
				s.T().Skip()
			}
		}
	}
	s.NoError(err)
}

func TestPlaceLimitOrderTaker(t *testing.T) {
	testParams := []testParams{
		// state
		{field: "LiquidityType", states: []string{LO, LP, LOLP}},
		{field: "TicksDistribution", states: []string{"1", "2", "10"}}, // these are not the ticks but the amount of ticks we want to distribute liquidity over
		{field: "OrderType", states: []string{
			dextypes.LimitOrderType_name[int32(dextypes.LimitOrderType_FILL_OR_KILL)],
			dextypes.LimitOrderType_name[int32(dextypes.LimitOrderType_IMMEDIATE_OR_CANCEL)],
		}},
		// msg
		{field: "AmountIn", states: []string{FiftyPCT, TwoHundredPct}},
		{field: "MaxAmountOut", states: []string{ZeroPCT, FiftyPCT, HundredPct, TwoHundredPct}},
		{field: "LimitPrice", states: []string{LOWSELLPRICE, AVGSELLPRICE, HIGHSELLPRICE}},
	}
	testCasesRaw := generatePermutations(testParams)
	testCases := hydrateAllPlaceLOTakerTestCases(testCasesRaw)

	s := new(DexStateTestSuite)
	s.SetT(t)
	s.SetupTest()

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s.SetT(t)
			tc.printTestInfo(t)

			s.setupLoTakerState(tc)
			s.fundCreatorBalanceDefault(tc.PairID)
			//

			resp, err := s.makePlaceTakerLO(s.creator, tc.AmountIn, tc.PairID.Token0, tc.LimitPrice.String(), dextypes.LimitOrderType(tc.OrderType), tc.MaxAmountOut)

			s.handleTakerErrors(tc, err)

			expIn, expOut := ExpectedInOut(tc)
			// TODO: fix rounding issues
			s.intsApproxEqual("", expIn, resp.CoinIn.Amount, 10)
			s.intsApproxEqual("", expOut, resp.TakerCoinOut.Amount, 10)

			s.True(
				tc.LimitPrice.MulInt(resp.CoinIn.Amount).TruncateInt().LTE(resp.TakerCoinOut.Amount),
			)

			if tc.MaxAmountOut != nil {
				s.True(resp.TakerCoinOut.Amount.LTE(*tc.MaxAmountOut))
			}

			if tc.OrderType == int32(dextypes.LimitOrderType_FILL_OR_KILL) {
				// we should fill either AmountIn or MaxAmountOut
				s.Condition(func() bool {
					if tc.MaxAmountOut != nil {
						return resp.TakerCoinOut.Amount.Sub(*tc.MaxAmountOut).Abs().LTE(math.NewInt(1)) || resp.CoinIn.Amount.Sub(tc.AmountIn.Amount).Abs().LTE(math.NewInt(1))
					}
					return resp.CoinIn.Amount.Sub(tc.AmountIn.Amount).Abs().LTE(math.NewInt(1))
				})
			}
		})
	}
}
