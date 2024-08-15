package dex_state_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v4/testutil/apptesting"
	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	dexkeeper "github.com/neutron-org/neutron/v4/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
	"github.com/stretchr/testify/require"
)

// Shared Setup Code //////////////////////////////////////////////////////////

// Bools
const (
	True  string = "True"
	False        = "False"
)

// Percents
const (
	ZeroPCT    string = "0"
	FiftyPCT          = "50"
	HundredPct        = "100"
)

// ExistingShareHolders
const (
	None               string = "None"
	Creator                   = "Creator"
	OneOther                  = "OneOther"
	OneOtherAndCreator        = "OneOtherAndCreator"
)

// LiquidityDistribution
const (
	TokenA0TokenB0 string = "TokenA0TokenB0"
	TokenA0TokenB1        = "TokenA0TokenB1"
	TokenA0TokenB2        = "TokenA0TokenB2"
	TokenA1TokenB0        = "TokenA1TokenB0"
	TokenA1TokenB1        = "TokenA1TokenB1"
	TokenA1TokenB2        = "TokenA1TokenB2"
	TokenA2TokenB0        = "TokenA2TokenB0"
	TokenA2TokenB1        = "TokenA2TokenB1"
	TokenA2TokenB2        = "TokenA2TokenB2"
)

const (
	BaseTokenAmount = 1_000_000
	DefaultTick     = 0
	DefaultFee      = 1
)

var BaseTokenAmountInt = math.NewInt(BaseTokenAmount)

type testParams struct {
	field  string
	states []string
}

type LiquidityDistribution struct {
	TokenA sdk.Coin
	TokenB sdk.Coin
}

func (l LiquidityDistribution) doubleSided() bool {
	return l.TokenA.Amount.IsPositive() && l.TokenB.Amount.IsPositive()
}

func (l LiquidityDistribution) empty() bool {
	return l.TokenA.Amount.IsZero() && l.TokenB.Amount.IsZero()
}

func (l LiquidityDistribution) singleSided() bool {
	return !l.doubleSided() && !l.empty()
}

func (l LiquidityDistribution) hasTokenA() bool {
	return l.TokenA.Amount.IsPositive()
}

func (l LiquidityDistribution) hasTokenB() bool {
	return l.TokenB.Amount.IsPositive()
}

type SharedParams struct {
	Tick   int64
	Fee    uint64
	PairID dextypes.PairID
}

var DefaultSharedParams SharedParams = SharedParams{
	Tick: DefaultTick,
	Fee:  DefaultFee,
}

func (s *DexStateTestSuite) intsEqual(field string, expected, actual math.Int) {
	s.True(actual.Equal(expected), "For %v: Expected %v Got %v", field, expected, actual)
}

func parseLiquidityDistribution(liquidityDistribution string) LiquidityDistribution {
	switch liquidityDistribution {
	case TokenA0TokenB0:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.ZeroInt()), TokenB: sdk.NewCoin("TokenB", math.ZeroInt())}
	case TokenA0TokenB1:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.ZeroInt()), TokenB: sdk.NewCoin("TokenB", math.NewInt(1).Mul(BaseTokenAmountInt))}
	case TokenA0TokenB2:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.ZeroInt()), TokenB: sdk.NewCoin("TokenB", math.NewInt(2).Mul(BaseTokenAmountInt))}
	case TokenA1TokenB0:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.ZeroInt())}
	case TokenA1TokenB1:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(1).Mul(BaseTokenAmountInt))}
	case TokenA1TokenB2:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(2).Mul(BaseTokenAmountInt))}
	case TokenA2TokenB0:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(2).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.ZeroInt())}
	case TokenA2TokenB1:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(2).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(1).Mul(BaseTokenAmountInt))}
	case TokenA2TokenB2:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(2).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(2).Mul(BaseTokenAmountInt))}
	default:
		panic("invalid liquidity distribution")
	}
}

func parseBool(b string) bool {
	switch b {
	case True:
		return true
	case False:
		return false
	default:
		panic("invalid bool")

	}
}

func splitLiquidityDistribution(liquidityDistribution LiquidityDistribution, n int64) []LiquidityDistribution {
	nInt := math.NewInt(n)
	amount0 := liquidityDistribution.TokenA.Amount.Quo(nInt)
	amount1 := liquidityDistribution.TokenB.Amount.Quo(nInt)

	result := make([]LiquidityDistribution, n)
	for i := range n {

		result[i] = LiquidityDistribution{
			TokenA: sdk.NewCoin(liquidityDistribution.TokenA.Denom, amount0),
			TokenB: sdk.NewCoin(liquidityDistribution.TokenB.Denom, amount1),
		}

	}

	return result

}

func cloneMap(original map[string]string) map[string]string {
	// Create a new map
	cloned := make(map[string]string)
	// Copy each key-value pair from the original map to the new map
	for key, value := range original {
		cloned[key] = value
	}

	return cloned
}

