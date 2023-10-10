package keeper_test

import (
	"context"
	"math"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	dualityapp "github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	math_utils "github.com/neutron-org/neutron/utils/math"
	. "github.com/neutron-org/neutron/x/dex/keeper"
	. "github.com/neutron-org/neutron/x/dex/keeper/internal/testutils"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/suite"
)

// / Test suite
type MsgServerTestSuite struct {
	suite.Suite
	app       *dualityapp.App
	msgServer types.MsgServer
	ctx       sdk.Context
	alice     sdk.AccAddress
	bob       sdk.AccAddress
	carol     sdk.AccAddress
	dan       sdk.AccAddress
	goCtx     context.Context
}

var defaultPairID *types.PairID = &types.PairID{Token0: "TokenA", Token1: "TokenB"}

var defaultTradePairID0To1 *types.TradePairID = &types.TradePairID{
	TakerDenom: "TokenA",
	MakerDenom: "TokenB",
}

var defaultTradePairID1To0 *types.TradePairID = &types.TradePairID{
	TakerDenom: "TokenB",
	MakerDenom: "TokenA",
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) SetupTest() {
	app := testutil.Setup(s.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockGasMeter(sdk.NewInfiniteGasMeter())

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	accAlice := app.AccountKeeper.NewAccountWithAddress(ctx, s.alice)
	app.AccountKeeper.SetAccount(ctx, accAlice)
	accBob := app.AccountKeeper.NewAccountWithAddress(ctx, s.bob)
	app.AccountKeeper.SetAccount(ctx, accBob)
	accCarol := app.AccountKeeper.NewAccountWithAddress(ctx, s.carol)
	app.AccountKeeper.SetAccount(ctx, accCarol)
	accDan := app.AccountKeeper.NewAccountWithAddress(ctx, s.dan)
	app.AccountKeeper.SetAccount(ctx, accDan)

	s.app = app
	s.msgServer = NewMsgServerImpl(app.DexKeeper)
	s.ctx = ctx
	s.goCtx = sdk.WrapSDKContext(ctx)
	s.alice = sdk.AccAddress([]byte("alice"))
	s.bob = sdk.AccAddress([]byte("bob"))
	s.carol = sdk.AccAddress([]byte("carol"))
	s.dan = sdk.AccAddress([]byte("dan"))
}

/// Fund accounts

func (s *MsgServerTestSuite) fundAccountBalances(account sdk.AccAddress, aBalance, bBalance int64) {
	aBalanceInt := sdkmath.NewInt(aBalance)
	bBalanceInt := sdkmath.NewInt(bBalance)
	balances := sdk.NewCoins(NewACoin(aBalanceInt), NewBCoin(bBalanceInt))
	err := FundAccount(s.app.BankKeeper, s.ctx, account, balances)
	s.Assert().NoError(err)
	s.assertAccountBalances(account, aBalance, bBalance)
}

func (s *MsgServerTestSuite) fundAccountBalancesWithDenom(
	addr sdk.AccAddress,
	amounts sdk.Coins,
) error {
	if err := s.app.BankKeeper.MintCoins(s.ctx, types.ModuleName, amounts); err != nil {
		return err
	}

	return s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, addr, amounts)
}

func (s *MsgServerTestSuite) fundAliceBalances(a, b int64) {
	s.fundAccountBalances(s.alice, a, b)
}

func (s *MsgServerTestSuite) fundBobBalances(a, b int64) {
	s.fundAccountBalances(s.bob, a, b)
}

func (s *MsgServerTestSuite) fundCarolBalances(a, b int64) {
	s.fundAccountBalances(s.carol, a, b)
}

func (s *MsgServerTestSuite) fundDanBalances(a, b int64) {
	s.fundAccountBalances(s.dan, a, b)
}

/// Assert balances

func (s *MsgServerTestSuite) assertAccountBalancesInt(
	account sdk.AccAddress,
	aBalance sdkmath.Int,
	bBalance sdkmath.Int,
) {
	aActual := s.app.BankKeeper.GetBalance(s.ctx, account, "TokenA").Amount
	s.Assert().True(aBalance.Equal(aActual), "expected %s != actual %s", aBalance, aActual)

	bActual := s.app.BankKeeper.GetBalance(s.ctx, account, "TokenB").Amount
	s.Assert().True(bBalance.Equal(bActual), "expected %s != actual %s", bBalance, bActual)
}

func (s *MsgServerTestSuite) assertAccountBalances(
	account sdk.AccAddress,
	aBalance int64,
	bBalance int64,
) {
	s.assertAccountBalancesInt(account, sdkmath.NewInt(aBalance), sdkmath.NewInt(bBalance))
}

func (s *MsgServerTestSuite) assertAccountBalanceWithDenom(
	account sdk.AccAddress,
	denom string,
	expBalance int64,
) {
	actualBalance := s.app.BankKeeper.GetBalance(s.ctx, account, denom).Amount
	expBalanceInt := sdkmath.NewInt(expBalance)
	s.Assert().
		True(expBalanceInt.Equal(actualBalance), "expected %s != actual %s", expBalance, actualBalance)
}

func (s *MsgServerTestSuite) assertAliceBalances(a, b int64) {
	s.assertAccountBalances(s.alice, a, b)
}

func (s *MsgServerTestSuite) assertAliceBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.alice, a, b)
}

func (s *MsgServerTestSuite) assertBobBalances(a, b int64) {
	s.assertAccountBalances(s.bob, a, b)
}

func (s *MsgServerTestSuite) assertBobBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.bob, a, b)
}

