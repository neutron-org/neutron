package dex_state_test

import (
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	dextypes "github.com/neutron-org/neutron/v7/x/dex/types"
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
		coins := sdk.NewCoins(liquidityDistr.TokenA.CeilToCoin(), liquidityDistr.TokenB.CeilToCoin())
		s.FundAcc(s.creator, coins)

		s.makeDepositSuccess(s.creator, liquidityDistr, false)
	case OneOther:
		coins := sdk.NewCoins(liquidityDistr.TokenA.CeilToCoin(), liquidityDistr.TokenB.CeilToCoin())
		s.FundAcc(s.alice, coins)

		s.makeDepositSuccess(s.alice, liquidityDistr, false)
	case OneOtherAndCreator:
		splitLiqDistrArr := splitLiquidityDistribution(liquidityDistr, 2)

		coins := sdk.NewCoins(splitLiqDistrArr.TokenA.CeilToCoin(), splitLiqDistrArr.TokenB.CeilToCoin())
		s.FundAcc(s.creator, coins)

		coins = sdk.NewCoins(splitLiqDistrArr.TokenA.CeilToCoin(), splitLiqDistrArr.TokenB.CeilToCoin())
		s.FundAcc(s.alice, coins)

		s.makeDepositSuccess(s.creator, splitLiqDistrArr, false)
		s.makeDepositSuccess(s.alice, splitLiqDistrArr, false)
	}

	// handle pool value increase

	if !params.PoolValueIncrease.empty() {
		// Increase the value of the pool. This is analogous to a pool being swapped through
		pool, found := s.App.DexKeeper.GetPool(s.Ctx, params.PairID, params.Tick, params.Fee)
		s.True(found, "Pool not found")

		pool.LowerTick0.DecReservesMakerDenom = pool.LowerTick0.DecReservesMakerDenom.Add(params.PoolValueIncrease.TokenA.Amount)
		pool.UpperTick1.DecReservesMakerDenom = pool.UpperTick1.DecReservesMakerDenom.Add(params.PoolValueIncrease.TokenB.Amount)
		s.App.DexKeeper.UpdatePool(s.Ctx, pool)

		// Add fund dex with the additional balance
		err := s.App.BankKeeper.MintCoins(s.Ctx, dextypes.ModuleName, sdk.NewCoins(params.PoolValueIncrease.TokenA.CeilToCoin(), params.PoolValueIncrease.TokenB.CeilToCoin()))
		s.NoError(err)
	}
}

func CalcTotalPreDepositLiquidity(params depositTestParams) LiquidityDistribution {
	return LiquidityDistribution{
		TokenA: params.ExistingLiquidityDistribution.TokenA.Add(params.PoolValueIncrease.TokenA),
		TokenB: params.ExistingLiquidityDistribution.TokenB.Add(params.PoolValueIncrease.TokenB),
	}
}

func CalcDepositAmountNoAutoswap(params depositTestParams) (resultAmountA, resultAmountB math_utils.PrecDec) {
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
		return math_utils.ZeroPrecDec(), depositB
	// Pool only has TokenA, can deposit all of depositA
	case existingB.IsZero():
		return depositA, math_utils.ZeroPrecDec()
	// Pool has a ratio of A and B, deposit must match this ratio
	case existingA.IsPositive() && existingB.IsPositive():
		maxAmountA := depositB.Mul(existingA).Quo(existingB)
		resultAmountA = math_utils.MinPrecDec(depositA, maxAmountA)
		maxAmountB := depositA.Mul(existingB).Quo(existingA)
		resultAmountB = math_utils.MinPrecDec(depositB, maxAmountB)

		return resultAmountA, resultAmountB
	default:
		panic("unhandled deposit calc case")
	}
}

func calcCurrentShareValue(params depositTestParams, existingValue math_utils.PrecDec) math_utils.PrecDec {
	initialValueA := params.ExistingLiquidityDistribution.TokenA.Amount
	initialValueB := params.ExistingLiquidityDistribution.TokenB.Amount

	existingShares := calcDepositValueAsToken0(params.Tick, initialValueA, initialValueB)
	if existingShares.IsZero() {
		return math_utils.OnePrecDec()
	}

	currentShareValue := existingShares.Quo(existingValue)

	return currentShareValue
}

func calcAutoswapAmount(params depositTestParams) (swapAmountA, swapAmountB math_utils.PrecDec) {
	existingLiquidity := CalcTotalPreDepositLiquidity(params)
	existingA := existingLiquidity.TokenA.Amount
	existingB := existingLiquidity.TokenB.Amount
	depositAmountA := params.DepositAmounts.TokenA.Amount
	depositAmountB := params.DepositAmounts.TokenB.Amount
	price1To0 := dextypes.MustCalcPrice(-params.Tick)
	if existingA.IsZero() && existingB.IsZero() {
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec()
	}

	existingADec := existingA
	existingBDec := existingB
	// swapAmount = (reserves0*depositAmount1 - reserves1*depositAmount0) / (price * reserves1  + reserves0)
	swapAmount := existingADec.Mul(depositAmountB).Sub(existingBDec.Mul(depositAmountA)).
		Quo(existingADec.Add(existingBDec.Quo(price1To0)))

	switch {
	case swapAmount.IsZero(): // nothing to be swapped
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec()

	case swapAmount.IsPositive(): // Token1 needs to be swapped
		return math_utils.ZeroPrecDec(), swapAmount

	default: // Token0 needs to be swapped
		amountSwappedAs1 := swapAmount.Neg()

		amountSwapped0 := amountSwappedAs1.Quo(price1To0)
		return amountSwapped0, math_utils.ZeroPrecDec()
	}
}

