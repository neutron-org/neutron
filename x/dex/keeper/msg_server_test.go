//nolint:unused,unparam // Lots of useful test helper fns that we don't want to delete, also extra params we need to keep
package keeper_test

import (
	"math"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/testutil/apptesting"
	"github.com/neutron-org/neutron/v6/testutil/common/sample"
	testkeeper "github.com/neutron-org/neutron/v6/testutil/dex/keeper"
	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	dexkeeper "github.com/neutron-org/neutron/v6/x/dex/keeper"
	testutils "github.com/neutron-org/neutron/v6/x/dex/keeper/internal/testutils"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// Test suite
type DexTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer types.MsgServer
	alice     sdk.AccAddress
	bob       sdk.AccAddress
	carol     sdk.AccAddress
	dan       sdk.AccAddress
}

var defaultPairID = &types.PairID{Token0: "TokenA", Token1: "TokenB"}

var denomMultiple = sdkmath.NewInt(1000000)

var defaultTradePairID0To1 = &types.TradePairID{
	TakerDenom: "TokenA",
	MakerDenom: "TokenB",
}

var defaultTradePairID1To0 = &types.TradePairID{
	TakerDenom: "TokenB",
	MakerDenom: "TokenA",
}

func TestDexTestSuite(t *testing.T) {
	suite.Run(t, new(DexTestSuite))
}

func (s *DexTestSuite) SetupTest() {
	s.Setup()

	s.alice = []byte("alice")
	s.bob = []byte("bob")
	s.carol = []byte("carol")
	s.dan = []byte("dan")

	s.msgServer = dexkeeper.NewMsgServerImpl(s.App.DexKeeper)
}

// NOTE: In order to simulate more realistic trade volume and avoid inadvertent failures due to ErrInvalidPositionSpread
// all of the basic user operations (fundXXXBalance, assertXXXBalance, XXXLimitsSells, etc.) treat TokenA and TokenB
// as BIG tokens with an exponent of 6. Ie. fundAliceBalance(10, 10) funds alices account with 10,000,000 small TokenA and TokenB.
// For tests requiring more accuracy methods that take Ints (ie. assertXXXAccountBalancesInt, NewWithdrawlInt) are used
// and assume that amount are being provided in terms of small tokens.

// Example:
// s.fundAliceBalances(10, 10)
// s.assertAliceBalances(10, 10) ==> True
// s.assertAliceBalancesInt(sdkmath.NewInt(10_000_000), sdkmath.NewInt(10_000_000)) ==> true

// Fund accounts

func (s *DexTestSuite) fundAccountBalances(account sdk.AccAddress, aBalance, bBalance int64) {
	aBalanceInt := sdkmath.NewInt(aBalance).Mul(denomMultiple)
	bBalanceInt := sdkmath.NewInt(bBalance).Mul(denomMultiple)
	balances := sdk.NewCoins(testutils.NewACoin(aBalanceInt), testutils.NewBCoin(bBalanceInt))

	testutils.FundAccount(s.App.BankKeeper, s.Ctx, account, balances)
	s.assertAccountBalances(account, aBalance, bBalance)
}

func (s *DexTestSuite) fundAccountBalancesWithDenom(
	addr sdk.AccAddress,
	amounts sdk.Coins,
) {
	testutils.FundAccount(s.App.BankKeeper, s.Ctx, addr, amounts)
}

func (s *DexTestSuite) fundAliceBalances(a, b int64) {
	s.fundAccountBalances(s.alice, a, b)
}

func (s *DexTestSuite) fundBobBalances(a, b int64) {
	s.fundAccountBalances(s.bob, a, b)
}

func (s *DexTestSuite) fundCarolBalances(a, b int64) {
	s.fundAccountBalances(s.carol, a, b)
}

func (s *DexTestSuite) fundDanBalances(a, b int64) {
	s.fundAccountBalances(s.dan, a, b)
}

/// Assert balances

func (s *DexTestSuite) assertAccountBalancesInt(
	account sdk.AccAddress,
	aBalance sdkmath.Int,
	bBalance sdkmath.Int,
) {
	aActual := s.App.BankKeeper.GetBalance(s.Ctx, account, "TokenA").Amount
	s.Assert().True(aBalance.Equal(aActual), "expected %s != actual %s", aBalance, aActual)

	bActual := s.App.BankKeeper.GetBalance(s.Ctx, account, "TokenB").Amount
	s.Assert().True(bBalance.Equal(bActual), "expected %s != actual %s", bBalance, bActual)
}

func (s *DexTestSuite) assertAccountBalances(
	account sdk.AccAddress,
	aBalance int64,
	bBalance int64,
) {
	s.assertAccountBalancesInt(account, sdkmath.NewInt(aBalance).Mul(denomMultiple), sdkmath.NewInt(bBalance).Mul(denomMultiple))
}

func (s *DexTestSuite) assertAccountBalanceWithDenomInt(
	account sdk.AccAddress,
	denom string,
	expBalance sdkmath.Int,
) {
	actualBalance := s.App.BankKeeper.GetBalance(s.Ctx, account, denom).Amount
	s.Assert().
		True(expBalance.Equal(actualBalance), "expected %s != actual %s", expBalance, actualBalance)
}

func (s *DexTestSuite) assertAccountBalanceWithDenom(
	account sdk.AccAddress,
	denom string,
	expBalance int64,
) {
	expBalanceInt := sdkmath.NewInt(expBalance).Mul(denomMultiple)
	s.assertAccountBalanceWithDenomInt(account, denom, expBalanceInt)
}

func (s *DexTestSuite) assertAliceBalances(a, b int64) {
	s.assertAccountBalances(s.alice, a, b)
}

func (s *DexTestSuite) assertAliceBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.alice, a, b)
}

func (s *DexTestSuite) assertBobBalances(a, b int64) {
	s.assertAccountBalances(s.bob, a, b)
}

func (s *DexTestSuite) assertBobBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.bob, a, b)
}

func (s *DexTestSuite) assertCarolBalances(a, b int64) {
	s.assertAccountBalances(s.carol, a, b)
}

func (s *DexTestSuite) assertCarolBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.carol, a, b)
}

func (s *DexTestSuite) assertDanBalances(a, b int64) {
	s.assertAccountBalances(s.dan, a, b)
}

func (s *DexTestSuite) assertDanBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.dan, a, b)
}

func (s *DexTestSuite) assertDexBalances(a, b int64) {
	s.assertAccountBalances(s.App.AccountKeeper.GetModuleAddress("dex"), a, b)
}

func (s *DexTestSuite) assertDexBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.App.AccountKeeper.GetModuleAddress("dex"), a, b)
}

func (s *DexTestSuite) assertDexBalanceWithDenom(denom string, expectedAmount int64) {
	s.assertAccountBalanceWithDenom(
		s.App.AccountKeeper.GetModuleAddress("dex"),
		denom,
		expectedAmount,
	)
}

func (s *DexTestSuite) assertDexBalanceWithDenomInt(denom string, expectedAmount sdkmath.Int) {
	s.assertAccountBalanceWithDenomInt(
		s.App.AccountKeeper.GetModuleAddress("dex"),
		denom,
		expectedAmount,
	)
}

func (s *DexTestSuite) traceBalances() {
	aliceA := s.App.BankKeeper.GetBalance(s.Ctx, s.alice, "TokenA")
	aliceB := s.App.BankKeeper.GetBalance(s.Ctx, s.alice, "TokenB")
	bobA := s.App.BankKeeper.GetBalance(s.Ctx, s.bob, "TokenA")
	bobB := s.App.BankKeeper.GetBalance(s.Ctx, s.bob, "TokenB")
	carolA := s.App.BankKeeper.GetBalance(s.Ctx, s.carol, "TokenA")
	carolB := s.App.BankKeeper.GetBalance(s.Ctx, s.carol, "TokenB")
	danA := s.App.BankKeeper.GetBalance(s.Ctx, s.dan, "TokenA")
	danB := s.App.BankKeeper.GetBalance(s.Ctx, s.dan, "TokenB")
	s.T().Logf(
		"Alice: %+v %+v\nBob: %+v %+v\nCarol: %+v %+v\nDan: %+v %+v",
		aliceA, aliceB,
		bobA, bobB,
		carolA, carolB,
		danA, danB,
	)
}

/// Place limit order

