package dex_state_test

import (
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/neutron-org/neutron/v7/testutil/apptesting"
	"github.com/neutron-org/neutron/v7/testutil/common/sample"
	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	dexkeeper "github.com/neutron-org/neutron/v7/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v7/x/dex/types"
)

// Constants //////////////////////////////////////////////////////////////////

// Bools
const (
	True  = "True"
	False = "False"
)

// Percents
const (
	ZeroPCT       = "0"
	FiftyPCT      = "50"
	HundredPct    = "100"
	TwoHundredPct = "200"
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
	DefaultTick            = 6932 // 1.0001^6932 ~ 2.00003
	DefaultFee             = 200
	DefaultStartingBalance = 10_000_000
)

var (
	BaseTokenAmountInt        = math.NewInt(BaseTokenAmount)
	DefaultStartingBalanceInt = math.NewInt(DefaultStartingBalance)
)

type Balances struct {
	Dex     sdk.Coins
	Creator sdk.Coins
	Alice   sdk.Coins
	Total   sdk.Coins
}

type BalanceDelta struct {
	Dex     math.Int
	Creator math.Int
	Alice   math.Int
	Total   math.Int
}

func BalancesDelta(b1, b2 Balances, denom string) BalanceDelta {
	return BalanceDelta{
		Dex:     b1.Dex.AmountOf(denom).Sub(b2.Dex.AmountOf(denom)),
		Creator: b1.Creator.AmountOf(denom).Sub(b2.Creator.AmountOf(denom)),
		Alice:   b1.Alice.AmountOf(denom).Sub(b2.Alice.AmountOf(denom)),
		Total:   b1.Total.AmountOf(denom).Sub(b2.Total.AmountOf(denom)),
	}
}

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

//nolint:unused
func (l LiquidityDistribution) doubleSided() bool {
	return l.TokenA.Amount.IsPositive() && l.TokenB.Amount.IsPositive()
}

func (l LiquidityDistribution) empty() bool {
	return l.TokenA.Amount.IsZero() && l.TokenB.Amount.IsZero()
}

//nolint:unused
func (l LiquidityDistribution) singleSided() bool {
	return !l.doubleSided() && !l.empty()
}

func (l LiquidityDistribution) hasTokenA() bool {
	return l.TokenA.Amount.IsPositive()
}

func (l LiquidityDistribution) hasTokenB() bool {
	return l.TokenB.Amount.IsPositive()
}

func splitLiquidityDistribution(liquidityDistribution LiquidityDistribution, n int64) LiquidityDistribution {
	nInt := math.NewInt(n)
	amount0 := liquidityDistribution.TokenA.Amount.Quo(nInt)
	amount1 := liquidityDistribution.TokenB.Amount.Quo(nInt)

	return LiquidityDistribution{
		TokenA: sdk.NewCoin(liquidityDistribution.TokenA.Denom, amount0),
		TokenB: sdk.NewCoin(liquidityDistribution.TokenB.Denom, amount1),
	}
}

// State Parsers //////////////////////////////////////////////////////////////

func parseInt(v string) int {
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}
	return i
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
func (s *DexStateTestSuite) GetBalances() Balances {
	var snap Balances
	snap.Creator = s.App.BankKeeper.GetAllBalances(s.Ctx, s.creator)
	snap.Alice = s.App.BankKeeper.GetAllBalances(s.Ctx, s.alice)
	snap.Dex = s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress("dex"))
	resp, err := s.App.BankKeeper.TotalSupply(s.Ctx, &types.QueryTotalSupplyRequest{})
	if err != nil {
		panic(err)
	}
	snap.Total = resp.Supply
	var key []byte
	if resp.Pagination != nil {
		key = resp.Pagination.NextKey
	}
	for key != nil {
		resp, err = s.App.BankKeeper.TotalSupply(s.Ctx, &types.QueryTotalSupplyRequest{
			Pagination: &query.PageRequest{
				Key:        key,
				Offset:     0,
				Limit:      0,
				CountTotal: false,
				Reverse:    false,
			},
		})
		if err != nil {
			panic(err)
		}
		snap.Total = snap.Total.Add(resp.Supply...)
		if resp.Pagination != nil {
			key = resp.Pagination.NextKey
		}
	}

	return snap
}

func (s *DexStateTestSuite) makeDepositDefault(addr sdk.AccAddress, depositAmts LiquidityDistribution, disableAutoSwap bool) (*dextypes.MsgDepositResponse, error) {
	return s.makeDeposit(addr, depositAmts, DefaultFee, DefaultTick, disableAutoSwap)
}

func (s *DexStateTestSuite) makeDeposit(addr sdk.AccAddress, depositAmts LiquidityDistribution, fee uint64, tick int64, disableAutoSwap bool) (*dextypes.MsgDepositResponse, error) {
	return s.msgServer.Deposit(s.Ctx, &dextypes.MsgDeposit{
		Creator:         addr.String(),
		Receiver:        addr.String(),
		TokenA:          depositAmts.TokenA.Denom,
		TokenB:          depositAmts.TokenB.Denom,
		AmountsA:        []math.Int{depositAmts.TokenA.Amount},
		AmountsB:        []math.Int{depositAmts.TokenB.Amount},
		TickIndexesAToB: []int64{tick},
		Fees:            []uint64{fee},
		Options:         []*dextypes.DepositOptions{{DisableAutoswap: disableAutoSwap}},
	})
}

//nolint:unparam
func (s *DexStateTestSuite) makeDepositSuccess(addr sdk.AccAddress, depositAmts LiquidityDistribution, disableAutoSwap bool) *dextypes.MsgDepositResponse {
	resp, err := s.makeDepositDefault(addr, depositAmts, disableAutoSwap)
	s.NoError(err)

	return resp
}

