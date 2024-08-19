package dex_state_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/testutil/apptesting"
	"github.com/neutron-org/neutron/v4/testutil/common/sample"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	dexkeeper "github.com/neutron-org/neutron/v4/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
)

// Constants //////////////////////////////////////////////////////////////////

// Bools
const (
	True  = "True"
	False = "False"
)

// Percents
const (
	ZeroPCT    = "0"
	FiftyPCT   = "50"
	HundredPct = "100"
)

// ExistingShareHolders
const (
	None               = "None"
	Creator            = "Creator"
	OneOther           = "OneOther"
	OneOtherAndCreator = "OneOtherAndCreator"
)

// LiquidityDistribution
//
//nolint:gosec
const (
	TokenA0TokenB0 = "TokenA0TokenB0"
	TokenA0TokenB1 = "TokenA0TokenB1"
	TokenA0TokenB2 = "TokenA0TokenB2"
	TokenA1TokenB0 = "TokenA1TokenB0"
	TokenA1TokenB1 = "TokenA1TokenB1"
	TokenA1TokenB2 = "TokenA1TokenB2"
	TokenA2TokenB0 = "TokenA2TokenB0"
	TokenA2TokenB1 = "TokenA2TokenB1"
	TokenA2TokenB2 = "TokenA2TokenB2"
)

// Default Values
const (
	BaseTokenAmount        = 1_000_000
	DefaultTick            = 0
	DefaultFee             = 1
	DefaultStartingBalance = 10_000_000
)

var (
	BaseTokenAmountInt        = math.NewInt(BaseTokenAmount)
	DefaultStartingBalanceInt = math.NewInt(DefaultStartingBalance)
)

type SharedParams struct {
	Tick     int64
	Fee      uint64
	PairID   *dextypes.PairID
	TestName string
}

var DefaultSharedParams = SharedParams{
	Tick: DefaultTick,
	Fee:  DefaultFee,
}

// Types //////////////////////////////////////////////////////////////////////

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

// State Parsers //////////////////////////////////////////////////////////////

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

func parseLiquidityDistribution(liquidityDistribution string, pairID *dextypes.PairID) LiquidityDistribution {
	tokenA := pairID.Token0
	tokenB := pairID.Token1
	switch liquidityDistribution {
	case TokenA0TokenB0:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.ZeroInt()), TokenB: sdk.NewCoin(tokenB, math.ZeroInt())}
	case TokenA0TokenB1:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.ZeroInt()), TokenB: sdk.NewCoin(tokenB, math.NewInt(1).Mul(BaseTokenAmountInt))}
	case TokenA0TokenB2:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.ZeroInt()), TokenB: sdk.NewCoin(tokenB, math.NewInt(2).Mul(BaseTokenAmountInt))}
	case TokenA1TokenB0:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin(tokenB, math.ZeroInt())}
	case TokenA1TokenB1:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin(tokenB, math.NewInt(1).Mul(BaseTokenAmountInt))}
	case TokenA1TokenB2:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.NewInt(1).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin(tokenB, math.NewInt(2).Mul(BaseTokenAmountInt))}
	case TokenA2TokenB0:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.NewInt(2).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin(tokenB, math.ZeroInt())}
	case TokenA2TokenB1:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.NewInt(2).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin(tokenB, math.NewInt(1).Mul(BaseTokenAmountInt))}
	case TokenA2TokenB2:
		return LiquidityDistribution{TokenA: sdk.NewCoin(tokenA, math.NewInt(2).Mul(BaseTokenAmountInt)), TokenB: sdk.NewCoin(tokenB, math.NewInt(2).Mul(BaseTokenAmountInt))}
	default:
		panic("invalid liquidity distribution")
	}
}

// Misc. Helpers //////////////////////////////////////////////////////////////
func (s *DexStateTestSuite) makeDeposit(addr sdk.AccAddress, depositAmts LiquidityDistribution, disableAutoSwap bool) (*dextypes.MsgDepositResponse, error) {
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

func calcDepositValueAsToken0(tick int64, amount0, amount1 math.Int) math_utils.PrecDec {
	price1To0CenterTick := dextypes.MustCalcPrice(tick)
	amount1ValueAsToken0 := price1To0CenterTick.MulInt(amount1)
	depositValue := amount1ValueAsToken0.Add(math_utils.NewPrecDecFromInt(amount0))

	return depositValue
}

func generatePairID(i int) *dextypes.PairID {
	token0 := fmt.Sprintf("TokenA%d", i)
	token1 := fmt.Sprintf("TokenB%d", i+1)
	return dextypes.MustNewPairID(token0, token1)
}

func (s *DexStateTestSuite) fundCreatorBalanceDefault(pairID *dextypes.PairID) {
	coins := sdk.NewCoins(
		sdk.NewCoin(pairID.Token0, DefaultStartingBalanceInt),
		sdk.NewCoin(pairID.Token1, DefaultStartingBalanceInt),
	)
	s.FundAcc(s.creator, coins)
}

// Assertions /////////////////////////////////////////////////////////////////

func (s *DexStateTestSuite) intsEqual(field string, expected, actual math.Int) {
	s.True(actual.Equal(expected), "For %v: Expected %v Got %v", field, expected, actual)
}

func (s *DexStateTestSuite) assertBalance(addr sdk.AccAddress, denom string, expected math.Int) {
	trueBalance := s.App.BankKeeper.GetBalance(s.Ctx, addr, denom)
	s.intsEqual(fmt.Sprintf("Balance %s", denom), expected, trueBalance.Amount)
}

func (s *DexStateTestSuite) assertCreatorBalance(denom string, expected math.Int) {
	s.assertBalance(s.creator, denom, expected)
}

func (s *DexStateTestSuite) assertDexBalance(denom string, expected math.Int) {
	s.assertBalance(s.App.AccountKeeper.GetModuleAddress("dex"), denom, expected)
}

func (s *DexStateTestSuite) assertPoolBalance(pairID *dextypes.PairID, tick int64, fee uint64, expectedA, expectedB math.Int) {
	pool, found := s.App.DexKeeper.GetPool(s.Ctx, pairID, tick, fee)
	s.True(found, "Pool not found")

	reservesA := pool.LowerTick0.ReservesMakerDenom
	reservesB := pool.UpperTick1.ReservesMakerDenom

	s.intsEqual("Pool ReservesA", expectedA, reservesA)
	s.intsEqual("Pool ReservesB", expectedB, reservesB)
}

// Core Test Setup ////////////////////////////////////////////////////////////

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

func cloneMap(original map[string]string) map[string]string {
	// Create a new map
	cloned := make(map[string]string)
	// Copy each key-value pair from the original map to the new map
	for key, value := range original {
		cloned[key] = value
	}

	return cloned
}

type testParams struct {
	field  string
	states []string
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

func removeDuplicateTests[T depositTestParams](testCases []T) []T {
	result := make([]T, 0)
	seenTCs := make(map[string]bool)
	for _, tc := range testCases {
		tcStr := fmt.Sprintf("%v", tc)
		if _, ok := seenTCs[tcStr]; !ok {
			result = append(result, tc)
		}
		seenTCs[tcStr] = true
	}
	return result
}