func (s *MsgServerTestSuite) assertCarolBalances(a, b int64) {
	s.assertAccountBalances(s.carol, a, b)
}

func (s *MsgServerTestSuite) assertCarolBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.carol, a, b)
}

func (s *MsgServerTestSuite) assertDanBalances(a, b int64) {
	s.assertAccountBalances(s.dan, a, b)
}

func (s *MsgServerTestSuite) assertDanBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.dan, a, b)
}

func (s *MsgServerTestSuite) assertDexBalances(a, b int64) {
	s.assertAccountBalances(s.app.AccountKeeper.GetModuleAddress("dex"), a, b)
}

func (s *MsgServerTestSuite) assertDexBalanceWithDenom(denom string, expectedAmount int64) {
	s.assertAccountBalanceWithDenom(
		s.app.AccountKeeper.GetModuleAddress("dex"),
		denom,
		expectedAmount,
	)
}

func (s *MsgServerTestSuite) assertDexBalancesInt(a, b sdkmath.Int) {
	s.assertAccountBalancesInt(s.app.AccountKeeper.GetModuleAddress("dex"), a, b)
}

func (s *MsgServerTestSuite) traceBalances() {
	aliceA := s.app.BankKeeper.GetBalance(s.ctx, s.alice, "TokenA")
	aliceB := s.app.BankKeeper.GetBalance(s.ctx, s.alice, "TokenB")
	bobA := s.app.BankKeeper.GetBalance(s.ctx, s.bob, "TokenA")
	bobB := s.app.BankKeeper.GetBalance(s.ctx, s.bob, "TokenB")
	carolA := s.app.BankKeeper.GetBalance(s.ctx, s.carol, "TokenA")
	carolB := s.app.BankKeeper.GetBalance(s.ctx, s.carol, "TokenB")
	danA := s.app.BankKeeper.GetBalance(s.ctx, s.dan, "TokenA")
	danB := s.app.BankKeeper.GetBalance(s.ctx, s.dan, "TokenB")
	s.T().Logf(
		"Alice: %+v %+v\nBob: %+v %+v\nCarol: %+v %+v\nDan: %+v %+v",
		aliceA, aliceB,
		bobA, bobB,
		carolA, carolB,
		danA, danB,
	)
}

/// Place limit order

func (s *MsgServerTestSuite) aliceLimitSells(
	selling string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	return s.limitSellsSuccess(s.alice, selling, tick, amountIn, orderTypeOpt...)
}

func (s *MsgServerTestSuite) bobLimitSells(
	selling string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	return s.limitSellsSuccess(s.bob, selling, tick, amountIn, orderTypeOpt...)
}

func (s *MsgServerTestSuite) carolLimitSells(
	selling string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	return s.limitSellsSuccess(s.carol, selling, tick, amountIn, orderTypeOpt...)
}

func (s *MsgServerTestSuite) danLimitSells(
	selling string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	return s.limitSellsSuccess(s.dan, selling, tick, amountIn, orderTypeOpt...)
}

func (s *MsgServerTestSuite) limitSellsSuccess(
	account sdk.AccAddress,
	tokenIn string,
	tick, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) string {
	trancheKey, err := s.limitSells(account, tokenIn, tick, amountIn, orderTypeOpt...)
	s.Assert().Nil(err)
	return trancheKey
}

func (s *MsgServerTestSuite) aliceLimitSellsGoodTil(
	selling string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	return s.limitSellsGoodTil(s.alice, selling, tick, amountIn, goodTil)
}

func (s *MsgServerTestSuite) bobLimitSellsGoodTil(
	selling string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	return s.limitSellsGoodTil(s.bob, selling, tick, amountIn, goodTil)
}

func (s *MsgServerTestSuite) carolLimitSellsGoodTil(
	selling string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	return s.limitSellsGoodTil(s.carol, selling, tick, amountIn, goodTil)
}

func (s *MsgServerTestSuite) danLimitSellsGoodTil(
	selling string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	return s.limitSellsGoodTil(s.dan, selling, tick, amountIn, goodTil)
}

func (s *MsgServerTestSuite) assertAliceLimitSellFails(
	err error,
	selling string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	s.assertLimitSellFails(s.alice, err, selling, tickIndexNormalized, amountIn, orderTypeOpt...)
}

func (s *MsgServerTestSuite) assertBobLimitSellFails(
	err error,
	selling string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	s.assertLimitSellFails(s.bob, err, selling, tickIndexNormalized, amountIn, orderTypeOpt...)
}

func (s *MsgServerTestSuite) assertCarolLimitSellFails(
	err error,
	selling string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	s.assertLimitSellFails(s.carol, err, selling, tickIndexNormalized, amountIn, orderTypeOpt...)
}

func (s *MsgServerTestSuite) assertDanLimitSellFails(
	err error,
	selling string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	s.assertLimitSellFails(s.dan, err, selling, tickIndexNormalized, amountIn, orderTypeOpt...)
}

func (s *MsgServerTestSuite) assertLimitSellFails(
	account sdk.AccAddress,
	expectedErr error,
	tokenIn string,
	tickIndexNormalized, amountIn int,
	orderTypeOpt ...types.LimitOrderType,
) {
	_, err := s.limitSells(account, tokenIn, tickIndexNormalized, amountIn, orderTypeOpt...)
	s.Assert().ErrorIs(err, expectedErr)
}

func (s *MsgServerTestSuite) aliceLimitSellsWithMaxOut(
	selling string,
	tick, amountIn, maxAmountOut int,
) string {
	return s.limitSellsWithMaxOut(s.alice, selling, tick, amountIn, maxAmountOut)
}