func (s *DexStateTestSuite) makeWithdraw(addr sdk.AccAddress, tokenA, tokenB string, sharesToRemove math.Int) (*dextypes.MsgWithdrawalResponse, error) {
	return s.msgServer.Withdrawal(s.Ctx, &dextypes.MsgWithdrawal{
		Creator:         addr.String(),
		Receiver:        addr.String(),
		TokenA:          tokenA,
		TokenB:          tokenB,
		SharesToRemove:  []math.Int{sharesToRemove},
		TickIndexesAToB: []int64{DefaultTick},
		Fees:            []uint64{DefaultFee},
	})
}

func (s *DexStateTestSuite) makePlaceTakerLO(addr sdk.AccAddress, amountIn sdk.Coin, tokenOut, sellPrice string, orderType dextypes.LimitOrderType, maxAmountOut *math.Int) (*dextypes.MsgPlaceLimitOrderResponse, error) {
	p, err := math_utils.NewPrecDecFromStr(sellPrice)
	if err != nil {
		panic(err)
	}
	return s.msgServer.PlaceLimitOrder(s.Ctx, &dextypes.MsgPlaceLimitOrder{
		Creator:          addr.String(),
		Receiver:         addr.String(),
		TokenIn:          amountIn.Denom,
		TokenOut:         tokenOut,
		TickIndexInToOut: 0,
		AmountIn:         amountIn.Amount,
		OrderType:        orderType,
		ExpirationTime:   nil,
		MaxAmountOut:     maxAmountOut,
		LimitSellPrice:   &p,
	})
}

func (s *DexStateTestSuite) makePlaceLO(addr sdk.AccAddress, amountIn sdk.Coin, tokenOut, sellPrice string, orderType dextypes.LimitOrderType, expTime *time.Time) (*dextypes.MsgPlaceLimitOrderResponse, error) {
	p, err := math_utils.NewPrecDecFromStr(sellPrice)
	if err != nil {
		panic(err)
	}
	return s.msgServer.PlaceLimitOrder(s.Ctx, &dextypes.MsgPlaceLimitOrder{
		Creator:          addr.String(),
		Receiver:         addr.String(),
		TokenIn:          amountIn.Denom,
		TokenOut:         tokenOut,
		TickIndexInToOut: 0,
		AmountIn:         amountIn.Amount,
		OrderType:        orderType,
		ExpirationTime:   expTime,
		MaxAmountOut:     nil,
		LimitSellPrice:   &p,
	})
}

func (s *DexStateTestSuite) makePlaceLOSuccess(addr sdk.AccAddress, amountIn sdk.Coin, tokenOut, sellPrice string, orderType dextypes.LimitOrderType, expTime *time.Time) *dextypes.MsgPlaceLimitOrderResponse {
	resp, err := s.makePlaceLO(addr, amountIn, tokenOut, sellPrice, orderType, expTime)
	s.NoError(err)
	return resp
}

func (s *DexStateTestSuite) makeCancel(addr sdk.AccAddress, trancheKey string) (*dextypes.MsgCancelLimitOrderResponse, error) {
	return s.msgServer.CancelLimitOrder(s.Ctx, &dextypes.MsgCancelLimitOrder{
		Creator:    addr.String(),
		TrancheKey: trancheKey,
	})
}

func (s *DexStateTestSuite) makeWithdrawFilled(addr sdk.AccAddress, trancheKey string) (*dextypes.MsgWithdrawFilledLimitOrderResponse, error) {
	return s.msgServer.WithdrawFilledLimitOrder(s.Ctx, &dextypes.MsgWithdrawFilledLimitOrder{
		Creator:    addr.String(),
		TrancheKey: trancheKey,
	})
}

func (s *DexStateTestSuite) makeWithdrawFilledSuccess(addr sdk.AccAddress, trancheKey string) *dextypes.MsgWithdrawFilledLimitOrderResponse {
	resp, err := s.makeWithdrawFilled(addr, trancheKey)
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

func (s *DexStateTestSuite) intsApproxEqual(field string, expected, actual math.Int, absPrecision int64) {
	s.True(actual.Sub(expected).Abs().LTE(math.NewInt(absPrecision)), "For %v: Expected %v (+-%d) Got %v)", field, expected, absPrecision, actual)
}

func (s *DexStateTestSuite) assertBalance(addr sdk.AccAddress, denom string, expected math.Int) {
	trueBalance := s.App.BankKeeper.GetBalance(s.Ctx, addr, denom)
	s.intsApproxEqual(fmt.Sprintf("Balance %s", denom), expected, trueBalance.Amount, 1)
}

func (s *DexStateTestSuite) assertBalanceWithPrecision(addr sdk.AccAddress, denom string, expected math.Int, prec int64) {
	trueBalance := s.App.BankKeeper.GetBalance(s.Ctx, addr, denom)
	s.intsApproxEqual(fmt.Sprintf("Balance %s", denom), expected, trueBalance.Amount, prec)
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

	s.intsApproxEqual("Pool ReservesA", expectedA, reservesA, 1)
	s.intsApproxEqual("Pool ReservesB", expectedB, reservesB, 1)
}

// Core Test Setup ////////////////////////////////////////////////////////////

type DexStateTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer dextypes.MsgServer
	creator   sdk.AccAddress
	alice     sdk.AccAddress
	bob       sdk.AccAddress
}

func (s *DexStateTestSuite) SetupTest() {
	s.Setup()
	s.creator = sdk.MustAccAddressFromBech32(sample.AccAddress())
	s.alice = sdk.MustAccAddressFromBech32(sample.AccAddress())
	s.bob = sdk.MustAccAddressFromBech32(sample.AccAddress())

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
