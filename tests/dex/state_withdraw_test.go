package dex_state_test

import (
	"fmt"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

type withdrawTestParams struct {
	// State Conditions
	DepositState
	// Message Variants
	SharesToRemoveAmm int64
}

func (p withdrawTestParams) printTestInfo(t *testing.T) {
	t.Logf(`
		Existing Shareholders: %s
		Existing Liquidity Distribution: %v
		Shares to remove: %v`,
		p.ExistingShareHolders,
		p.ExistingLiquidityDistribution,
		p.SharesToRemoveAmm,
	)
}

func (s *DexStateTestSuite) handleWithdrawFailureCases(params withdrawTestParams, err error) {
	if params.SharesToRemoveAmm == 0 {
		s.ErrorIs(err, dextypes.ErrZeroWithdraw)
	} else {
		s.NoError(err)
	}
}

func hydrateWithdrawTestCase(params map[string]string, pairID *dextypes.PairID) withdrawTestParams {
	existingShareHolders := params["ExistingShareHolders"]
	var liquidityDistribution LiquidityDistribution

	if existingShareHolders == None {
		liquidityDistribution = parseLiquidityDistribution(TokenA0TokenB0, pairID)
	} else {
		liquidityDistribution = parseLiquidityDistribution(params["LiquidityDistribution"], pairID)
	}

	sharesToRemove, err := strconv.ParseInt(params["SharesToRemoveAmm"], 10, 64)
	if err != nil {
		panic(fmt.Sprintln("invalid SharesToRemoveAmm", err))
	}

	var valueIncrease LiquidityDistribution
	if liquidityDistribution.empty() {
		valueIncrease = parseLiquidityDistribution(TokenA0TokenB0, pairID)
	} else {
		valueIncrease = parseLiquidityDistribution(params["PoolValueIncrease"], pairID)
	}

	return withdrawTestParams{
		DepositState: DepositState{
			ExistingShareHolders:          existingShareHolders,
			ExistingLiquidityDistribution: liquidityDistribution,
			SharedParams:                  DefaultSharedParams,
			PoolValueIncrease:             valueIncrease,
		},
		SharesToRemoveAmm: sharesToRemove,
	}
}

func hydrateAllWithdrawTestCases(paramsList []map[string]string) []withdrawTestParams {
	allTCs := make([]withdrawTestParams, 0)
	for i, paramsRaw := range paramsList {
		pairID := generatePairID(i)
		tc := hydrateWithdrawTestCase(paramsRaw, pairID)
		tc.PairID = pairID
		allTCs = append(allTCs, tc)
	}

	// De-dupe test cases hydration creates some duplicates
	//return removeDuplicateTests(allTCs)
	return allTCs
}

func TestWithdraw(t *testing.T) {
	testParams := []testParams{
		{field: "ExistingShareHolders", states: []string{Creator, OneOtherAndCreator}},
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
		{field: "PoolValueIncrease", states: []string{TokenA0TokenB0}},
		{field: "SharesToRemoveAmm", states: []string{ZeroPCT, FiftyPCT, HundredPct}},
	}
	testCasesRaw := generatePermutations(testParams)
	testCases := hydrateAllWithdrawTestCases(testCasesRaw)

	s := new(DexStateTestSuite)
	s.SetT(t)
	s.SetupTest()

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s.SetT(t)
			tc.printTestInfo(t)

			s.setupDepositState(tc.DepositState)
			s.fundCreatorBalanceDefault(tc.PairID)
			//
			poolID, found := s.App.DexKeeper.GetPoolIDByParams(s.Ctx, tc.PairID, tc.Tick, tc.Fee)
			if tc.ExistingShareHolders == None {
				// This is the ID that will be used when the pool is created
				poolID = s.App.DexKeeper.GetPoolCount(s.Ctx)
			} else {
				require.True(t, found, "Pool not found after deposit")
			}
			poolDenom := dextypes.NewPoolDenom(poolID)
			balancesBefore := s.GetBalances()
			existingSharesOwned := balancesBefore.Creator.AmountOf(poolDenom)
			toWithdraw := existingSharesOwned.MulRaw(tc.SharesToRemoveAmm).QuoRaw(100)
			//// Do the actual Withdraw
			_, err := s.makeWithdraw(
				s.creator,
				tc.ExistingLiquidityDistribution.TokenA.Denom,
				tc.ExistingLiquidityDistribution.TokenB.Denom,
				toWithdraw,
			)

			//
			// Assert new state is correct
			s.handleWithdrawFailureCases(tc, err)

			TokenABalanceBefore := balancesBefore.Creator.AmountOf(tc.ExistingLiquidityDistribution.TokenA.Denom)
			TokenBBalanceBefore := balancesBefore.Creator.AmountOf(tc.ExistingLiquidityDistribution.TokenB.Denom)

			balancesAfter := s.GetBalances()
			TokenABalanceAfter := balancesAfter.Creator.AmountOf(tc.ExistingLiquidityDistribution.TokenA.Denom)
			TokenBBalanceAfter := balancesAfter.Creator.AmountOf(tc.ExistingLiquidityDistribution.TokenB.Denom)
			// Assertion 1
			// toWithdraw = withdrawnTokenA + withdrawnTokenB*priceTakerToMaker
			priceTakerToMaker := dextypes.MustCalcPrice(-1 * tc.Tick)
			s.Require().Equal(
				toWithdraw,
				TokenABalanceAfter.Sub(TokenABalanceBefore).Add(
					priceTakerToMaker.MulInt(TokenBBalanceAfter.Sub(TokenBBalanceBefore)).TruncateInt(),
				),
			)
			newExistingSharesOwned := balancesAfter.Creator.AmountOf(poolDenom)
			// Assertion 2
			// exact amount of shares burned from a `creator` account
			s.intsApproxEqual("New shares owned", newExistingSharesOwned, existingSharesOwned.Sub(toWithdraw))

			// Assertion 3
			// exact amount of shares burned not just moved
			newExistingSharesTotal := balancesAfter.Total.AmountOf(poolDenom)
			existingSharesTotal := balancesBefore.Total.AmountOf(poolDenom)
			s.intsApproxEqual("New total shares supply", newExistingSharesTotal, existingSharesTotal.Sub(toWithdraw))

			// Assertion 4
			// Withdrawn ratio equals pool liquidity ratio (dex balance of the tokens)
			// Ac/Bc = Ap/Bp => Ac*Bp = Ap*Bc, modified the equation to avoid div operation
			balDeltaTokenA := BalancesDelta(balancesAfter, balancesBefore, tc.ExistingLiquidityDistribution.TokenA.Denom)
			balDeltaTokenB := BalancesDelta(balancesAfter, balancesBefore, tc.ExistingLiquidityDistribution.TokenB.Denom)
			s.intsApproxEqual("",
				balDeltaTokenA.Creator.Mul(balancesBefore.Dex.AmountOf(tc.ExistingLiquidityDistribution.TokenB.Denom)),
				balDeltaTokenB.Creator.Mul(balancesBefore.Dex.AmountOf(tc.ExistingLiquidityDistribution.TokenA.Denom)),
			)
		})
	}
}