func generatePermutations(testStates []testParams) []map[string]string {
	result := make([]map[string]string, 0)

	var generate func(int, map[string]string)
	generate = func(index int, current map[string]string) {

		// Base Case
		if index == len(testStates) {
			result = append(result, current)
			return
		}

		// Iterate through all possible values and create new states
		for _, value := range testStates[index].states {
			fieldName := testStates[index].field
			temp := cloneMap(current)
			temp[fieldName] = value
			generate(index+1, temp)
		}

	}
	emptyMap := make(map[string]string)
	generate(0, emptyMap)

	return result
}

func (s *DexStateTestSuite) makeDeposit(addr sdk.AccAddress, depositAmts LiquidityDistribution, disableAutoSwap bool) (*dextypes.MsgDepositResponse, error) {
	coins := sdk.NewCoins(depositAmts.TokenA, depositAmts.TokenB)
	s.FundAcc(addr, coins)

	return s.msgServer.Deposit(s.Ctx, &dextypes.MsgDeposit{

		Creator:         addr.String(),
		Receiver:        addr.String(),
		TokenA:          depositAmts.TokenA.Denom,
		TokenB:          depositAmts.TokenB.Denom,
		AmountsA:        []math.Int{depositAmts.TokenA.Amount},
		AmountsB:        []math.Int{depositAmts.TokenB.Amount},
		TickIndexesAToB: []int64{DefaultTick},
		Fees:            []uint64{DefaultFee},
		Options:         []*dextypes.DepositOptions{{DisableAutoswap: disableAutoSwap}},
	})
}

func (s *DexStateTestSuite) makeDepositSuccess(addr sdk.AccAddress, depositAmts LiquidityDistribution, disableAutoSwap bool) *dextypes.MsgDepositResponse {
	resp, err := s.makeDeposit(addr, depositAmts, disableAutoSwap)
	s.NoError(err)

	return resp
}

type DexStateTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer dextypes.MsgServer
	creator   sdk.AccAddress
	alice     sdk.AccAddress
}

func (s *DexStateTestSuite) SetupTest(t *testing.T) {
	s.Setup()
	s.creator = sdk.MustAccAddressFromBech32(sample.AccAddress())
	s.alice = sdk.MustAccAddressFromBech32(sample.AccAddress())

	s.msgServer = dexkeeper.NewMsgServerImpl(s.App.DexKeeper)
}

// Deposit State Test /////////////////////////////////////////////////////////
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
		pool, found := s.App.DexKeeper.GetPool(s.Ctx, &params.PairID, params.Tick, params.Fee)
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

func calcDepositValueAsToken0(tick int64, amount0, amount1 math.Int) math_utils.PrecDec {
	price1To0CenterTick := dextypes.MustCalcPrice(tick)
	amount1ValueAsToken0 := price1To0CenterTick.MulInt(amount1)
	depositValue := amount1ValueAsToken0.Add(math_utils.NewPrecDecFromInt(amount0))

	return depositValue

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

func HydrateDepositTestCase(params map[string]string) depositTestParams {
	existingShareHolders := params["ExistingShareHolders"]
	var liquidityDistribution LiquidityDistribution

	if existingShareHolders == None {
		liquidityDistribution = parseLiquidityDistribution(TokenA0TokenB0)
	} else {
		liquidityDistribution = parseLiquidityDistribution(params["LiquidityDistribution"])
	}

	var valueIncrease LiquidityDistribution
	if liquidityDistribution.empty() {
		valueIncrease = parseLiquidityDistribution(TokenA0TokenB0)
	} else {
		valueIncrease = parseLiquidityDistribution(params["PoolValueIncrease"])
	}

	sharedParams := DefaultSharedParams
	sharedParams.PairID = dextypes.PairID{Token0: "TokenA", Token1: "TokenB"}

	return depositTestParams{
		ExistingShareHolders:  existingShareHolders,
		LiquidityDistribution: liquidityDistribution,
		DisableAutoswap:       parseBool(params["DisableAutoswap"]),
		PoolValueIncrease:     valueIncrease,
		DepositAmounts:        parseLiquidityDistribution(params["DepositAmounts"]),
		SharedParams:          sharedParams,
	}
}

func HydrateAllDepositTestCases(paramsList []map[string]string) []depositTestParams {
	allTCs := make([]depositTestParams, 0)
	for _, paramsRaw := range paramsList {
		allTCs = append(allTCs, HydrateDepositTestCase(paramsRaw))
	}

	result := make([]depositTestParams, 0)

	// De-dupe test cases hydration creates some duplicates
	seenTCs := make(map[string]bool)
	for _, tc := range allTCs {
		tcStr := fmt.Sprintf("%v", tc)
		if _, ok := seenTCs[tcStr]; !ok {
			result = append(result, tc)
		}
		seenTCs[tcStr] = true
	}

	return result
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

	for _, tc := range testCases {
		testName := fmt.Sprintf("%v", tc)
		t.Run(testName, func(t *testing.T) {
			s := new(DexStateTestSuite)
			s.SetT(t)
			// TODO: we don't want to rebuild the app for every test. Instead we should just use new pools
			s.SetupTest(t)

			s.setupDepositState(tc)

			poolID, found := s.App.DexKeeper.GetPoolIDByParams(s.Ctx, &tc.PairID, tc.Tick, tc.Fee)

			if tc.ExistingShareHolders == None {
				poolID = 0
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