func (s *DexTestSuite) aliceLimitSells(
	selling string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	return s.limitSellsSuccess(s.alice, selling, tick, amountIn, orderTypeOpt...)
}

func (s *DexTestSuite) bobLimitSells(
	selling string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	return s.limitSellsSuccess(s.bob, selling, tick, amountIn, orderTypeOpt...)
}

func (s *DexTestSuite) carolLimitSells(
	selling string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	return s.limitSellsSuccess(s.carol, selling, tick, amountIn, orderTypeOpt...)
}

func (s *DexTestSuite) danLimitSells(
	selling string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	return s.limitSellsSuccess(s.dan, selling, tick, amountIn, orderTypeOpt...)
}

func (s *DexTestSuite) limitSellsSuccess(
	account sdk.AccAddress,
	tokenIn string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	trancheKey, err := s.limitSells(account, tokenIn, tick, amountIn, orderTypeOpt...)
	s.Assert().Nil(err)
	return trancheKey
}

func (s *DexTestSuite) aliceLimitSellsGoodTil(
	selling string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	return s.limitSellsGoodTil(s.alice, selling, tick, amountIn, goodTil)
}

func (s *DexTestSuite) bobLimitSellsGoodTil(
	selling string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	return s.limitSellsGoodTil(s.bob, selling, tick, amountIn, goodTil)
}

func (s *DexTestSuite) carolLimitSellsGoodTil(
	selling string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	return s.limitSellsGoodTil(s.carol, selling, tick, amountIn, goodTil)
}

func (s *DexTestSuite) danLimitSellsGoodTil(
	selling string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	return s.limitSellsGoodTil(s.dan, selling, tick, amountIn, goodTil)
}

func (s *DexTestSuite) assertAliceLimitSellFails(
	err error,
	selling string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	s.assertLimitSellFails(s.alice, err, selling, tickIndexNormalized, amountIn, orderTypeOpt...)
}

func (s *DexTestSuite) assertBobLimitSellFails(
	err error,
	selling string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	s.assertLimitSellFails(s.bob, err, selling, tickIndexNormalized, amountIn, orderTypeOpt...)
}

func (s *DexTestSuite) assertCarolLimitSellFails(
	err error,
	selling string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	s.assertLimitSellFails(s.carol, err, selling, tickIndexNormalized, amountIn, orderTypeOpt...)
}

func (s *DexTestSuite) assertDanLimitSellFails(
	err error,
	selling string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	s.assertLimitSellFails(s.dan, err, selling, tickIndexNormalized, amountIn, orderTypeOpt...)
}

func (s *DexTestSuite) assertLimitSellFails(
	account sdk.AccAddress,
	expectedErr error,
	tokenIn string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	_, err := s.limitSells(account, tokenIn, tickIndexNormalized, amountIn, orderTypeOpt...)
	s.Assert().ErrorIs(err, expectedErr)
}

func (s *DexTestSuite) aliceLimitSellsWithMaxOut(
	selling string,
	tick, amountIn, maxAmountOut int,
) string {
	return s.limitSellsWithMaxOut(s.alice, selling, tick, amountIn, maxAmountOut)
}

func (s *DexTestSuite) bobLimitSellsWithMaxOut(
	selling string,
	tick, amountIn, maxAmountOut int,
) string {
	return s.limitSellsWithMaxOut(s.bob, selling, tick, amountIn, maxAmountOut)
}

func (s *DexTestSuite) carolLimitSellsWithMaxOut(
	selling string,
	tick, amountIn, maxAmountOut int,
) string {
	return s.limitSellsWithMaxOut(s.carol, selling, tick, amountIn, maxAmountOut)
}

func (s *DexTestSuite) danLimitSellsWithMaxOut(
	selling string,
	tick, amountIn, maxAmountOut int,
) string {
	return s.limitSellsWithMaxOut(s.dan, selling, tick, amountIn, maxAmountOut)
}

func (s *DexTestSuite) limitSellsWithMaxOut(
	account sdk.AccAddress,
	tokenIn string,
	tick, amountIn int,
	maxAmoutOut int,
) string {
	tokenIn, tokenOut := dexkeeper.GetInOutTokens(tokenIn, "TokenA", "TokenB")
	maxAmountOutInt := sdkmath.NewInt(int64(maxAmoutOut)).Mul(denomMultiple)

	msg, err := s.msgServer.PlaceLimitOrder(s.Ctx, &types.MsgPlaceLimitOrder{
		Creator:          account.String(),
		Receiver:         account.String(),
		TokenIn:          tokenIn,
		TokenOut:         tokenOut,
		TickIndexInToOut: int64(tick),
		AmountIn:         sdkmath.NewInt(int64(amountIn)).Mul(denomMultiple),
		OrderType:        types.LimitOrderType_FILL_OR_KILL,
		MaxAmountOut:     &maxAmountOutInt,
	})

	s.Assert().NoError(err)

	return msg.TrancheKey
}

func (s *DexTestSuite) aliceLimitSellsWithMinAvgPrice(
	selling string,
	limitPrice math_utils.PrecDec,
	amountIn int,
	minAvgPrice math_utils.PrecDec,
	orderType types.LimitOrderType,
) (*types.MsgPlaceLimitOrderResponse, error) {
	return s.limitSellsWithMinAvgPrice(s.alice, selling, limitPrice, amountIn, minAvgPrice, orderType)
}

func (s *DexTestSuite) limitSellsWithMinAvgPrice(
	account sdk.AccAddress,
	tokenIn string,
	limitPrice math_utils.PrecDec,
	amountIn int,
	minAvgPrice math_utils.PrecDec,
	orderType types.LimitOrderType,
) (*types.MsgPlaceLimitOrderResponse, error) {
	tokenIn, tokenOut := dexkeeper.GetInOutTokens(tokenIn, "TokenA", "TokenB")

	return s.msgServer.PlaceLimitOrder(s.Ctx, &types.MsgPlaceLimitOrder{
		Creator:             account.String(),
		Receiver:            account.String(),
		TokenIn:             tokenIn,
		TokenOut:            tokenOut,
		TickIndexInToOut:    0,
		LimitSellPrice:      &limitPrice,
		AmountIn:            sdkmath.NewInt(int64(amountIn)).Mul(denomMultiple),
		OrderType:           orderType,
		MinAverageSellPrice: &minAvgPrice,
	})
}

func (s *DexTestSuite) limitSellsWithPrice(
	account sdk.AccAddress,
	tokenIn string,
	price math_utils.PrecDec,
	amountIn int,
) string {
	tokenIn, tokenOut := dexkeeper.GetInOutTokens(tokenIn, "TokenA", "TokenB")

	msg, err := s.msgServer.PlaceLimitOrder(s.Ctx, &types.MsgPlaceLimitOrder{
		Creator:        account.String(),
		Receiver:       account.String(),
		TokenIn:        tokenIn,
		TokenOut:       tokenOut,
		LimitSellPrice: &price,
		AmountIn:       sdkmath.NewInt(int64(amountIn)).Mul(denomMultiple),
		OrderType:      types.LimitOrderType_GOOD_TIL_CANCELLED,
	})

	s.Assert().NoError(err)

	return msg.TrancheKey
}

func (s *DexTestSuite) limitSellsInt(
	account sdk.AccAddress,
	tokenIn string,
	tickIndexNormalized int, amountIn sdkmath.Int,
	orderTypeOpt ...types.LimitOrderType,
) (string, error) {
	var orderType types.LimitOrderType
	if len(orderTypeOpt) == 0 {
		orderType = types.LimitOrderType_GOOD_TIL_CANCELLED
	} else {
		orderType = orderTypeOpt[0]
	}

	tradePairID := types.NewTradePairIDFromTaker(defaultPairID, tokenIn)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(int64(tickIndexNormalized))
	msg, err := s.msgServer.PlaceLimitOrder(s.Ctx, &types.MsgPlaceLimitOrder{
		Creator:          account.String(),
		Receiver:         account.String(),
		TokenIn:          tradePairID.TakerDenom,
		TokenOut:         tradePairID.MakerDenom,
		TickIndexInToOut: tickIndexTakerToMaker,
		AmountIn:         amountIn,
		OrderType:        orderType,
	})
	if err != nil {
		return "", err
	}

	return msg.TrancheKey, nil
}

