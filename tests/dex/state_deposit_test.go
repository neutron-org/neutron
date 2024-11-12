package dex_state_test

import (
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	math_utils "github.com/neutron-org/neutron/v5/utils/math"
	dextypes "github.com/neutron-org/neutron/v5/x/dex/types"
)

type DepositState struct {
	SharedParams
	// State Conditions
	ExistingShareHolders          string
	ExistingLiquidityDistribution LiquidityDistribution
	PoolValueIncrease             LiquidityDistribution
}

type depositTestParams struct {
	DepositState
	// Message Variants
	DisableAutoswap bool
	FailTxOnBEL     bool
	DepositAmounts  LiquidityDistribution
}

func (p depositTestParams) printTestInfo(t *testing.T) {
	t.Logf(`
		Existing Shareholders: %s
		Existing Liquidity Distribution: %v
		Pool Value Increase: %v
		Disable Autoswap: %t
		Fail Tx on BEL: %t
		Deposit Amounts: %v`,
		p.ExistingShareHolders,
		p.ExistingLiquidityDistribution,
		p.PoolValueIncrease,
		p.DisableAutoswap,
		p.FailTxOnBEL,
		p.DepositAmounts,
	)
}

func (s *DexStateTestSuite) setupDepositState(params DepositState) {
	// NOTE: for setup we know the deposit will be completely used so we fund the accounts before the deposit
	// so the expected account balance is unaffected.
	liquidityDistr := params.ExistingLiquidityDistribution

	switch params.ExistingShareHolders {
	case None:
		break
	case Creator:
		coins := sdk.NewCoins(liquidityDistr.TokenA, liquidityDistr.TokenB)
		s.FundAcc(s.creator, coins)

		s.makeDepositSuccess(s.creator, liquidityDistr, false)
	case OneOther:
		coins := sdk.NewCoins(liquidityDistr.TokenA, liquidityDistr.TokenB)
		s.FundAcc(s.alice, coins)

		s.makeDepositSuccess(s.alice, liquidityDistr, false)
	case OneOtherAndCreator:
		splitLiqDistrArr := splitLiquidityDistribution(liquidityDistr, 2)

		coins := sdk.NewCoins(splitLiqDistrArr.TokenA, splitLiqDistrArr.TokenB)
		s.FundAcc(s.creator, coins)

		coins = sdk.NewCoins(splitLiqDistrArr.TokenA, splitLiqDistrArr.TokenB)
		s.FundAcc(s.alice, coins)

		s.makeDepositSuccess(s.creator, splitLiqDistrArr, false)
		s.makeDepositSuccess(s.alice, splitLiqDistrArr, false)
	}

	// handle pool value increase

	if !params.PoolValueIncrease.empty() {
		// Increase the value of the pool. This is analogous to a pool being swapped through
		pool, found := s.App.DexKeeper.GetPool(s.Ctx, params.PairID, params.Tick, params.Fee)
		s.True(found, "Pool not found")

		pool.LowerTick0.ReservesMakerDenom = pool.LowerTick0.ReservesMakerDenom.Add(params.PoolValueIncrease.TokenA.Amount)
		pool.UpperTick1.ReservesMakerDenom = pool.UpperTick1.ReservesMakerDenom.Add(params.PoolValueIncrease.TokenB.Amount)
		s.App.DexKeeper.UpdatePool(s.Ctx, pool)

		// Add fund dex with the additional balance
		err := s.App.BankKeeper.MintCoins(s.Ctx, dextypes.ModuleName, sdk.NewCoins(params.PoolValueIncrease.TokenA, params.PoolValueIncrease.TokenB))
		s.NoError(err)
	}
}

func CalcTotalPreDepositLiquidity(params depositTestParams) LiquidityDistribution {
	return LiquidityDistribution{
		TokenA: params.ExistingLiquidityDistribution.TokenA.Add(params.PoolValueIncrease.TokenA),
		TokenB: params.ExistingLiquidityDistribution.TokenB.Add(params.PoolValueIncrease.TokenB),
	}
}

func CalcDepositOutput(params depositTestParams) (resultAmountA, resultAmountB math.Int) {
	depositA := params.DepositAmounts.TokenA.Amount
	depositB := params.DepositAmounts.TokenB.Amount

	existingLiquidity := CalcTotalPreDepositLiquidity(params)
	existingA := existingLiquidity.TokenA.Amount
	existingB := existingLiquidity.TokenB.Amount

	switch {
	// Pool is empty can deposit full amounts
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
		maxAmountA := math.LegacyNewDecFromInt(depositB).MulInt(existingA).QuoInt(existingB).TruncateInt()
		resultAmountA = math.MinInt(depositA, maxAmountA)
		maxAmountB := math.LegacyNewDecFromInt(depositA).MulInt(existingB).QuoInt(existingA).TruncateInt()
		resultAmountB = math.MinInt(depositB, maxAmountB)

		return resultAmountA, resultAmountB
	default:
		panic("unhandled deposit calc case")
	}
}

