package dex_state_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
	"github.com/stretchr/testify/require"
)

type depositTestParams struct {
	SharedParams
	// State Conditions
	ExistingShareHolders  string
	LiquidityDistribution LiquidityDistribution
	PoolValueIncrease     LiquidityDistribution
	// Message Variants
	DisableAutoswap bool
	FailTxOnBEL     bool
	DepositAmounts  LiquidityDistribution
}

func (s *DexStateTestSuite) setupDepositState(params depositTestParams) {
	liquidityDistr := params.LiquidityDistribution

	switch params.ExistingShareHolders {
	case None:
		break
	case Creator:
		s.makeDepositSuccess(s.creator, liquidityDistr, false)
	case OneOther:
		s.makeDepositSuccess(s.alice, liquidityDistr, false)
	case OneOtherAndCreator:
		liqDistrArr := splitLiquidityDistribution(liquidityDistr, 2)
		s.makeDepositSuccess(s.creator, liqDistrArr[1], false)
		s.makeDepositSuccess(s.alice, liqDistrArr[0], false)
	}

	// handle pool value increase

	if !params.PoolValueIncrease.empty() {
		// Increase the value of the pool. This is analogous to a pool being swapped through
		pool, found := s.App.DexKeeper.GetPool(s.Ctx, params.PairID, params.Tick, params.Fee)
		s.True(found, "Pool not found")

		pool.LowerTick0.ReservesMakerDenom = pool.LowerTick0.ReservesMakerDenom.Add(params.PoolValueIncrease.TokenA.Amount)
		pool.UpperTick1.ReservesMakerDenom = pool.UpperTick1.ReservesMakerDenom.Add(params.PoolValueIncrease.TokenB.Amount)
		s.App.DexKeeper.SetPool(s.Ctx, pool)
	}
}

func CalcTotalPreDepositLiquidity(params depositTestParams) LiquidityDistribution {
	return LiquidityDistribution{
		TokenA: params.LiquidityDistribution.TokenA.Add(params.PoolValueIncrease.TokenA),
		TokenB: params.LiquidityDistribution.TokenB.Add(params.PoolValueIncrease.TokenB),
	}
}

func CalcDepositOutput(params depositTestParams) (resultAmountA, resultAmountB math.Int) {
	depositA := params.DepositAmounts.TokenA.Amount
	depositB := params.DepositAmounts.TokenB.Amount

	existingLiquidity := CalcTotalPreDepositLiquidity(params)
	existingA := existingLiquidity.TokenA.Amount
	existingB := existingLiquidity.TokenB.Amount

	switch {
	//Pool is empty can deposit full amounts
	case existingA.IsZero() && existingB.IsZero():
		return depositA, depositB
	// Pool only has TokenB, can deposit all of depositB
	case existingA.IsZero():
		return math.ZeroInt(), depositB
	// Pool only has TokenA, can deposit all of depositA
	case existingB.IsZero():
		return depositA, math.ZeroInt()
	// Pool has a ratio of A and B, deposit must match this ratio
	case existingA.IsPositive() && existingB.IsPositive():
		targetRatioA := math.LegacyNewDecFromInt(existingA).Quo(math.LegacyNewDecFromInt(existingB))
		maxAmountA := math.LegacyNewDecFromInt(depositB).Mul(targetRatioA).TruncateInt()
		resultAmountA = math.MinInt(depositA, maxAmountA)
		targetRatioB := math.LegacyOneDec().Quo(targetRatioA)
		maxAmountB := math.LegacyNewDecFromInt(depositA).Mul(targetRatioB).TruncateInt()
		resultAmountB = math.MinInt(depositB, maxAmountB)

		return resultAmountA, resultAmountB
	default:
		panic("unhandled deposit calc case")
	}
}