func (s *DexTestSuite) limitSellsIntSuccess(
	account sdk.AccAddress,
	tokenIn string,
	tickIndexNormalized int,
	amountIn sdkmath.Int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	trancheKey, err := s.limitSellsInt(account, tokenIn, tickIndexNormalized, amountIn, orderTypeOpt...)
	s.NoError(err)

	return trancheKey
}

func (s *DexTestSuite) limitSells(
	account sdk.AccAddress,
	tokenIn string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) (string, error) {
	return s.limitSellsInt(account, tokenIn, tickIndexNormalized, sdkmath.NewInt(int64(amountIn)).Mul(denomMultiple), orderTypeOpt...)
}

func (s *DexTestSuite) limitSellsGoodTil(
	account sdk.AccAddress,
	tokenIn string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	tradePairID := types.NewTradePairIDFromTaker(defaultPairID, tokenIn)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(int64(tick))

	msg, err := s.msgServer.PlaceLimitOrder(s.Ctx, &types.MsgPlaceLimitOrder{
		Creator:          account.String(),
		Receiver:         account.String(),
		TokenIn:          tradePairID.TakerDenom,
		TokenOut:         tradePairID.MakerDenom,
		TickIndexInToOut: tickIndexTakerToMaker,
		AmountIn:         sdkmath.NewInt(int64(amountIn)).Mul(denomMultiple),
		OrderType:        types.LimitOrderType_GOOD_TIL_TIME,
		ExpirationTime:   &goodTil,
	})

	s.Assert().NoError(err)

	return msg.TrancheKey
}

// / Deposit
type Deposit struct {
	AmountA   sdkmath.Int
	AmountB   sdkmath.Int
	TickIndex int64
	Fee       uint64
	Options   *types.DepositOptions
}

func NewDepositInt(amountA, amountB sdkmath.Int, tickIndex, fee int) *Deposit {
	return &Deposit{
		AmountA:   amountA,
		AmountB:   amountB,
		TickIndex: int64(tickIndex),
		Fee:       uint64(fee), //nolint:gosec
	}
}

func NewDeposit(amountA, amountB, tickIndex, fee int) *Deposit {
	return NewDepositInt(sdkmath.NewInt(int64(amountA)).Mul(denomMultiple), sdkmath.NewInt(int64(amountB)).Mul(denomMultiple), tickIndex, fee)
}

func NewDepositWithOptions(
	amountA, amountB, tickIndex, fee int,
	options types.DepositOptions,
) *Deposit {
	return &Deposit{
		AmountA:   sdkmath.NewInt(int64(amountA)).Mul(denomMultiple),
		AmountB:   sdkmath.NewInt(int64(amountB)).Mul(denomMultiple),
		TickIndex: int64(tickIndex),
		Fee:       uint64(fee), //nolint:gosec
		Options:   &options,
	}
}

func (s *DexTestSuite) aliceDeposits(deposits ...*Deposit) *types.MsgDepositResponse {
	return s.depositsSuccess(s.alice, deposits)
}

func (s *DexTestSuite) bobDeposits(deposits ...*Deposit) *types.MsgDepositResponse {
	return s.depositsSuccess(s.bob, deposits)
}

func (s *DexTestSuite) carolDeposits(deposits ...*Deposit) *types.MsgDepositResponse {
	return s.depositsSuccess(s.carol, deposits)
}

func (s *DexTestSuite) danDeposits(deposits ...*Deposit) *types.MsgDepositResponse {
	return s.depositsSuccess(s.dan, deposits)
}

func (s *DexTestSuite) depositsSuccess(
	account sdk.AccAddress,
	deposits []*Deposit,
	pairID ...types.PairID,
) *types.MsgDepositResponse {
	resp, err := s.deposits(account, deposits, pairID...)
	s.Assert().Nil(err)
	return resp
}

func (s *DexTestSuite) deposits(
	account sdk.AccAddress,
	deposits []*Deposit,
	pairID ...types.PairID,
) (*types.MsgDepositResponse, error) {
	amountsA := make([]sdkmath.Int, len(deposits))
	amountsB := make([]sdkmath.Int, len(deposits))
	tickIndexes := make([]int64, len(deposits))
	fees := make([]uint64, len(deposits))
	options := make([]*types.DepositOptions, len(deposits))
	for i, e := range deposits {
		amountsA[i] = e.AmountA
		amountsB[i] = e.AmountB
		tickIndexes[i] = e.TickIndex
		fees[i] = e.Fee
		if e.Options != nil {
			options[i] = e.Options
		} else {
			options[i] = &types.DepositOptions{}
		}

	}

	var tokenA, tokenB string
	switch {
	case len(pairID) == 0:
		tokenA = "TokenA"
		tokenB = "TokenB"
	case len(pairID) == 1:
		tokenA = pairID[0].Token0
		tokenB = pairID[0].Token1
	case len(pairID) > 1:
		s.Assert().Fail("Only 1 pairID can be provided")
	}

	msg := &types.MsgDeposit{
		Creator:         account.String(),
		Receiver:        account.String(),
		TokenA:          tokenA,
		TokenB:          tokenB,
		AmountsA:        amountsA,
		AmountsB:        amountsB,
		TickIndexesAToB: tickIndexes,
		Fees:            fees,
		Options:         options,
	}
	err := msg.Validate()
	if err != nil {
		return &types.MsgDepositResponse{}, err
	}
	return s.msgServer.Deposit(s.Ctx, msg)
}

func (s *DexTestSuite) getLiquidityAtTick(tickIndex int64, fee uint64) (sdkmath.Int, sdkmath.Int) {
	pool, err := s.App.DexKeeper.GetOrInitPool(s.Ctx, defaultPairID, tickIndex, fee)
	s.Assert().NoError(err)

	liquidityA := pool.LowerTick0.ReservesMakerDenom
	liquidityB := pool.UpperTick1.ReservesMakerDenom

	return liquidityA, liquidityB
}

func (s *DexTestSuite) getLiquidityAtTickWithDenom(
	pairID *types.PairID,
	tickIndex int64,
	fee uint64,
) (sdkmath.Int, sdkmath.Int) {
	pool, err := s.App.DexKeeper.GetOrInitPool(s.Ctx, pairID, tickIndex, fee)
	s.Assert().NoError(err)

	liquidityA := pool.LowerTick0.ReservesMakerDenom
	liquidityB := pool.UpperTick1.ReservesMakerDenom

	return liquidityA, liquidityB
}

func (s *DexTestSuite) assertAliceDepositFails(err error, deposits ...*Deposit) {
	s.assertDepositFails(s.alice, err, deposits...)
}

func (s *DexTestSuite) assertBobDepositFails(err error, deposits ...*Deposit) {
	s.assertDepositFails(s.bob, err, deposits...)
}

func (s *DexTestSuite) assertCarolDepositFails(err error, deposits ...*Deposit) {
	s.assertDepositFails(s.carol, err, deposits...)
}

func (s *DexTestSuite) assertDanDepositFails(err error, deposits ...*Deposit) {
	s.assertDepositFails(s.dan, err, deposits...)
}

func (s *DexTestSuite) assertDepositFails(
	account sdk.AccAddress,
	expectedErr error,
	deposits ...*Deposit,
) {
	_, err := s.deposits(account, deposits)
	s.Assert().ErrorIs(err, expectedErr)
}

func (s *DexTestSuite) assertDepositReponse(
	depositResponse, expectedDepositResponse DepositReponse,
) {
	for i := range expectedDepositResponse.amountsA {
		s.Assert().Equal(
			depositResponse.amountsA[i],
			expectedDepositResponse.amountsA[i],
			"Assertion failed for response.amountsA[%d]", i,
		)
		s.Assert().Equal(
			depositResponse.amountsB[i],
			expectedDepositResponse.amountsB[i],
			"Assertion failed for response.amountsB[%d]", i,
		)
	}
}

type DepositReponse struct {
	amountsA []sdkmath.Int
	amountsB []sdkmath.Int
}

// Withdraw
type Withdrawal struct {
	TickIndex int64
	Fee       uint64
	Shares    sdkmath.Int
}

func NewWithdrawalInt(shares sdkmath.Int, tick int64, fee uint64) *Withdrawal {
	return &Withdrawal{
		Shares:    shares,
		Fee:       fee,
		TickIndex: tick,
	}
}

// Multiples amount of shares to represent BIGtoken with exponent 6
func NewWithdrawal(shares, tick int64, fee uint64) *Withdrawal {
	return NewWithdrawalInt(sdkmath.NewInt(shares).Mul(denomMultiple), tick, fee)
}