func calcCurrentShareValue(params depositTestParams) math_utils.PrecDec {
	initialValueA := params.ExistingLiquidityDistribution.TokenA.Amount
	initialValueB := params.ExistingLiquidityDistribution.TokenB.Amount

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

func calcAutoSwapResidualValue(params depositTestParams, residual0, residual1 math.Int) math_utils.PrecDec {
	swapFeeDeduction := dextypes.MustCalcPrice(int64(params.Fee))

	switch {
	// We must autoswap TokenA
	case residual0.IsPositive() && residual1.IsPositive():
		panic("residual0 and residual1 cannot both be positive")
	case residual0.IsPositive():
		return swapFeeDeduction.MulInt(residual0)
	case residual1.IsPositive():
		price1To0CenterTick := dextypes.MustCalcPrice(-1 * params.Tick)
		token1AsToken0 := price1To0CenterTick.MulInt(residual1)
		return swapFeeDeduction.Mul(token1AsToken0)
	default:
		panic("residual0 and residual1 cannot both be zero")

	}
}

func calcExpectedDepositAmounts(params depositTestParams) (tokenAAmount, tokenBAmount, sharesIssued math.Int) {
	amountAWithoutAutoswap, amountBWithoutAutoswap := CalcDepositOutput(params)

	sharesIssuedWithoutAutoswap := calcDepositValueAsToken0(params.Tick, amountAWithoutAutoswap, amountBWithoutAutoswap)

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

	s.NoError(err)
}

func hydrateDepositTestCase(params map[string]string, pairID *dextypes.PairID) depositTestParams {
	existingShareHolders := params["ExistingShareHolders"]
	var liquidityDistribution LiquidityDistribution

	if existingShareHolders == None {
		liquidityDistribution = parseLiquidityDistribution(TokenA0TokenB0, pairID)
	} else {
		liquidityDistribution = parseLiquidityDistribution(params["LiquidityDistribution"], pairID)
	}

	var valueIncrease LiquidityDistribution
	if liquidityDistribution.empty() {
		// Cannot increase value on empty pool
		valueIncrease = parseLiquidityDistribution(TokenA0TokenB0, pairID)
	} else {
		valueIncrease = parseLiquidityDistribution(params["PoolValueIncrease"], pairID)
	}

	return depositTestParams{
		DepositState: DepositState{
			ExistingShareHolders:          existingShareHolders,
			ExistingLiquidityDistribution: liquidityDistribution,
			PoolValueIncrease:             valueIncrease,
			SharedParams:                  DefaultSharedParams,
		},
		DepositAmounts:  parseLiquidityDistribution(params["DepositAmounts"], pairID),
		DisableAutoswap: parseBool(params["DisableAutoswap"]),
	}
}

func hydrateAllDepositTestCases(paramsList []map[string]string) []depositTestParams {
	allTCs := make([]depositTestParams, 0)
	for i, paramsRaw := range paramsList {
		pairID := generatePairID(i)
		tc := hydrateDepositTestCase(paramsRaw, pairID)
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
		{field: "DepositAmounts", states: []string{
			TokenA0TokenB1,
			TokenA0TokenB2,
			TokenA1TokenB1,
			TokenA1TokenB2,
			TokenA2TokenB2,
		}},
		// TODO: test over a list of Fees/Ticks
	}
	testCasesRaw := generatePermutations(testParams)
	testCases := hydrateAllDepositTestCases(testCasesRaw)

	s := new(DexStateTestSuite)
	s.SetT(t)
	s.SetupTest()

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s.SetT(t)
			tc.printTestInfo(t)

			s.setupDepositState(tc.DepositState)
			s.fundCreatorBalanceDefault(tc.PairID)

			poolID, found := s.App.DexKeeper.GetPoolIDByParams(s.Ctx, tc.PairID, tc.Tick, tc.Fee)
			if tc.ExistingShareHolders == None {
				// There is no pool yet. This is the ID that will be used when the pool is created
				poolID = s.App.DexKeeper.GetPoolCount(s.Ctx)
			} else {
				require.True(t, found, "Pool not found after deposit")
			}
			poolDenom := dextypes.NewPoolDenom(poolID)
			existingSharesOwned := s.App.BankKeeper.GetBalance(s.Ctx, s.creator, poolDenom)

			// Do the actual deposit
			resp, err := s.makeDepositDefault(s.creator, tc.DepositAmounts, tc.DisableAutoswap)

			// Assert that if there is an error it is expected
			s.handleBaseFailureCases(tc, err)

			expectedDepositA, expectedDepositB, expectedShares := calcExpectedDepositAmounts(tc)

			// Check that response is correct
			s.intsApproxEqual("Response Deposit0", expectedDepositA, resp.Reserve0Deposited[0], 1)
			s.intsApproxEqual("Response Deposit1", expectedDepositB, resp.Reserve1Deposited[0], 1)

			expectedTotalShares := existingSharesOwned.Amount.Add(expectedShares)
			s.assertCreatorBalance(poolDenom, expectedTotalShares)

			// Assert Creator Balance is correct
			expectedBalanceA := DefaultStartingBalanceInt.Sub(expectedDepositA)
			expectedBalanceB := DefaultStartingBalanceInt.Sub(expectedDepositB)
			s.assertCreatorBalance(tc.PairID.Token0, expectedBalanceA)
			s.assertCreatorBalance(tc.PairID.Token1, expectedBalanceB)

			// Assert dex state is correct
			dexBalanceBeforeDeposit := CalcTotalPreDepositLiquidity(tc)
			expectedDexBalanceA := dexBalanceBeforeDeposit.TokenA.Amount.Add(expectedDepositA)
			expectedDexBalanceB := dexBalanceBeforeDeposit.TokenB.Amount.Add(expectedDepositB)
			s.assertPoolBalance(tc.PairID, tc.Tick, tc.Fee, expectedDexBalanceA, expectedDexBalanceB)
			s.assertDexBalance(tc.PairID.Token0, expectedDexBalanceA)
			s.assertDexBalance(tc.PairID.Token1, expectedDexBalanceB)
		})
	}
}