func calcExpectedDepositAmounts(params depositTestParams) (tokenAAmount, tokenBAmount math_utils.PrecDec, sharesIssued math.Int) {
	var depositValueAsToken0 math_utils.PrecDec
	var inAmountA math_utils.PrecDec
	var inAmountB math_utils.PrecDec

	existingLiquidity := CalcTotalPreDepositLiquidity(params)
	existingA := existingLiquidity.TokenA.Amount
	existingB := existingLiquidity.TokenB.Amount
	existingValueAsToken0 := calcDepositValueAsToken0(params.Tick, existingA, existingB)

	if params.DisableAutoswap {
		inAmountA, inAmountB = CalcDepositAmountNoAutoswap(params)
		depositValueAsToken0 = calcDepositValueAsToken0(params.Tick, inAmountA, inAmountB)

		shareValue := calcCurrentShareValue(params, existingValueAsToken0)
		sharesIssued = depositValueAsToken0.Mul(shareValue).TruncateInt()

		return inAmountA, inAmountB, sharesIssued
	} // else
	autoSwapAmountA, autoswapAmountB := calcAutoswapAmount(params)
	autoswapValueAsToken0 := calcDepositValueAsToken0(params.Tick, autoSwapAmountA, autoswapAmountB)

	autoswapFeeAsPrice := dextypes.MustCalcPrice(-int64(params.Fee)) //nolint:gosec
	autoswapFeePct := math_utils.OnePrecDec().Sub(autoswapFeeAsPrice)
	autoswapFee := autoswapValueAsToken0.Mul(autoswapFeePct)

	inAmountA = params.DepositAmounts.TokenA.Amount
	inAmountB = params.DepositAmounts.TokenB.Amount

	fullDepositValueAsToken0 := calcDepositValueAsToken0(params.Tick, inAmountA, inAmountB)
	depositAmountMinusFee := fullDepositValueAsToken0.Sub(autoswapFee)
	currentValueWithAutoswapFee := existingValueAsToken0.Add(autoswapFee)
	shareValue := calcCurrentShareValue(params, currentValueWithAutoswapFee)

	sharesIssued = depositAmountMinusFee.Mul(shareValue).TruncateInt()

	return inAmountA, inAmountB, sharesIssued
}

func (s *DexStateTestSuite) handleBaseFailureCases(params depositTestParams, err error) {
	currentLiquidity := CalcTotalPreDepositLiquidity(params)
	// cannot deposit single sided liquidity into a non-empty pool if you are missing one of the tokens in the pool
	if !currentLiquidity.empty() && params.DisableAutoswap {
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
			s.Equal(expectedDepositA, resp.DecReserve0Deposited[0], "Response Deposit0")
			s.Equal(expectedDepositB, resp.DecReserve1Deposited[0], "Response Deposit1")

			expectedTotalShares := existingSharesOwned.Amount.Add(expectedShares)
			// For sanity check we use a slightly different share calculation. This can result in off by 1 error
			s.assertApproxCreatorBalance(poolDenom, expectedTotalShares)

			// Assert Creator Balance is correct
			expectedBalanceA := DefaultStartingBalanceInt.Sub(expectedDepositA.Ceil().TruncateInt())
			expectedBalanceB := DefaultStartingBalanceInt.Sub(expectedDepositB.Ceil().TruncateInt())
			s.assertCreatorBalance(tc.PairID.Token0, expectedBalanceA)
			s.assertCreatorBalance(tc.PairID.Token1, expectedBalanceB)

			// Assert dex state is correct
			dexBalanceBeforeDeposit := CalcTotalPreDepositLiquidity(tc)
			expectedDexBalanceA := dexBalanceBeforeDeposit.TokenA.Amount.Add(expectedDepositA)
			expectedDexBalanceB := dexBalanceBeforeDeposit.TokenB.Amount.Add(expectedDepositB)
			s.assertPoolBalance(tc.PairID, tc.Tick, tc.Fee, expectedDexBalanceA, expectedDexBalanceB)
			s.assertDexBalance(tc.PairID.Token0, expectedDexBalanceA.Ceil().TruncateInt())
			s.assertDexBalance(tc.PairID.Token1, expectedDexBalanceB.Ceil().TruncateInt())
		})
	}

	s.TearDownTest()
}