func (s *DexTestSuite) aliceWithdraws(withdrawals ...*Withdrawal) {
	s.withdraws(s.alice, withdrawals...)
}

func (s *DexTestSuite) bobWithdraws(withdrawals ...*Withdrawal) {
	s.withdraws(s.bob, withdrawals...)
}

func (s *DexTestSuite) carolWithdraws(withdrawals ...*Withdrawal) {
	s.withdraws(s.carol, withdrawals...)
}

func (s *DexTestSuite) danWithdraws(withdrawals ...*Withdrawal) {
	s.withdraws(s.dan, withdrawals...)
}

func (s *DexTestSuite) withdraws(account sdk.AccAddress, withdrawals ...*Withdrawal) {
	tickIndexes := make([]int64, len(withdrawals))
	fee := make([]uint64, len(withdrawals))
	sharesToRemove := make([]sdkmath.Int, len(withdrawals))
	for i, e := range withdrawals {
		tickIndexes[i] = e.TickIndex
		fee[i] = e.Fee
		sharesToRemove[i] = e.Shares
	}

	_, err := s.msgServer.Withdrawal(s.Ctx, &types.MsgWithdrawal{
		Creator:         account.String(),
		Receiver:        account.String(),
		TokenA:          "TokenA",
		TokenB:          "TokenB",
		SharesToRemove:  sharesToRemove,
		TickIndexesAToB: tickIndexes,
		Fees:            fee,
	})
	s.Assert().Nil(err)
}

func (s *DexTestSuite) aliceWithdrawFails(expectedErr error, withdrawals ...*Withdrawal) {
	s.withdrawFails(s.alice, expectedErr, withdrawals...)
}

func (s *DexTestSuite) bobWithdrawFails(expectedErr error, withdrawals ...*Withdrawal) {
	s.withdrawFails(s.bob, expectedErr, withdrawals...)
}

func (s *DexTestSuite) carolWithdrawFails(expectedErr error, withdrawals ...*Withdrawal) {
	s.withdrawFails(s.carol, expectedErr, withdrawals...)
}

func (s *DexTestSuite) danWithdrawFails(expectedErr error, withdrawals ...*Withdrawal) {
	s.withdrawFails(s.dan, expectedErr, withdrawals...)
}

func (s *DexTestSuite) withdrawFails(
	account sdk.AccAddress,
	expectedErr error,
	withdrawals ...*Withdrawal,
) {
	tickIndexes := make([]int64, len(withdrawals))
	fee := make([]uint64, len(withdrawals))
	sharesToRemove := make([]sdkmath.Int, len(withdrawals))
	for i, e := range withdrawals {
		tickIndexes[i] = e.TickIndex
		fee[i] = e.Fee
		sharesToRemove[i] = e.Shares
	}

	_, err := s.msgServer.Withdrawal(s.Ctx, &types.MsgWithdrawal{
		Creator:         account.String(),
		Receiver:        account.String(),
		TokenA:          "TokenA",
		TokenB:          "TokenB",
		SharesToRemove:  sharesToRemove,
		TickIndexesAToB: tickIndexes,
		Fees:            fee,
	})
	s.Assert().NotNil(err)
	s.Assert().ErrorIs(err, expectedErr)
}

/// Cancel limit order

func (s *DexTestSuite) aliceCancelsLimitSell(trancheKey string) {
	s.cancelsLimitSell(s.alice, trancheKey)
}

func (s *DexTestSuite) bobCancelsLimitSell(trancheKey string) {
	s.cancelsLimitSell(s.bob, trancheKey)
}

func (s *DexTestSuite) carolCancelsLimitSell(trancheKey string) {
	s.cancelsLimitSell(s.carol, trancheKey)
}

func (s *DexTestSuite) danCancelsLimitSell(trancheKey string) {
	s.cancelsLimitSell(s.dan, trancheKey)
}

func (s *DexTestSuite) cancelsLimitSell(account sdk.AccAddress, trancheKey string) {
	_, err := s.msgServer.CancelLimitOrder(s.Ctx, &types.MsgCancelLimitOrder{
		Creator:    account.String(),
		TrancheKey: trancheKey,
	})
	s.Assert().Nil(err)
}

func (s *DexTestSuite) aliceCancelsLimitSellFails(trancheKey string, expectedErr error) {
	s.cancelsLimitSellFails(s.alice, trancheKey, expectedErr)
}

func (s *DexTestSuite) bobCancelsLimitSellFails(trancheKey string, expectedErr error) {
	s.cancelsLimitSellFails(s.bob, trancheKey, expectedErr)
}

func (s *DexTestSuite) carolCancelsLimitSellFails(trancheKey string, expectedErr error) {
	s.cancelsLimitSellFails(s.carol, trancheKey, expectedErr)
}

func (s *DexTestSuite) danCancelsLimitSellFails(trancheKey string, expectedErr error) {
	s.cancelsLimitSellFails(s.dan, trancheKey, expectedErr)
}

func (s *DexTestSuite) cancelsLimitSellFails(
	account sdk.AccAddress,
	trancheKey string,
	expectedErr error,
) {
	_, err := s.msgServer.CancelLimitOrder(s.Ctx, &types.MsgCancelLimitOrder{
		Creator:    account.String(),
		TrancheKey: trancheKey,
	})
	s.Assert().ErrorIs(err, expectedErr)
}

/// MultiHopSwap