func (s *MsgServerTestSuite) bobLimitSellsWithMaxOut(
	selling string,
	tick, amountIn, maxAmountOut int,
) string {
	return s.limitSellsWithMaxOut(s.bob, selling, tick, amountIn, maxAmountOut)
}

func (s *MsgServerTestSuite) carolLimitSellsWithMaxOut(
	selling string,
	tick, amountIn, maxAmountOut int,
) string {
	return s.limitSellsWithMaxOut(s.carol, selling, tick, amountIn, maxAmountOut)
}

func (s *MsgServerTestSuite) danLimitSellsWithMaxOut(
	selling string,
	tick, amountIn, maxAmountOut int,
) string {
	return s.limitSellsWithMaxOut(s.dan, selling, tick, amountIn, maxAmountOut)
}

func (s *MsgServerTestSuite) limitSellsWithMaxOut(
	account sdk.AccAddress,
	tokenIn string,
	tick, amountIn int,
	maxAmoutOut int,
) string {
	tokenIn, tokenOut := GetInOutTokens(tokenIn, "TokenA", "TokenB")
	maxAmountOutInt := sdkmath.NewInt(int64(maxAmoutOut))

	msg, err := s.msgServer.PlaceLimitOrder(s.goCtx, &types.MsgPlaceLimitOrder{
		Creator:          account.String(),
		Receiver:         account.String(),
		TokenIn:          tokenIn,
		TokenOut:         tokenOut,
		TickIndexInToOut: int64(tick),
		AmountIn:         sdkmath.NewInt(int64(amountIn)),
		OrderType:        types.LimitOrderType_FILL_OR_KILL,
		MaxAmountOut:     &maxAmountOutInt,
	})

	s.Assert().NoError(err)

	return msg.TrancheKey
}

func (s *MsgServerTestSuite) limitSells(
	account sdk.AccAddress,
	tokenIn string,
	tickIndexNormalized, amountIn int,
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
	msg, err := s.msgServer.PlaceLimitOrder(s.goCtx, &types.MsgPlaceLimitOrder{
		Creator:          account.String(),
		Receiver:         account.String(),
		TokenIn:          tradePairID.TakerDenom,
		TokenOut:         tradePairID.MakerDenom,
		TickIndexInToOut: tickIndexTakerToMaker,
		AmountIn:         sdkmath.NewInt(int64(amountIn)),
		OrderType:        orderType,
	})

	return msg.TrancheKey, err
}