func calcCurrentShareValue(params depositTestParams) math_utils.PrecDec {
	initialValueA := params.LiquidityDistribution.TokenA.Amount
	initialValueB := params.LiquidityDistribution.TokenB.Amount

	existingShares := calcDepositValueAsToken0(params.Tick, initialValueA, initialValueB).TruncateInt()
	if existingShares.IsZero() {
		return math_utils.OnePrecDec()
	}

	totalValueA := initialValueA.Add(params.PoolValueIncrease.TokenA.Amount)
	totalValueB := initialValueB.Add(params.PoolValueIncrease.TokenB.Amount)

	totalPreDepositValue := calcDepositValueAsToken0(params.Tick, totalValueA, totalValueB)
	currentShareValue := math_utils.NewPrecDecFromInt(existingShares).Quo(totalPreDepositValue)

	return currentShareValue
}

func calcDepositValue(params depositTestParams, depositAmount0, depositAmount1 math.Int) math_utils.PrecDec {
	rawValueDeposit := calcDepositValueAsToken0(params.Tick, depositAmount0, depositAmount1)

	return rawValueDeposit
}

func calcAutoSwapResidualValue(params depositTestParams, residual0, residual1 math.Int) math_utils.PrecDec {
	swapFeeDeduction := dextypes.MustCalcPrice(int64(params.Fee))

	switch {
	// We must autoswap TokenA
	case residual0.IsPositive() && residual1.IsPositive():
		panic("residual0 and residual1 cannot both be positive")
	case residual0.IsPositive():
		return swapFeeDeduction.MulInt(residual0)
	case residual1.IsPositive():
		price1To0CenterTick := dextypes.MustCalcPrice(params.Tick)
		token1AsToken0 := price1To0CenterTick.MulInt(residual1)
		return swapFeeDeduction.Mul(token1AsToken0)
	default:
		panic("residual0 and residual1 cannot both be zero")

	}
}

func calcExpectedDepositAmounts(params depositTestParams) (tokenAAmount, tokenBAmount, sharesIssued math.Int) {

	amountAWithoutAutoswap, amountBWithoutAutoswap := CalcDepositOutput(params)

	sharesIssuedWithoutAutoswap := calcDepositValue(params, amountAWithoutAutoswap, amountBWithoutAutoswap)

	residualA := params.DepositAmounts.TokenA.Amount.Sub(amountAWithoutAutoswap)
	residualB := params.DepositAmounts.TokenB.Amount.Sub(amountBWithoutAutoswap)

	autoswapSharesIssued := math_utils.ZeroPrecDec()
	if !params.DisableAutoswap && (residualA.IsPositive() || residualB.IsPositive()) {
		autoswapSharesIssued = calcAutoSwapResidualValue(params, residualA, residualB)
		tokenAAmount = params.DepositAmounts.TokenA.Amount
		tokenBAmount = params.DepositAmounts.TokenB.Amount
	} else {
		tokenAAmount = amountAWithoutAutoswap
		tokenBAmount = amountBWithoutAutoswap
	}

	totalDepositValue := autoswapSharesIssued.Add(sharesIssuedWithoutAutoswap)
	currentShareValue := calcCurrentShareValue(params)
	sharesIssued = totalDepositValue.Mul(currentShareValue).TruncateInt()

	return tokenAAmount, tokenBAmount, sharesIssued
}

func (s *DexStateTestSuite) handleBaseFailureCases(params depositTestParams, err error) {
	currentLiquidity := CalcTotalPreDepositLiquidity(params)
	// cannot deposit single sided liquidity into a non-empty pool if you are missing one of the tokens in the pool
	if !currentLiquidity.empty() {
		if (!params.DepositAmounts.hasTokenA() && currentLiquidity.hasTokenA()) || (!params.DepositAmounts.hasTokenB() && currentLiquidity.hasTokenB()) {
			s.ErrorIs(err, dextypes.ErrZeroTrueDeposit)
			s.T().Skip("Ending test due to expected error")
		}
	}
}