func (s *DexTestSuite) aliceMultiHopSwaps(
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwaps(s.alice, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *DexTestSuite) bobMultiHopSwaps(
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwaps(s.bob, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *DexTestSuite) carolMultiHopSwaps(
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwaps(s.carol, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *DexTestSuite) danMultiHopSwaps(
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwaps(s.dan, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *DexTestSuite) multiHopSwaps(
	account sdk.AccAddress,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	msg := types.NewMsgMultiHopSwap(
		account.String(),
		account.String(),
		routes,
		sdkmath.NewInt(int64(amountIn)).Mul(denomMultiple),
		exitLimitPrice,
		pickBest,
	)
	_, err := s.msgServer.MultiHopSwap(s.Ctx, msg)
	s.Assert().Nil(err)
}

func (s *DexTestSuite) aliceEstimatesMultiHopSwap(
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) (coinOut sdk.Coin) {
	multiHopRoutes := make([]*types.MultiHopRoute, len(routes))
	for i, hops := range routes {
		multiHopRoutes[i] = &types.MultiHopRoute{Hops: hops}
	}
	msg := &types.QueryEstimateMultiHopSwapRequest{
		Creator:        s.alice.String(),
		Receiver:       s.alice.String(),
		Routes:         multiHopRoutes,
		AmountIn:       sdkmath.NewInt(int64(amountIn)).Mul(denomMultiple),
		ExitLimitPrice: exitLimitPrice,
		PickBestRoute:  pickBest,
	}
	res, err := s.App.DexKeeper.EstimateMultiHopSwap(s.Ctx, msg)
	s.Require().Nil(err)
	return res.CoinOut
}

func (s *DexTestSuite) aliceEstimatesMultiHopSwapFails(
	expectedErr error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	multiHopRoutes := make([]*types.MultiHopRoute, len(routes))
	for i, hops := range routes {
		multiHopRoutes[i] = &types.MultiHopRoute{Hops: hops}
	}
	msg := &types.QueryEstimateMultiHopSwapRequest{
		Creator:        s.alice.String(),
		Receiver:       s.alice.String(),
		Routes:         multiHopRoutes,
		AmountIn:       sdkmath.NewInt(int64(amountIn)).Mul(denomMultiple),
		ExitLimitPrice: exitLimitPrice,
		PickBestRoute:  pickBest,
	}
	_, err := s.App.DexKeeper.EstimateMultiHopSwap(s.Ctx, msg)
	s.Assert().ErrorIs(err, expectedErr)
}

func (s *DexTestSuite) aliceMultiHopSwapFails(
	err error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwapFails(s.alice, err, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *DexTestSuite) bobMultiHopSwapFails(
	err error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwapFails(s.bob, err, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *DexTestSuite) carolMultiHopSwapFails(
	err error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwapFails(s.carol, err, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *DexTestSuite) danMultiHopSwapFails(
	err error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwapFails(s.dan, err, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *DexTestSuite) multiHopSwapFails(
	account sdk.AccAddress,
	expectedErr error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	msg := types.NewMsgMultiHopSwap(
		account.String(),
		account.String(),
		routes,
		sdkmath.NewInt(int64(amountIn)).Mul(denomMultiple),
		exitLimitPrice,
		pickBest,
	)
	_, err := s.msgServer.MultiHopSwap(s.Ctx, msg)
	s.Assert().ErrorIs(err, expectedErr)
}

/// Withdraw filled limit order

func (s *DexTestSuite) aliceWithdrawsLimitSell(trancheKey string) {
	s.withdrawsLimitSell(s.alice, trancheKey)
}

func (s *DexTestSuite) bobWithdrawsLimitSell(trancheKey string) {
	s.withdrawsLimitSell(s.bob, trancheKey)
}

func (s *DexTestSuite) carolWithdrawsLimitSell(trancheKey string) {
	s.withdrawsLimitSell(s.carol, trancheKey)
}

func (s *DexTestSuite) danWithdrawsLimitSell(trancheKey string) {
	s.withdrawsLimitSell(s.dan, trancheKey)
}

func (s *DexTestSuite) withdrawsLimitSell(account sdk.AccAddress, trancheKey string) {
	_, err := s.msgServer.WithdrawFilledLimitOrder(s.Ctx, &types.MsgWithdrawFilledLimitOrder{
		Creator:    account.String(),
		TrancheKey: trancheKey,
	})
	s.Assert().Nil(err)
}

func (s *DexTestSuite) aliceWithdrawLimitSellFails(expectedErr error, trancheKey string) {
	s.withdrawLimitSellFails(s.alice, expectedErr, trancheKey)
}

func (s *DexTestSuite) bobWithdrawLimitSellFails(expectedErr error, trancheKey string) {
	s.withdrawLimitSellFails(s.bob, expectedErr, trancheKey)
}

func (s *DexTestSuite) carolWithdrawLimitSellFails(expectedErr error, trancheKey string) {
	s.withdrawLimitSellFails(s.carol, expectedErr, trancheKey)
}

func (s *DexTestSuite) danWithdrawLimitSellFails(expectedErr error, trancheKey string) {
	s.withdrawLimitSellFails(s.dan, expectedErr, trancheKey)
}

func (s *DexTestSuite) withdrawLimitSellFails(
	account sdk.AccAddress,
	expectedErr error,
	trancheKey string,
) {
	_, err := s.msgServer.WithdrawFilledLimitOrder(s.Ctx, &types.MsgWithdrawFilledLimitOrder{
		Creator:    account.String(),
		TrancheKey: trancheKey,
	})
	s.Assert().ErrorIs(err, expectedErr)
}

// Shares
func (s *DexTestSuite) getPoolShares(
	token0 string,
	token1 string,
	tick int64,
	fee uint64,
) (shares sdkmath.Int) {
	poolID, found := s.App.DexKeeper.GetPoolIDByParams(s.Ctx, &types.PairID{Token0: token0, Token1: token1}, tick, fee)
	if !found {
		return sdkmath.ZeroInt()
	}
	poolDenom := types.NewPoolDenom(poolID)
	return s.App.BankKeeper.GetSupply(s.Ctx, poolDenom).Amount
}

func (s *DexTestSuite) assertPoolShares(
	tick int64,
	fee uint64,
	sharesExpected uint64,
) {
	sharesExpectedInt := sdkmath.NewIntFromUint64(sharesExpected).Mul(denomMultiple)
	sharesOwned := s.getPoolShares("TokenA", "TokenB", tick, fee)
	s.Assert().Equal(sharesExpectedInt, sharesOwned)
}

func (s *DexTestSuite) getAccountShares(
	account sdk.AccAddress,
	token0 string,
	token1 string,
	tick int64,
	fee uint64,
) (shares sdkmath.Int) {
	id, found := s.App.DexKeeper.GetPoolIDByParams(s.Ctx, types.MustNewPairID(token0, token1), tick, fee)
	if !found {
		return sdkmath.ZeroInt()
	}

	poolDenom := types.NewPoolDenom(id)
	return s.App.BankKeeper.GetBalance(s.Ctx, account, poolDenom).Amount
}

func (s *DexTestSuite) assertAccountSharesInt(
	account sdk.AccAddress,
	tick int64,
	fee uint64,
	sharesExpected sdkmath.Int,
) {
	sharesOwned := s.getAccountShares(account, "TokenA", "TokenB", tick, fee)
	s.Assert().
		Equal(sharesExpected, sharesOwned, "expected %s != actual %s", sharesExpected, sharesOwned)
}

func (s *DexTestSuite) assertAccountShares(
	account sdk.AccAddress,
	tick int64,
	fee uint64,
	sharesExpected uint64,
) {
	sharesExpectedInt := sdkmath.NewIntFromUint64(sharesExpected).Mul(denomMultiple)
	s.assertAccountSharesInt(account, tick, fee, sharesExpectedInt)
}

func (s *DexTestSuite) assertAliceShares(tick int64, fee, sharesExpected uint64) {
	s.assertAccountShares(s.alice, tick, fee, sharesExpected)
}

func (s *DexTestSuite) assertBobShares(tick int64, fee, sharesExpected uint64) {
	s.assertAccountShares(s.bob, tick, fee, sharesExpected)
}

func (s *DexTestSuite) assertCarolShares(tick int64, fee, sharesExpected uint64) {
	s.assertAccountShares(s.carol, tick, fee, sharesExpected)
}

func (s *DexTestSuite) assertDanShares(tick int64, fee, sharesExpected uint64) {
	s.assertAccountShares(s.dan, tick, fee, sharesExpected)
}

// Ticks
func (s *DexTestSuite) assertCurrentTicks(
	expected1To0 int64,
	expected0To1 int64,
) {
	s.assertCurr0To1(expected0To1)
	s.assertCurr1To0(expected1To0)
}

func (s *DexTestSuite) assertCurr0To1(curr0To1Expected int64) {
	curr0To1Actual, found := s.App.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(
		s.Ctx,
		defaultTradePairID0To1,
	)
	if curr0To1Expected == math.MaxInt64 {
		s.Assert().False(found)
	} else {
		s.Assert().Equal(curr0To1Expected, curr0To1Actual)
	}
}

func (s *DexTestSuite) assertCurr1To0(curr1To0Expected int64) {
	curr1to0Actual, found := s.App.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(
		s.Ctx,
		defaultTradePairID1To0,
	)
	if curr1To0Expected == math.MinInt64 {
		s.Assert().False(found)
	} else {
		s.Assert().Equal(curr1To0Expected, curr1to0Actual)
	}
}

// Pool liquidity (i.e. deposited rather than LO)
func (s *DexTestSuite) assertLiquidityAtTickInt(
	amountA, amountB sdkmath.Int,
	tickIndex int64,
	fee uint64,
) {
	liquidityA, liquidityB := s.getLiquidityAtTick(tickIndex, fee)
	s.Assert().
		True(amountA.Equal(liquidityA), "liquidity A: actual %s, expected %s", liquidityA, amountA)
	s.Assert().
		True(amountB.Equal(liquidityB), "liquidity B: actual %s, expected %s", liquidityB, amountB)
}

func (s *DexTestSuite) assertLiquidityAtTick(
	amountA, amountB int64,
	tickIndex int64,
	fee uint64,
) {
	amountAInt := sdkmath.NewInt(amountA).Mul(denomMultiple)
	amountBInt := sdkmath.NewInt(amountB).Mul(denomMultiple)
	s.assertLiquidityAtTickInt(amountAInt, amountBInt, tickIndex, fee)
}

func (s *DexTestSuite) assertLiquidityAtTickWithDenomInt(
	pairID *types.PairID,
	expected0, expected1 sdkmath.Int,
	tickIndex int64,
	fee uint64,
) {
	liquidity0, liquidity1 := s.getLiquidityAtTickWithDenom(pairID, tickIndex, fee)
	s.Assert().
		True(expected0.Equal(liquidity0), "liquidity 0: actual %s, expected %s", liquidity0, expected0)
	s.Assert().
		True(expected1.Equal(liquidity1), "liquidity 1: actual %s, expected %s", liquidity1, expected1)
}

func (s *DexTestSuite) assertLiquidityAtTickWithDenom(
	pairID *types.PairID,
	expected0,
	expected1,
	tickIndex int64,
	fee uint64,
) {
	expected0Int := sdkmath.NewInt(expected0).Mul(denomMultiple)
	expected1Int := sdkmath.NewInt(expected1).Mul(denomMultiple)
	s.assertLiquidityAtTickWithDenomInt(pairID, expected0Int, expected1Int, tickIndex, fee)
}

func (s *DexTestSuite) assertPoolLiquidity(
	amountA, amountB int64,
	tickIndex int64,
	fee uint64,
) {
	s.assertLiquidityAtTick(amountA, amountB, tickIndex, fee)
}

func (s *DexTestSuite) assertNoLiquidityAtTick(tickIndex int64, fee uint64) {
	s.assertLiquidityAtTick(0, 0, tickIndex, fee)
}

// Filled limit liquidity
func (s *DexTestSuite) assertAliceLimitFilledAtTickAtIndex(
	selling string,
	amount int,
	tickIndex int64,
	trancheKey string,
) {
	s.assertLimitFilledAtTickAtIndex(s.alice, selling, amount, tickIndex, trancheKey)
}

func (s *DexTestSuite) assertBobLimitFilledAtTickAtIndex(
	selling string,
	amount int,
	tickIndex int64,
	trancheKey string,
) {
	s.assertLimitFilledAtTickAtIndex(s.bob, selling, amount, tickIndex, trancheKey)
}

func (s *DexTestSuite) assertCarolLimitFilledAtTickAtIndex(
	selling string,
	amount int,
	tickIndex int64,
	trancheKey string,
) {
	s.assertLimitFilledAtTickAtIndex(s.carol, selling, amount, tickIndex, trancheKey)
}

func (s *DexTestSuite) assertDanLimitFilledAtTickAtIndex(
	selling string,
	amount int,
	tickIndex int64,
	trancheKey string,
) {
	s.assertLimitFilledAtTickAtIndex(s.dan, selling, amount, tickIndex, trancheKey)
}

func (s *DexTestSuite) assertLimitFilledAtTickAtIndex(
	account sdk.AccAddress,
	selling string,
	amount int,
	tickIndex int64,
	trancheKey string,
) {
	userShares, totalShares := s.getLimitUserSharesAtTick(
		account,
		selling,
		tickIndex,
	), s.getLimitTotalSharesAtTick(
		selling,
		tickIndex,
	)
	userRatio := math_utils.NewPrecDecFromInt(userShares).QuoInt(totalShares)
	filled := s.getLimitFilledLiquidityAtTickAtIndex(selling, tickIndex, trancheKey)
	amt := sdkmath.NewInt(int64(amount)).Mul(denomMultiple)
	userFilled := userRatio.MulInt(filled).RoundInt()
	s.Assert().True(amt.Equal(userFilled))
}

// Limit liquidity
func (s *DexTestSuite) assertAliceLimitLiquidityAtTick(
	selling string,
	amount int,
	tickIndex int64,
) {
	s.assertAccountLimitLiquidityAtTick(s.alice, selling, amount, tickIndex)
}

func (s *DexTestSuite) assertBobLimitLiquidityAtTick(
	selling string,
	amount int,
	tickIndex int64,
) {
	s.assertAccountLimitLiquidityAtTick(s.bob, selling, amount, tickIndex)
}

func (s *DexTestSuite) assertCarolLimitLiquidityAtTick(
	selling string,
	amount int,
	tickIndex int64,
) {
	s.assertAccountLimitLiquidityAtTick(s.carol, selling, amount, tickIndex)
}

func (s *DexTestSuite) assertDanLimitLiquidityAtTick(
	selling string,
	amount int,
	tickIndex int64,
) {
	s.assertAccountLimitLiquidityAtTick(s.dan, selling, amount, tickIndex)
}

func (s *DexTestSuite) assertAccountLimitLiquidityAtTick(
	account sdk.AccAddress,
	selling string,
	amount int,
	tickIndexNormalized int64,
) {
	userShares := s.getLimitUserSharesAtTick(account, selling, tickIndexNormalized)
	totalShares := s.getLimitTotalSharesAtTick(selling, tickIndexNormalized)
	userRatio := math_utils.NewPrecDecFromInt(userShares).QuoInt(totalShares)
	userLiquidity := userRatio.MulInt64(int64(amount)).TruncateInt()

	s.assertLimitLiquidityAtTick(selling, tickIndexNormalized, userLiquidity.Int64())
}

func (s *DexTestSuite) assertLimitLiquidityAtTick(
	selling string,
	tickIndexNormalized, amount int64,
) {
	s.assertLimitLiquidityAtTickInt(selling, tickIndexNormalized, sdkmath.NewInt(amount).Mul(denomMultiple))
}

func (s *DexTestSuite) assertLimitLiquidityAtTickInt(
	selling string,
	tickIndexNormalized int64,
	amount sdkmath.Int,
) {
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(tickIndexNormalized)
	tranches := s.App.DexKeeper.GetAllLimitOrderTrancheAtIndex(
		s.Ctx,
		tradePairID,
		tickIndexTakerToMaker,
	)
	liquidity := sdkmath.ZeroInt()
	for _, t := range tranches {
		if !t.IsExpired(s.Ctx) {
			liquidity = liquidity.Add(t.ReservesMakerDenom)
		}
	}

	s.Assert().
		True(amount.Equal(liquidity), "Incorrect liquidity: expected %s, have %s", amount.String(), liquidity.String())
}

func (s *DexTestSuite) assertFillAndPlaceTrancheKeys(
	selling string,
	tickIndexNormalized int64,
	expectedFill, expectedPlace string,
) {
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(tickIndexNormalized)
	placeTranche := s.App.DexKeeper.GetGTCPlaceTranche(s.Ctx, tradePairID, tickIndexTakerToMaker)
	fillTranche, foundFill := s.App.DexKeeper.GetFillTranche(
		s.Ctx,
		tradePairID,
		tickIndexTakerToMaker,
	)
	placeKey, fillKey := "", ""
	if placeTranche != nil {
		placeKey = placeTranche.Key.TrancheKey
	}

	if foundFill {
		fillKey = fillTranche.Key.TrancheKey
	}
	s.Assert().Equal(expectedFill, fillKey)
	s.Assert().Equal(expectedPlace, placeKey)
}

// Limit order map helpers
func (s *DexTestSuite) getLimitUserSharesAtTick(
	account sdk.AccAddress,
	selling string,
	tickIndexNormalized int64,
) sdkmath.Int {
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(tickIndexNormalized)
	tranches := s.App.DexKeeper.GetAllLimitOrderTrancheAtIndex(
		s.Ctx,
		tradePairID,
		tickIndexTakerToMaker,
	)
	fillTranche := tranches[0]
	// get user shares and total shares
	userShares := s.getLimitUserSharesAtTickAtIndex(account, fillTranche.Key.TrancheKey)
	if len(tranches) >= 2 {
		userShares = userShares.Add(
			s.getLimitUserSharesAtTickAtIndex(account, tranches[1].Key.TrancheKey),
		)
	}

	return userShares
}

func (s *DexTestSuite) getLimitUserSharesAtTickAtIndex(
	account sdk.AccAddress,
	trancheKey string,
) sdkmath.Int {
	userShares, found := s.App.DexKeeper.GetLimitOrderTrancheUser(
		s.Ctx,
		account.String(),
		trancheKey,
	)
	s.Assert().True(found, "Failed to get limit order user shares for index %s", trancheKey)
	return userShares.SharesOwned
}

func (s *DexTestSuite) getLimitTotalSharesAtTick(
	selling string,
	tickIndexNormalized int64,
) sdkmath.Int {
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(tickIndexNormalized)
	tranches := s.App.DexKeeper.GetAllLimitOrderTrancheAtIndex(
		s.Ctx,
		tradePairID,
		tickIndexTakerToMaker,
	)
	// get user shares and total shares
	totalShares := sdkmath.ZeroInt()
	for _, t := range tranches {
		totalShares = totalShares.Add(t.TotalMakerDenom)
	}

	return totalShares
}

func (s *DexTestSuite) getLimitFilledLiquidityAtTickAtIndex(
	selling string,
	tickIndex int64,
	trancheKey string,
) sdkmath.Int {
	// grab fill tranche reserves and shares
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tranche, _, found := s.App.DexKeeper.FindLimitOrderTranche(s.Ctx, &types.LimitOrderTrancheKey{
		TradePairId:           tradePairID,
		TickIndexTakerToMaker: tickIndex,
		TrancheKey:            trancheKey,
	})
	s.Assert().True(found, "Failed to get limit order filled reserves for index %s", trancheKey)

	return tranche.ReservesTakerDenom
}

func (s *DexTestSuite) getLimitReservesAtTickAtKey(
	selling string,
	tickIndex int64,
	trancheKey string,
) sdkmath.Int {
	// grab fill tranche reserves and shares
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tranche, _, found := s.App.DexKeeper.FindLimitOrderTranche(s.Ctx, &types.LimitOrderTrancheKey{
		TradePairId:           tradePairID,
		TickIndexTakerToMaker: tickIndex,
		TrancheKey:            trancheKey,
	})
	s.Assert().True(found, "Failed to get limit order reserves for index %s", trancheKey)

	return tranche.ReservesMakerDenom
}

func (s *DexTestSuite) assertNLimitOrderExpiration(expected int) {
	exps := s.App.DexKeeper.GetAllLimitOrderExpiration(s.Ctx)
	s.Assert().Equal(expected, len(exps))
}

func (s *DexTestSuite) nextBlockWithTime(blockTime time.Time) {
	newCtx := s.Ctx.WithBlockTime(blockTime)
	s.Ctx = newCtx
	_, err := s.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s.App.LastBlockHeight() + 1,
		Time:   blockTime,
	})
	require.NoError(s.T(), err)
	_, err = s.App.Commit()
	require.NoError(s.T(), err)
}

func (s *DexTestSuite) beginBlockWithTime(blockTime time.Time) {
	s.Ctx = s.Ctx.WithBlockTime(blockTime)
	_, err := s.App.BeginBlocker(s.Ctx)
	s.NoError(err)
}

func TestMsgDepositValidate(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	msgServer := dexkeeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgDeposit
		expectedErr error
	}{
		{
			"invalid creator",
			types.MsgDeposit{
				Creator:         "invalid_address",
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrInvalidAddress,
		},
		{
			"invalid receiver",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        "invalid address",
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrInvalidAddress,
		},
		{
			"invalid denom A",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "er",
				TokenB:          "factory/neutron1rxel5kdhu089fdk4xugmryx0y2wzjx8rqsa6hu/validDenom2",
				Fees:            []uint64{1},
				TickIndexesAToB: []int64{1},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrInvalidDenom,
		},
		{
			"invalid denom B",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "er",
				Fees:            []uint64{1},
				TickIndexesAToB: []int64{1},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrInvalidDenom,
		},
		{
			"denoms match",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenA",
				Fees:            []uint64{1},
				TickIndexesAToB: []int64{1},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrInvalidDenom,
		},
		{
			"invalid fee indexes length",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "factory/neutron1rxel5kdhu089fdk4xugmryx0y2wzjx8rqsa6hu/validDenom2",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{},
				AmountsA:        []sdkmath.Int{},
				AmountsB:        []sdkmath.Int{},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrUnbalancedTxArray,
		},
		{
			"invalid tick indexes length",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdkmath.Int{},
				AmountsB:        []sdkmath.Int{},
				Options:         []*types.DepositOptions{{DisableAutoswap: true}},
			},
			types.ErrUnbalancedTxArray,
		},
		{
			"invalid amounts A length",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{},
				Options:         []*types.DepositOptions{{DisableAutoswap: true}},
			},
			types.ErrUnbalancedTxArray,
		},
		{
			"invalid amounts B length",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				AmountsA:        []sdkmath.Int{},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: true}},
			},
			types.ErrUnbalancedTxArray,
		},
		{
			"invalid options length",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{1},
				TickIndexesAToB: []int64{1},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{},
			},
			types.ErrUnbalancedTxArray,
		},
		{
			"invalid no deposit",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				AmountsA:        []sdkmath.Int{},
				AmountsB:        []sdkmath.Int{},
				Options:         []*types.DepositOptions{},
			},
			types.ErrZeroDeposit,
		},
		{
			"invalid duplicate deposit",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{1, 2, 1},
				TickIndexesAToB: []int64{0, 0, 0},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt(), sdkmath.OneInt(), sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt(), sdkmath.OneInt(), sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}, {DisableAutoswap: false}, {DisableAutoswap: false}},
			},
			types.ErrDuplicatePoolDeposit,
		},
		{
			"invalid no deposit",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdkmath.Int{sdkmath.ZeroInt()},
				AmountsB:        []sdkmath.Int{sdkmath.ZeroInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrZeroDeposit,
		},
		{
			"invalid tick + fee upper",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{3},
				TickIndexesAToB: []int64{559678},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrTickOutsideRange,
		},
		{
			"invalid tick + fee lower",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{50},
				TickIndexesAToB: []int64{-559631},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrTickOutsideRange,
		},
		{
			"invalid fee overflow",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{559681},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false}},
			},
			types.ErrInvalidFee,
		},
		{
			"SwapOnDeposit without autoswap",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{1},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: true, SwapOnDeposit: true}},
			},
			types.ErrSwapOnDepositWithoutAutoswap,
		},
		{
			"invalid slop tolerance",
			types.MsgDeposit{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{1},
				TickIndexesAToB: []int64{0},
				AmountsA:        []sdkmath.Int{sdkmath.OneInt()},
				AmountsB:        []sdkmath.Int{sdkmath.OneInt()},
				Options:         []*types.DepositOptions{{DisableAutoswap: false, SwapOnDeposit: true, SwapOnDepositSlopToleranceBps: 10001}},
			},
			types.ErrInvalidSlopTolerance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.Deposit(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgWithdrawalValidate(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	msgServer := dexkeeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgWithdrawal
		expectedErr error
	}{
		{
			"invalid creator",
			types.MsgWithdrawal{
				Creator:         "invalid_address",
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []sdkmath.Int{sdkmath.OneInt()},
			},
			types.ErrInvalidAddress,
		},
		{
			"invalid receiver",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        "invalid_address",
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []sdkmath.Int{sdkmath.OneInt()},
			},
			types.ErrInvalidAddress,
		},
		{
			"invalid TokenA",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "er",
				TokenB:          "TokenB",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []sdkmath.Int{sdkmath.OneInt()},
			},
			types.ErrInvalidDenom,
		},
		{
			"invalid TokenB",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "er",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []sdkmath.Int{sdkmath.OneInt()},
			},
			types.ErrInvalidDenom,
		},
		{
			"invalid fee indexes length",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []sdkmath.Int{sdkmath.OneInt()},
			},
			types.ErrUnbalancedTxArray,
		},
		{
			"invalid tick indexes length",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{},
				SharesToRemove:  []sdkmath.Int{sdkmath.OneInt()},
			},
			types.ErrUnbalancedTxArray,
		},
		{
			"invalid shares to remove length",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []sdkmath.Int{},
			},
			types.ErrUnbalancedTxArray,
		},
		{
			"no withdraw specs",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{},
				TickIndexesAToB: []int64{},
				SharesToRemove:  []sdkmath.Int{},
			},
			types.ErrZeroWithdraw,
		},
		{
			"no withdraw specs",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{0},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []sdkmath.Int{sdkmath.ZeroInt()},
			},
			types.ErrZeroWithdraw,
		},
		{
			"invalid tick + fee upper",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{3},
				TickIndexesAToB: []int64{559678},
				SharesToRemove:  []sdkmath.Int{sdkmath.OneInt()},
			},
			types.ErrTickOutsideRange,
		},
		{
			"invalid tick + fee lower",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{50},
				TickIndexesAToB: []int64{-559631},
				SharesToRemove:  []sdkmath.Int{sdkmath.OneInt()},
			},
			types.ErrTickOutsideRange,
		},
		{
			"invalid fee overflow",
			types.MsgWithdrawal{
				Creator:         sample.AccAddress(),
				Receiver:        sample.AccAddress(),
				TokenA:          "TokenA",
				TokenB:          "TokenB",
				Fees:            []uint64{559681},
				TickIndexesAToB: []int64{0},
				SharesToRemove:  []sdkmath.Int{sdkmath.OneInt()},
			},
			types.ErrInvalidFee,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.Withdrawal(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgPlaceLimitOrderValidate(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	msgServer := dexkeeper.NewMsgServerImpl(*k)

	ZEROINT := sdkmath.ZeroInt()
	ONEINT := sdkmath.OneInt()
	ZERODEC := math_utils.ZeroPrecDec()
	TINYDEC := math_utils.MustNewPrecDecFromStr("0.000000000000000000000000494")
	HUGEDEC := math_utils.MustNewPrecDecFromStr("2020125331305056766452345.127500016657360222036663652")
	FIVEDEC := math_utils.NewPrecDec(5)
	tests := []struct {
		name        string
		msg         types.MsgPlaceLimitOrder
		expectedErr error
	}{
		{
			"invalid creator",
			types.MsgPlaceLimitOrder{
				Creator:          "invalid_address",
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         sdkmath.OneInt(),
			},
			types.ErrInvalidAddress,
		},
		{
			"invalid receiver",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         "invalid_address",
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         sdkmath.OneInt(),
			},
			types.ErrInvalidAddress,
		},
		{
			"invalid TokenIn",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "er",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         sdkmath.OneInt(),
			},
			types.ErrInvalidDenom,
		},
		{
			"invalid TokenOut",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "er",
				TickIndexInToOut: 0,
				AmountIn:         sdkmath.OneInt(),
			},
			types.ErrInvalidDenom,
		},
		{
			"denoms match",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenA",
				TickIndexInToOut: 0,
				AmountIn:         sdkmath.OneInt(),
			},
			types.ErrInvalidDenom,
		},
		{
			"invalid zero limit order",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         sdkmath.ZeroInt(),
			},
			types.ErrZeroLimitOrder,
		},
		{
			"zero maxOut",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         sdkmath.OneInt(),
				MaxAmountOut:     &ZEROINT,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			types.ErrZeroMaxAmountOut,
		},
		{
			"max out with maker order",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 0,
				AmountIn:         sdkmath.OneInt(),
				MaxAmountOut:     &ONEINT,
				OrderType:        types.LimitOrderType_GOOD_TIL_CANCELLED,
			},
			types.ErrInvalidMaxAmountOutForMaker,
		},
		{
			"tick outside range upper",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: 700_000,
				AmountIn:         sdkmath.OneInt(),
				OrderType:        types.LimitOrderType_GOOD_TIL_CANCELLED,
			},
			types.ErrTickOutsideRange,
		},
		{
			"tick outside range lower",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				TickIndexInToOut: -600_000,
				AmountIn:         sdkmath.OneInt(),
				OrderType:        types.LimitOrderType_GOOD_TIL_CANCELLED,
			},
			types.ErrTickOutsideRange,
		},
		{
			"price < minPrice",
			types.MsgPlaceLimitOrder{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				TokenIn:        "TokenA",
				TokenOut:       "TokenB",
				LimitSellPrice: &TINYDEC,
				AmountIn:       sdkmath.OneInt(),
			},
			types.ErrPriceOutsideRange,
		},
		{
			"price > maxPrice",
			types.MsgPlaceLimitOrder{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				TokenIn:        "TokenA",
				TokenOut:       "TokenB",
				LimitSellPrice: &HUGEDEC,
				AmountIn:       sdkmath.OneInt(),
			},
			types.ErrPriceOutsideRange,
		},
		{
			"invalid tickIndexInToOut & LimitSellPrice",
			types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "TokenA",
				TokenOut:         "TokenB",
				LimitSellPrice:   &FIVEDEC,
				TickIndexInToOut: 6,
				AmountIn:         sdkmath.OneInt(),
			},
			types.ErrInvalidPriceAndTick,
		},
		{
			"invalid zero min average sell price",
			types.MsgPlaceLimitOrder{
				Creator:             sample.AccAddress(),
				Receiver:            sample.AccAddress(),
				TokenIn:             "TokenA",
				TokenOut:            "TokenB",
				LimitSellPrice:      &FIVEDEC,
				AmountIn:            sdkmath.OneInt(),
				MinAverageSellPrice: &ZERODEC,
			},
			types.ErrZeroMinAverageSellPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.PlaceLimitOrder(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgWithdrawFilledLimitOrderValidate(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	msgServer := dexkeeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgWithdrawFilledLimitOrder
		expectedErr error
	}{
		{
			"invalid creator",
			types.MsgWithdrawFilledLimitOrder{
				Creator:    "invalid_address",
				TrancheKey: "ORDER123",
			},
			types.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.WithdrawFilledLimitOrder(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgCancelLimitOrderValidate(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	msgServer := dexkeeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgCancelLimitOrder
		expectedErr error
	}{
		{
			"invalid creator",
			types.MsgCancelLimitOrder{
				Creator:    "invalid_address",
				TrancheKey: "ORDER123",
			},
			types.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.CancelLimitOrder(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgMultiHopSwapValidate(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	msgServer := dexkeeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgMultiHopSwap
		expectedErr error
	}{
		{
			"invalid creator address",
			types.MsgMultiHopSwap{
				Creator:  "invalid_address",
				Receiver: sample.AccAddress(),
				Routes: []*types.MultiHopRoute{
					{Hops: []string{"TokenA", "TokenB", "TokenC"}},
				},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			types.ErrInvalidAddress,
		},
		{
			"invalid receiver address",
			types.MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: "invalid_address",
				Routes: []*types.MultiHopRoute{
					{Hops: []string{"TokenA", "TokenB", "TokenC"}},
				},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			types.ErrInvalidAddress,
		},
		{
			"missing route",
			types.MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*types.MultiHopRoute{},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			types.ErrMissingMultihopRoute,
		},
		{
			"invalid exit tokens",
			types.MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Routes: []*types.MultiHopRoute{
					{Hops: []string{"TokenA", "TokenB", "TokenC"}},
					{Hops: []string{"TokenA", "TokenB", "TokenZ"}},
				},
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			types.ErrMultihopExitTokensMismatch,
		},
		{
			"invalid amountIn",
			types.MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*types.MultiHopRoute{{Hops: []string{"TokenA", "TokenB", "TokenC"}}},
				AmountIn:       sdkmath.NewInt(-1),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			types.ErrZeroSwap,
		},
		{
			"cycles in hops",
			types.MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Routes: []*types.MultiHopRoute{
					{Hops: []string{"TokenA", "TokenB", "TokenC"}},                               // normal
					{Hops: []string{"TokenA", "TokenB", "TokenD", "TokenE", "TokenB", "TokenC"}}, // has cycle
				},
				AmountIn:       sdkmath.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			types.ErrCycleInHops,
		},
		{
			"invalid denom in route",
			types.MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Routes: []*types.MultiHopRoute{
					{Hops: []string{"TokenA", "TokenB", "TokenC"}},
					{Hops: []string{"TokenA", "TokenB", "TokenD", "TokenE", "er", "TokenC"}},
				},
				AmountIn:       sdkmath.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			types.ErrInvalidDenom,
		},
		{
			"entry token denom mismatch in route",
			types.MsgMultiHopSwap{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Routes: []*types.MultiHopRoute{
					{Hops: []string{"TokenA", "TokenB", "TokenC"}},
					{Hops: []string{"TokenD", "TokenB", "TokenC"}},
				},
				AmountIn:       sdkmath.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0.9"),
			},
			types.ErrMultihopEntryTokensMismatch,
		},
		{
			"zero exit limit price",
			types.MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*types.MultiHopRoute{{Hops: []string{"TokenA", "TokenB", "TokenC"}}},
				AmountIn:       sdkmath.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("0"),
			},
			types.ErrZeroExitPrice,
		},
		{
			"negative exit limit price",
			types.MsgMultiHopSwap{
				Creator:        sample.AccAddress(),
				Receiver:       sample.AccAddress(),
				Routes:         []*types.MultiHopRoute{{Hops: []string{"TokenA", "TokenB", "TokenC"}}},
				AmountIn:       sdkmath.OneInt(),
				ExitLimitPrice: math_utils.MustNewPrecDecFromStr("-0.5"),
			},
			types.ErrZeroExitPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.MultiHopSwap(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgUpdateParamsValidate(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	msgServer := dexkeeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgUpdateParams
		expectedErr string
	}{
		{
			"empty authority",
			types.MsgUpdateParams{
				Authority: "",
			},
			"authority is invalid",
		},
		{
			"invalid authority",
			types.MsgUpdateParams{
				Authority: "invalid authority",
			},
			"authority is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.UpdateParams(ctx, &tt.msg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}