func (s *MsgServerTestSuite) limitSellsGoodTil(
	account sdk.AccAddress,
	tokenIn string,
	tick, amountIn int,
	goodTil time.Time,
) string {
	tradePairID := types.NewTradePairIDFromTaker(defaultPairID, tokenIn)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(int64(tick))

	msg, err := s.msgServer.PlaceLimitOrder(s.goCtx, &types.MsgPlaceLimitOrder{
		Creator:          account.String(),
		Receiver:         account.String(),
		TokenIn:          tradePairID.TakerDenom,
		TokenOut:         tradePairID.MakerDenom,
		TickIndexInToOut: tickIndexTakerToMaker,
		AmountIn:         sdkmath.NewInt(int64(amountIn)),
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
}

type DepositOptions struct {
	DisableAutoswap bool
}

type DepositWithOptions struct {
	AmountA   sdkmath.Int
	AmountB   sdkmath.Int
	TickIndex int64
	Fee       uint64
	Options   DepositOptions
}

func NewDeposit(amountA, amountB, tickIndex, fee int) *Deposit {
	return &Deposit{
		AmountA:   sdkmath.NewInt(int64(amountA)),
		AmountB:   sdkmath.NewInt(int64(amountB)),
		TickIndex: int64(tickIndex),
		Fee:       uint64(fee),
	}
}

func NewDepositWithOptions(
	amountA, amountB, tickIndex, fee int,
	options DepositOptions,
) *DepositWithOptions {
	return &DepositWithOptions{
		AmountA:   sdkmath.NewInt(int64(amountA)),
		AmountB:   sdkmath.NewInt(int64(amountB)),
		TickIndex: int64(tickIndex),
		Fee:       uint64(fee),
		Options:   options,
	}
}

func (s *MsgServerTestSuite) aliceDeposits(deposits ...*Deposit) {
	s.deposits(s.alice, deposits)
}

func (s *MsgServerTestSuite) aliceDepositsWithOptions(deposits ...*DepositWithOptions) {
	s.depositsWithOptions(s.alice, deposits...)
}

func (s *MsgServerTestSuite) bobDeposits(deposits ...*Deposit) {
	s.deposits(s.bob, deposits)
}

func (s *MsgServerTestSuite) bobDepositsWithOptions(deposits ...*DepositWithOptions) {
	s.depositsWithOptions(s.bob, deposits...)
}

func (s *MsgServerTestSuite) carolDeposits(deposits ...*Deposit) {
	s.deposits(s.carol, deposits)
}

func (s *MsgServerTestSuite) danDeposits(deposits ...*Deposit) {
	s.deposits(s.dan, deposits)
}

func (s *MsgServerTestSuite) deposits(
	account sdk.AccAddress,
	deposits []*Deposit,
	pairID ...types.PairID,
) {
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
		options[i] = &types.DepositOptions{DisableAutoswap: false}
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

	_, err := s.msgServer.Deposit(s.goCtx, &types.MsgDeposit{
		Creator:         account.String(),
		Receiver:        account.String(),
		TokenA:          tokenA,
		TokenB:          tokenB,
		AmountsA:        amountsA,
		AmountsB:        amountsB,
		TickIndexesAToB: tickIndexes,
		Fees:            fees,
		Options:         options,
	})
	s.Assert().Nil(err)
}

func (s *MsgServerTestSuite) depositsWithOptions(
	account sdk.AccAddress,
	deposits ...*DepositWithOptions,
) {
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
		options[i] = &types.DepositOptions{
			DisableAutoswap: e.Options.DisableAutoswap,
		}
	}

	_, err := s.msgServer.Deposit(s.goCtx, &types.MsgDeposit{
		Creator:         account.String(),
		Receiver:        account.String(),
		TokenA:          "TokenA",
		TokenB:          "TokenB",
		AmountsA:        amountsA,
		AmountsB:        amountsB,
		TickIndexesAToB: tickIndexes,
		Fees:            fees,
		Options:         options,
	})
	s.Assert().Nil(err)
}

func (s *MsgServerTestSuite) getLiquidityAtTick(tickIndex int64, fee uint64) (sdkmath.Int, sdkmath.Int) {
	pool, err := s.app.DexKeeper.GetOrInitPool(s.ctx, defaultPairID, tickIndex, fee)
	s.Assert().NoError(err)

	liquidityA := pool.LowerTick0.ReservesMakerDenom
	liquidityB := pool.UpperTick1.ReservesMakerDenom

	return liquidityA, liquidityB
}

func (s *MsgServerTestSuite) getLiquidityAtTickWithDenom(
	pairID *types.PairID,
	tickIndex int64,
	fee uint64,
) (sdkmath.Int, sdkmath.Int) {
	pool, err := s.app.DexKeeper.GetOrInitPool(s.ctx, pairID, tickIndex, fee)
	s.Assert().NoError(err)

	liquidityA := pool.LowerTick0.ReservesMakerDenom
	liquidityB := pool.UpperTick1.ReservesMakerDenom

	return liquidityA, liquidityB
}

func (s *MsgServerTestSuite) assertAliceDepositFails(err error, deposits ...*Deposit) {
	s.assertDepositFails(s.alice, err, deposits...)
}

func (s *MsgServerTestSuite) assertBobDepositFails(err error, deposits ...*Deposit) {
	s.assertDepositFails(s.bob, err, deposits...)
}

func (s *MsgServerTestSuite) assertCarolDepositFails(err error, deposits ...*Deposit) {
	s.assertDepositFails(s.carol, err, deposits...)
}

func (s *MsgServerTestSuite) assertDanDepositFails(err error, deposits ...*Deposit) {
	s.assertDepositFails(s.dan, err, deposits...)
}

func (s *MsgServerTestSuite) assertDepositFails(
	account sdk.AccAddress,
	expectedErr error,
	deposits ...*Deposit,
) {
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
		options[i] = &types.DepositOptions{DisableAutoswap: true}
	}

	_, err := s.msgServer.Deposit(s.goCtx, &types.MsgDeposit{
		Creator:         account.String(),
		Receiver:        account.String(),
		TokenA:          "TokenA",
		TokenB:          "TokenB",
		AmountsA:        amountsA,
		AmountsB:        amountsB,
		TickIndexesAToB: tickIndexes,
		Fees:            fees,
		Options:         options,
	})
	s.Assert().NotNil(err)
	s.Assert().ErrorIs(err, expectedErr)
}

