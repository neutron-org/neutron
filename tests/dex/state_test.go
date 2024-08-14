package dex_state_test

import (
	"fmt"
	"reflect"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v4/testutil/apptesting"
	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	dexkeeper "github.com/neutron-org/neutron/v4/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
)

// Shared Setup Code //////////////////////////////////////////////////////////

// Bools
const (
	True  string = "True"
	False        = "False"
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
	TokenA0TokenB1 string = "TokenA0TokenB1"
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

func parseLiquidityDistribution(liquidityDistribution string) LiquidityDistribution {
	switch liquidityDistribution {
	case TokenA0TokenB1:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(0).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(1)).Mul(BaseTokenAmountInt)}
	case TokenA0TokenB2:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(0).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(2)).Mul(BaseTokenAmountInt)}
	case TokenA1TokenB0:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(0)).Mul(BaseTokenAmountInt)}
	case TokenA1TokenB1:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(1)).Mul(BaseTokenAmountInt)}
	case TokenA1TokenB2:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(2)).Mul(BaseTokenAmountInt)}
	case TokenA2TokenB0:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(2).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(0)).Mul(BaseTokenAmountInt)}
	case TokenA2TokenB1:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(2).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(1)).Mul(BaseTokenAmountInt)}
	case TokenA2TokenB2:
		return LiquidityDistribution{TokenA: sdk.NewCoin("TokenA", math.NewInt(2).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin("TokenB", math.NewInt(2)).Mul(BaseTokenAmountInt)}
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

func generatePermutations[T any](stateType T, testStates []testParams) []T {
	result := make([]T, 0)

	var generate func(index int, current T)
	generate = func(index int, current T) {

		// Base Case
		if index == len(testStates) {
			result = append(result, current)
			return
		}

		// Iterate through all possible values and create new states
		for _, value := range testStates[index].states {
			fieldName := testStates[index].field
			temp := current
			v := reflect.ValueOf(&temp).Elem()
			field := v.FieldByName(fieldName)
			field.SetString(value)
			generate(index+1, temp)
		}

	}

	generate(0, stateType)
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

func (s *DexStateTestSuite) SetupTest() {
	s.Setup()

	s.creator = sdk.MustAccAddressFromBech32(sample.AccAddress())
	s.alice = sdk.MustAccAddressFromBech32(sample.AccAddress())

	s.msgServer = dexkeeper.NewMsgServerImpl(s.App.DexKeeper)
}

// Deposit State Test /////////////////////////////////////////////////////////
type depositTestParams struct {
	// State Conditions
	ExistingShareHolders  string
	LiquidityDistribution string
	// Message Variants
	DisableAutoswap string
	FailTxOnBEL     string
	DepositAmounts  string
}

func (s *DexStateTestSuite) setupDepositState(params depositTestParams) {
	liquidityDistr := parseLiquidityDistribution(params.LiquidityDistribution)

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
}

func CalcDepositOutput(
	existingDistr, depositDistr LiquidityDistribution,
) (resultAmountA, resultAmountB math.Int) {
	depositA := depositDistr.TokenA.Amount
	depositB := depositDistr.TokenB.Amount
	existingA := existingDistr.TokenA.Amount
	existingB := existingDistr.TokenB.Amount

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
		maxAmountA := math.LegacyNewDecFromInt(depositA).Mul(targetRatioA).TruncateInt()
		resultAmountA = math.MinInt(depositA, maxAmountA)
		targetRatioB := math.LegacyOneDec().Quo(targetRatioA)
		maxAmountB := math.LegacyNewDecFromInt(depositB).Mul(targetRatioB).TruncateInt()
		resultAmountB = math.MinInt(depositB, maxAmountB)

		return resultAmountA, resultAmountB
	default:
		panic("unhandled deposit calc case")
	}
}

func calcExpectedDepositAmounts(existingDistr, depositDistr LiquidityDistribution, disableAutoSwap bool) (tokenAAmount, tokenBAmount math.Int) {

	amountAWithoutAutoswap, amountBWithoutAutoswap := CalcDepositOutput(existingDistr, depositDistr)

	if disableAutoSwap {
		return amountAWithoutAutoswap, amountBWithoutAutoswap
	}

}

func (s *DexStateTestSuite) validateDepositResult(params depositTestParams, _ *dextypes.MsgDepositResponse, err error) {
	// Handle case where disableAutoswap == true and deposit is imbalanced
	if params.DisableAutoswap == True && params.DepositAmounts != params.LiquidityDistribution {
		s.ErrorIs(err, dextypes.ErrZeroTrueDeposit)
	}

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
		// {field: "FailTxOnBEL", states: []string{True, False}}, I don't think this needs to be tested
		{field: "DepositAmounts", states: []string{
			TokenA0TokenB1,
			TokenA0TokenB2,
			TokenA1TokenB1,
			TokenA1TokenB2,
			TokenA2TokenB2,
		}},
	}
	testCases := generatePermutations(depositTestParams{}, testParams)

	for _, tc := range testCases {
		testName := fmt.Sprintf("%v", tc)
		t.Run(testName, func(t *testing.T) {
			s := new(DexStateTestSuite)
			s.setupDepositState(tc)
			disableAutoSwap := parseBool(tc.DisableAutoswap)
			depositAmts := parseLiquidityDistribution(tc.DepositAmounts)

			pairID := dextypes.PairID{Token0: "TokenA", Token1: "TokenB"}
			poolID, found := s.App.DexKeeper.GetPoolIDByParams(s.Ctx, &pairID, DefaultTick, DefaultFee)
			s.True(found, "Pool not found after deposit")
			poolDenom := dextypes.NewPoolDenom(poolID)

			existingSharesOwned := s.App.BankKeeper.GetBalance(s.Ctx, s.creator, poolDenom)
			resp, err := s.makeDeposit(s.creator, depositAmts, disableAutoSwap)
			s.validateDepositResult(tc, resp, err)

		})

	}

}