func HydrateDepositTestCase(params map[string]string, pairID *dextypes.PairID) depositTestParams {
	existingShareHolders := params["ExistingShareHolders"]
	var liquidityDistribution LiquidityDistribution

	if existingShareHolders == None {
		liquidityDistribution = parseLiquidityDistribution(TokenA0TokenB0, pairID)
	} else {
		liquidityDistribution = parseLiquidityDistribution(params["LiquidityDistribution"], pairID)
	}

	var valueIncrease LiquidityDistribution
	if liquidityDistribution.empty() {
		valueIncrease = parseLiquidityDistribution(TokenA0TokenB0, pairID)
	} else {
		valueIncrease = parseLiquidityDistribution(params["PoolValueIncrease"], pairID)
	}

	return depositTestParams{
		ExistingShareHolders:  existingShareHolders,
		LiquidityDistribution: liquidityDistribution,
		DisableAutoswap:       parseBool(params["DisableAutoswap"]),
		PoolValueIncrease:     valueIncrease,
		DepositAmounts:        parseLiquidityDistribution(params["DepositAmounts"], pairID),
		SharedParams:          DefaultSharedParams,
	}
}

func HydrateAllDepositTestCases(paramsList []map[string]string) []depositTestParams {
	allTCs := make([]depositTestParams, 0)
	for i, paramsRaw := range paramsList {
		pairID := generatePairID(i)
		tc := HydrateDepositTestCase(paramsRaw, pairID)
		tc.PairID = pairID
		allTCs = append(allTCs, tc)
	}

	// De-dupe test cases hydration creates some duplicates
	return removeDuplicateTests(allTCs)
}

func TestDeposit(t *testing.T) {
	testParams := []testParams{
		{field: "ExistingShareHolders", states: []string{None, Creator, OneOther}},
		{field: "LiquidityDistribution", states: []string{
			TokenA0TokenB1,
			TokenA0TokenB2,
			TokenA1TokenB0,
			TokenA1TokenB1,
			TokenA1TokenB2,
			TokenA2TokenB0,
			TokenA2TokenB1,
			TokenA2TokenB2,
		}},
		{field: "DisableAutoswap", states: []string{True, False}},
		{field: "PoolValueIncrease", states: []string{TokenA0TokenB0, TokenA1TokenB0, TokenA0TokenB1}},
		// {field: "FailTxOnBEL", states: []string{True, False}}, // I don't think this needs to be tested
		{field: "DepositAmounts", states: []string{
			TokenA0TokenB1,
			TokenA0TokenB2,
			TokenA1TokenB1,
			TokenA1TokenB2,
			TokenA2TokenB2,
		}},
	}
	testCasesRaw := generatePermutations(testParams)
	testCases := HydrateAllDepositTestCases(testCasesRaw)

	s := new(DexStateTestSuite)
	s.SetT(t)
	s.SetupTest(t)

	for _, tc := range testCases {
		testName := fmt.Sprintf("%v", tc)
		t.Run(testName, func(t *testing.T) {
			s.SetT(t)

			s.setupDepositState(tc)

			poolID, found := s.App.DexKeeper.GetPoolIDByParams(s.Ctx, tc.PairID, tc.Tick, tc.Fee)

			if tc.ExistingShareHolders == None {
				// This is the ID that will be used when the pool is created
				poolID = s.App.DexKeeper.GetPoolCount(s.Ctx)
			} else {
				require.True(t, found, "Pool not found after deposit")
			}

			poolDenom := dextypes.NewPoolDenom(poolID)

			existingSharesOwned := s.App.BankKeeper.GetBalance(s.Ctx, s.creator, poolDenom)

			// Do the actual deposit
			resp, err := s.makeDeposit(s.creator, tc.DepositAmounts, tc.DisableAutoswap)

			s.handleBaseFailureCases(tc, err)
			s.NoError(err)

			expectedDepositA, expectedDepositB, expectedShares := calcExpectedDepositAmounts(tc)

			//Check that response is correct
			s.intsEqual("Response Deposit0", expectedDepositA, resp.Reserve0Deposited[0])
			s.intsEqual("Response Deposit1", expectedDepositB, resp.Reserve1Deposited[0])

			newSharesOwned := s.App.BankKeeper.GetBalance(s.Ctx, s.creator, poolDenom)
			sharesIssued := newSharesOwned.Sub(existingSharesOwned)
			s.intsEqual("Shares Issued", expectedShares, sharesIssued.Amount)

			// TODO: balance checks for tokens
			// TODO: maybe check actual dex state
		})

	}

}