func (s *MsgServerTestSuite) assertDepositReponse(
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

func NewWithdrawal(shares, tick int64, fee uint64) *Withdrawal {
	return NewWithdrawalInt(sdkmath.NewInt(shares), tick, fee)
}

func (s *MsgServerTestSuite) aliceWithdraws(withdrawals ...*Withdrawal) {
	s.withdraws(s.alice, withdrawals...)
}

func (s *MsgServerTestSuite) bobWithdraws(withdrawals ...*Withdrawal) {
	s.withdraws(s.bob, withdrawals...)
}

func (s *MsgServerTestSuite) carolWithdraws(withdrawals ...*Withdrawal) {
	s.withdraws(s.carol, withdrawals...)
}

func (s *MsgServerTestSuite) danWithdraws(withdrawals ...*Withdrawal) {
	s.withdraws(s.dan, withdrawals...)
}

func (s *MsgServerTestSuite) withdraws(account sdk.AccAddress, withdrawals ...*Withdrawal) {
	tickIndexes := make([]int64, len(withdrawals))
	fee := make([]uint64, len(withdrawals))
	sharesToRemove := make([]sdkmath.Int, len(withdrawals))
	for i, e := range withdrawals {
		tickIndexes[i] = e.TickIndex
		fee[i] = e.Fee
		sharesToRemove[i] = e.Shares
	}

	_, err := s.msgServer.Withdrawal(s.goCtx, &types.MsgWithdrawal{
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

func (s *MsgServerTestSuite) aliceWithdrawFails(expectedErr error, withdrawals ...*Withdrawal) {
	s.withdrawFails(s.alice, expectedErr, withdrawals...)
}

func (s *MsgServerTestSuite) bobWithdrawFails(expectedErr error, withdrawals ...*Withdrawal) {
	s.withdrawFails(s.bob, expectedErr, withdrawals...)
}

func (s *MsgServerTestSuite) carolWithdrawFails(expectedErr error, withdrawals ...*Withdrawal) {
	s.withdrawFails(s.carol, expectedErr, withdrawals...)
}

func (s *MsgServerTestSuite) danWithdrawFails(expectedErr error, withdrawals ...*Withdrawal) {
	s.withdrawFails(s.dan, expectedErr, withdrawals...)
}

func (s *MsgServerTestSuite) withdrawFails(
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

	_, err := s.msgServer.Withdrawal(s.goCtx, &types.MsgWithdrawal{
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

func (s *MsgServerTestSuite) aliceCancelsLimitSell(trancheKey string) {
	s.cancelsLimitSell(s.alice, trancheKey)
}

func (s *MsgServerTestSuite) bobCancelsLimitSell(trancheKey string) {
	s.cancelsLimitSell(s.bob, trancheKey)
}

func (s *MsgServerTestSuite) carolCancelsLimitSell(trancheKey string) {
	s.cancelsLimitSell(s.carol, trancheKey)
}

func (s *MsgServerTestSuite) danCancelsLimitSell(trancheKey string) {
	s.cancelsLimitSell(s.dan, trancheKey)
}

func (s *MsgServerTestSuite) cancelsLimitSell(account sdk.AccAddress, trancheKey string) {
	_, err := s.msgServer.CancelLimitOrder(s.goCtx, &types.MsgCancelLimitOrder{
		Creator:    account.String(),
		TrancheKey: trancheKey,
	})
	s.Assert().Nil(err)
}

func (s *MsgServerTestSuite) aliceCancelsLimitSellFails(trancheKey string, expectedErr error) {
	s.cancelsLimitSellFails(s.alice, trancheKey, expectedErr)
}

func (s *MsgServerTestSuite) bobCancelsLimitSellFails(trancheKey string, expectedErr error) {
	s.cancelsLimitSellFails(s.bob, trancheKey, expectedErr)
}

func (s *MsgServerTestSuite) carolCancelsLimitSellFails(trancheKey string, expectedErr error) {
	s.cancelsLimitSellFails(s.carol, trancheKey, expectedErr)
}

func (s *MsgServerTestSuite) danCancelsLimitSellFails(trancheKey string, expectedErr error) {
	s.cancelsLimitSellFails(s.dan, trancheKey, expectedErr)
}

func (s *MsgServerTestSuite) cancelsLimitSellFails(
	account sdk.AccAddress,
	trancheKey string,
	expectedErr error,
) {
	_, err := s.msgServer.CancelLimitOrder(s.goCtx, &types.MsgCancelLimitOrder{
		Creator:    account.String(),
		TrancheKey: trancheKey,
	})
	s.Assert().ErrorIs(err, expectedErr)
}

/// MultiHopSwap

func (s *MsgServerTestSuite) aliceMultiHopSwaps(
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwaps(s.alice, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *MsgServerTestSuite) bobMultiHopSwaps(
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwaps(s.bob, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *MsgServerTestSuite) carolMultiHopSwaps(
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwaps(s.carol, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *MsgServerTestSuite) danMultiHopSwaps(
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwaps(s.dan, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *MsgServerTestSuite) multiHopSwaps(
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
		sdkmath.NewInt(int64(amountIn)),
		exitLimitPrice,
		pickBest,
	)
	_, err := s.msgServer.MultiHopSwap(s.goCtx, msg)
	s.Assert().Nil(err)
}

func (s *MsgServerTestSuite) aliceEstimatesMultiHopSwap(
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
		AmountIn:       sdkmath.NewInt(int64(amountIn)),
		ExitLimitPrice: exitLimitPrice,
		PickBestRoute:  pickBest,
	}
	res, err := s.app.DexKeeper.EstimateMultiHopSwap(s.goCtx, msg)
	s.Require().Nil(err)
	return res.CoinOut
}

func (s *MsgServerTestSuite) aliceEstimatesMultiHopSwapFails(
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
		AmountIn:       sdkmath.NewInt(int64(amountIn)),
		ExitLimitPrice: exitLimitPrice,
		PickBestRoute:  pickBest,
	}
	_, err := s.app.DexKeeper.EstimateMultiHopSwap(s.goCtx, msg)
	s.Assert().ErrorIs(err, expectedErr)
}

func (s *MsgServerTestSuite) aliceMultiHopSwapFails(
	err error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwapFails(s.alice, err, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *MsgServerTestSuite) bobMultiHopSwapFails(
	err error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwapFails(s.bob, err, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *MsgServerTestSuite) carolMultiHopSwapFails(
	err error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwapFails(s.carol, err, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *MsgServerTestSuite) danMultiHopSwapFails(
	err error,
	routes [][]string,
	amountIn int,
	exitLimitPrice math_utils.PrecDec,
	pickBest bool,
) {
	s.multiHopSwapFails(s.dan, err, routes, amountIn, exitLimitPrice, pickBest)
}

func (s *MsgServerTestSuite) multiHopSwapFails(
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
		sdkmath.NewInt(int64(amountIn)),
		exitLimitPrice,
		pickBest,
	)
	_, err := s.msgServer.MultiHopSwap(s.goCtx, msg)
	s.Assert().ErrorIs(err, expectedErr)
}

/// Withdraw filled limit order

func (s *MsgServerTestSuite) aliceWithdrawsLimitSell(trancheKey string) {
	s.withdrawsLimitSell(s.alice, trancheKey)
}

func (s *MsgServerTestSuite) bobWithdrawsLimitSell(trancheKey string) {
	s.withdrawsLimitSell(s.bob, trancheKey)
}

func (s *MsgServerTestSuite) carolWithdrawsLimitSell(trancheKey string) {
	s.withdrawsLimitSell(s.carol, trancheKey)
}

func (s *MsgServerTestSuite) danWithdrawsLimitSell(trancheKey string) {
	s.withdrawsLimitSell(s.dan, trancheKey)
}

func (s *MsgServerTestSuite) withdrawsLimitSell(account sdk.AccAddress, trancheKey string) {
	_, err := s.msgServer.WithdrawFilledLimitOrder(s.goCtx, &types.MsgWithdrawFilledLimitOrder{
		Creator:    account.String(),
		TrancheKey: trancheKey,
	})
	s.Assert().Nil(err)
}

func (s *MsgServerTestSuite) aliceWithdrawLimitSellFails(expectedErr error, trancheKey string) {
	s.withdrawLimitSellFails(s.alice, expectedErr, trancheKey)
}

func (s *MsgServerTestSuite) bobWithdrawLimitSellFails(expectedErr error, trancheKey string) {
	s.withdrawLimitSellFails(s.bob, expectedErr, trancheKey)
}

func (s *MsgServerTestSuite) carolWithdrawLimitSellFails(expectedErr error, trancheKey string) {
	s.withdrawLimitSellFails(s.carol, expectedErr, trancheKey)
}

func (s *MsgServerTestSuite) danWithdrawLimitSellFails(expectedErr error, trancheKey string) {
	s.withdrawLimitSellFails(s.dan, expectedErr, trancheKey)
}

func (s *MsgServerTestSuite) withdrawLimitSellFails(
	account sdk.AccAddress,
	expectedErr error,
	trancheKey string,
) {
	_, err := s.msgServer.WithdrawFilledLimitOrder(s.goCtx, &types.MsgWithdrawFilledLimitOrder{
		Creator:    account.String(),
		TrancheKey: trancheKey,
	})
	s.Assert().ErrorIs(err, expectedErr)
}

// Shares
func (s *MsgServerTestSuite) getPoolShares(
	token0 string,
	token1 string,
	tick int64,
	fee uint64,
) (shares sdkmath.Int) {
	poolID, found := s.app.DexKeeper.GetPoolIDByParams(s.ctx, &types.PairID{Token0: token0, Token1: token1}, tick, fee)
	if !found {
		return sdkmath.ZeroInt()
	}
	poolDenom := types.NewPoolDenom(poolID)
	return s.app.BankKeeper.GetSupply(s.ctx, poolDenom).Amount
}

func (s *MsgServerTestSuite) assertPoolShares(
	tick int64,
	fee uint64,
	sharesExpected uint64,
) {
	sharesExpectedInt := sdkmath.NewIntFromUint64(sharesExpected)
	sharesOwned := s.getPoolShares("TokenA", "TokenB", tick, fee)
	s.Assert().Equal(sharesExpectedInt, sharesOwned)
}

func (s *MsgServerTestSuite) getAccountShares(
	account sdk.AccAddress,
	token0 string,
	token1 string,
	tick int64,
	fee uint64,
) (shares sdkmath.Int) {
	id, found := s.app.DexKeeper.GetPoolIDByParams(s.ctx, types.MustNewPairID(token0, token1), tick, fee)
	if !found {
		return sdkmath.ZeroInt()
	}

	poolDenom := types.NewPoolDenom(id)
	return s.app.BankKeeper.GetBalance(s.ctx, account, poolDenom).Amount
}

func (s *MsgServerTestSuite) assertAccountShares(
	account sdk.AccAddress,
	tick int64,
	fee uint64,
	sharesExpected uint64,
) {
	sharesExpectedInt := sdkmath.NewIntFromUint64(sharesExpected)
	sharesOwned := s.getAccountShares(account, "TokenA", "TokenB", tick, fee)
	s.Assert().
		Equal(sharesExpectedInt, sharesOwned, "expected %s != actual %s", sharesExpected, sharesOwned)
}

func (s *MsgServerTestSuite) assertAliceShares(tick int64, fee, sharesExpected uint64) {
	s.assertAccountShares(s.alice, tick, fee, sharesExpected)
}

func (s *MsgServerTestSuite) assertBobShares(tick int64, fee, sharesExpected uint64) {
	s.assertAccountShares(s.bob, tick, fee, sharesExpected)
}

func (s *MsgServerTestSuite) assertCarolShares(tick int64, fee, sharesExpected uint64) {
	s.assertAccountShares(s.carol, tick, fee, sharesExpected)
}

func (s *MsgServerTestSuite) assertDanShares(tick int64, fee, sharesExpected uint64) {
	s.assertAccountShares(s.dan, tick, fee, sharesExpected)
}

// Ticks
func (s *MsgServerTestSuite) assertCurrentTicks(
	expected1To0 int64,
	expected0To1 int64,
) {
	s.assertCurr0To1(expected0To1)
	s.assertCurr1To0(expected1To0)
}

func (s *MsgServerTestSuite) assertCurr0To1(curr0To1Expected int64) {
	curr0To1Actual, found := s.app.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(
		s.ctx,
		defaultTradePairID0To1,
	)
	if curr0To1Expected == math.MaxInt64 {
		s.Assert().False(found)
	} else {
		s.Assert().Equal(curr0To1Expected, curr0To1Actual)
	}
}

func (s *MsgServerTestSuite) assertCurr1To0(curr1To0Expected int64) {
	curr1to0Actual, found := s.app.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(
		s.ctx,
		defaultTradePairID1To0,
	)
	if curr1To0Expected == math.MinInt64 {
		s.Assert().False(found)
	} else {
		s.Assert().Equal(curr1To0Expected, curr1to0Actual)
	}
}

// Pool liquidity (i.e. deposited rather than LO)
func (s *MsgServerTestSuite) assertLiquidityAtTick(
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

func (s *MsgServerTestSuite) assertLiquidityAtTickWithDenom(
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

func (s *MsgServerTestSuite) assertPoolLiquidity(
	amountA, amountB int,
	tickIndex int64,
	fee uint64,
) {
	s.assertLiquidityAtTick(sdkmath.NewInt(int64(amountA)), sdkmath.NewInt(int64(amountB)), tickIndex, fee)
}

func (s *MsgServerTestSuite) assertNoLiquidityAtTick(tickIndex int64, fee uint64) {
	s.assertLiquidityAtTick(sdkmath.ZeroInt(), sdkmath.ZeroInt(), tickIndex, fee)
}

// Filled limit liquidity
func (s *MsgServerTestSuite) assertAliceLimitFilledAtTickAtIndex(
	selling string,
	amount int,
	tickIndex int64,
	trancheKey string,
) {
	s.assertLimitFilledAtTickAtIndex(s.alice, selling, amount, tickIndex, trancheKey)
}

func (s *MsgServerTestSuite) assertBobLimitFilledAtTickAtIndex(
	selling string,
	amount int,
	tickIndex int64,
	trancheKey string,
) {
	s.assertLimitFilledAtTickAtIndex(s.bob, selling, amount, tickIndex, trancheKey)
}

func (s *MsgServerTestSuite) assertCarolLimitFilledAtTickAtIndex(
	selling string,
	amount int,
	tickIndex int64,
	trancheKey string,
) {
	s.assertLimitFilledAtTickAtIndex(s.carol, selling, amount, tickIndex, trancheKey)
}

func (s *MsgServerTestSuite) assertDanLimitFilledAtTickAtIndex(
	selling string,
	amount int,
	tickIndex int64,
	trancheKey string,
) {
	s.assertLimitFilledAtTickAtIndex(s.dan, selling, amount, tickIndex, trancheKey)
}

func (s *MsgServerTestSuite) assertLimitFilledAtTickAtIndex(
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
	amt := sdkmath.NewInt(int64(amount))
	userFilled := userRatio.MulInt(filled).RoundInt()
	s.Assert().True(amt.Equal(userFilled))
}

// Limit liquidity
func (s *MsgServerTestSuite) assertAliceLimitLiquidityAtTick(
	selling string,
	amount int,
	tickIndex int64,
) {
	s.assertAccountLimitLiquidityAtTick(s.alice, selling, amount, tickIndex)
}

func (s *MsgServerTestSuite) assertBobLimitLiquidityAtTick(
	selling string,
	amount int,
	tickIndex int64,
) {
	s.assertAccountLimitLiquidityAtTick(s.bob, selling, amount, tickIndex)
}

func (s *MsgServerTestSuite) assertCarolLimitLiquidityAtTick(
	selling string,
	amount int,
	tickIndex int64,
) {
	s.assertAccountLimitLiquidityAtTick(s.carol, selling, amount, tickIndex)
}

func (s *MsgServerTestSuite) assertDanLimitLiquidityAtTick(
	selling string,
	amount int,
	tickIndex int64,
) {
	s.assertAccountLimitLiquidityAtTick(s.dan, selling, amount, tickIndex)
}

func (s *MsgServerTestSuite) assertAccountLimitLiquidityAtTick(
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

func (s *MsgServerTestSuite) assertLimitLiquidityAtTick(
	selling string,
	tickIndexNormalized, amount int64,
) {
	s.assertLimitLiquidityAtTickInt(selling, tickIndexNormalized, sdkmath.NewInt(amount))
}

func (s *MsgServerTestSuite) assertLimitLiquidityAtTickInt(
	selling string,
	tickIndexNormalized int64,
	amount sdkmath.Int,
) {
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(tickIndexNormalized)
	tranches := s.app.DexKeeper.GetAllLimitOrderTrancheAtIndex(
		s.ctx,
		tradePairID,
		tickIndexTakerToMaker,
	)
	liquidity := sdkmath.ZeroInt()
	for _, t := range tranches {
		if !t.IsExpired(s.ctx) {
			liquidity = liquidity.Add(t.ReservesMakerDenom)
		}
	}

	s.Assert().
		True(amount.Equal(liquidity), "Incorrect liquidity: expected %s, have %s", amount.String(), liquidity.String())
}

func (s *MsgServerTestSuite) assertFillAndPlaceTrancheKeys(
	selling string,
	tickIndexNormalized int64,
	expectedFill, expectedPlace string,
) {
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(tickIndexNormalized)
	placeTranche := s.app.DexKeeper.GetPlaceTranche(s.ctx, tradePairID, tickIndexTakerToMaker)
	fillTranche, foundFill := s.app.DexKeeper.GetFillTranche(
		s.ctx,
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
func (s *MsgServerTestSuite) getLimitUserSharesAtTick(
	account sdk.AccAddress,
	selling string,
	tickIndexNormalized int64,
) sdkmath.Int {
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(tickIndexNormalized)
	tranches := s.app.DexKeeper.GetAllLimitOrderTrancheAtIndex(
		s.ctx,
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

func (s *MsgServerTestSuite) getLimitUserSharesAtTickAtIndex(
	account sdk.AccAddress,
	trancheKey string,
) sdkmath.Int {
	userShares, found := s.app.DexKeeper.GetLimitOrderTrancheUser(
		s.ctx,
		account.String(),
		trancheKey,
	)
	s.Assert().True(found, "Failed to get limit order user shares for index %s", trancheKey)
	return userShares.SharesOwned
}

func (s *MsgServerTestSuite) getLimitTotalSharesAtTick(
	selling string,
	tickIndexNormalized int64,
) sdkmath.Int {
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tickIndexTakerToMaker := tradePairID.TickIndexTakerToMaker(tickIndexNormalized)
	tranches := s.app.DexKeeper.GetAllLimitOrderTrancheAtIndex(
		s.ctx,
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

func (s *MsgServerTestSuite) getLimitFilledLiquidityAtTickAtIndex(
	selling string,
	tickIndex int64,
	trancheKey string,
) sdkmath.Int {
	// grab fill tranche reserves and shares
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tranche, _, found := s.app.DexKeeper.FindLimitOrderTranche(s.ctx, &types.LimitOrderTrancheKey{
		tradePairID,
		tickIndex,
		trancheKey,
	})
	s.Assert().True(found, "Failed to get limit order filled reserves for index %s", trancheKey)

	return tranche.ReservesTakerDenom
}

func (s *MsgServerTestSuite) getLimitReservesAtTickAtKey(
	selling string,
	tickIndex int64,
	trancheKey string,
) sdkmath.Int {
	// grab fill tranche reserves and shares
	tradePairID := defaultPairID.MustTradePairIDFromMaker(selling)
	tranche, _, found := s.app.DexKeeper.FindLimitOrderTranche(s.ctx, &types.LimitOrderTrancheKey{
		tradePairID, tickIndex, trancheKey,
	})
	s.Assert().True(found, "Failed to get limit order reserves for index %s", trancheKey)

	return tranche.ReservesMakerDenom
}

func (s *MsgServerTestSuite) assertNLimitOrderExpiration(expected int) {
	exps := s.app.DexKeeper.GetAllLimitOrderExpiration(s.ctx)
	s.Assert().Equal(expected, len(exps))
}

func (s *MsgServerTestSuite) calcAutoswapSharesMinted(
	centerTick int64,
	fee uint64,
	residual0, residual1, balanced0, balanced1, totalShares, valuePool int64,
) sdkmath.Int {
	residual0Int, residual1Int, balanced0Int, balanced1Int, totalSharesInt, valuePoolInt := sdkmath.NewInt(
		residual0,
	), sdkmath.NewInt(
		residual1,
	), sdkmath.NewInt(
		balanced0,
	), sdkmath.NewInt(
		balanced1,
	), sdkmath.NewInt(
		totalShares,
	), sdkmath.NewInt(
		valuePool,
	)

	// residualValue = 1.0001^-f * residualAmount0 + 1.0001^{i-f} * residualAmount1
	// balancedValue = balancedAmount0 + 1.0001^{i} * balancedAmount1
	// value = residualValue + balancedValue
	// shares minted = value * totalShares / valuePool

	centerPrice := types.MustCalcPrice(-1 * centerTick)
	leftPrice := types.MustCalcPrice(-1 * (centerTick - int64(fee)))
	discountPrice := types.MustCalcPrice(-1 * int64(fee))

	balancedValue := math_utils.NewPrecDecFromInt(balanced0Int).
		Add(centerPrice.MulInt(balanced1Int)).
		TruncateInt()
	residualValue := discountPrice.MulInt(residual0Int).
		Add(leftPrice.Mul(math_utils.NewPrecDecFromInt(residual1Int))).
		TruncateInt()
	valueMint := balancedValue.Add(residualValue)

	return valueMint.Mul(totalSharesInt).Quo(valuePoolInt)
}

func (s *MsgServerTestSuite) calcSharesMinted(centerTick, amount0Int, amount1Int int64) sdkmath.Int {
	amount0, amount1 := sdkmath.NewInt(amount0Int), sdkmath.NewInt(amount1Int)
	centerPrice := types.MustCalcPrice(-1 * centerTick)

	return math_utils.NewPrecDecFromInt(amount0).Add(centerPrice.Mul(math_utils.NewPrecDecFromInt(amount1))).TruncateInt()
}

func (s *MsgServerTestSuite) calcExpectedBalancesAfterWithdrawOnePool(
	sharesMinted sdkmath.Int,
	account sdk.AccAddress,
	tickIndex int64,
	fee uint64,
) (sdkmath.Int, sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	dexCurrentBalance0 := s.app.BankKeeper.GetBalance(
		s.ctx,
		s.app.AccountKeeper.GetModuleAddress("dex"),
		"TokenA",
	).Amount
	dexCurrentBalance1 := s.app.BankKeeper.GetBalance(
		s.ctx,
		s.app.AccountKeeper.GetModuleAddress("dex"),
		"TokenB",
	).Amount
	currentBalance0 := s.app.BankKeeper.GetBalance(s.ctx, account, "TokenA").Amount
	currentBalance1 := s.app.BankKeeper.GetBalance(s.ctx, account, "TokenB").Amount
	amountPool0, amountPool1 := s.getLiquidityAtTick(tickIndex, fee)
	poolShares := s.getPoolShares("TokenA", "TokenB", tickIndex, fee)

	amountOut0 := amountPool0.Mul(sharesMinted).Quo(poolShares)
	amountOut1 := amountPool1.Mul(sharesMinted).Quo(poolShares)

	expectedBalance0 := currentBalance0.Add(amountOut0)
	expectedBalance1 := currentBalance1.Add(amountOut1)
	dexExpectedBalance0 := dexCurrentBalance0.Sub(amountOut0)
	dexExpectedBalance1 := dexCurrentBalance1.Sub(amountOut1)

	return expectedBalance0, expectedBalance1, dexExpectedBalance0, dexExpectedBalance1
}

func (s *MsgServerTestSuite) nextBlockWithTime(blockTime time.Time) {
	newCtx := s.ctx.WithBlockTime(blockTime)
	s.ctx = newCtx
	s.goCtx = sdk.WrapSDKContext(newCtx)
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash,
		Time: blockTime,
	}})
}
