package dex_state_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v8/utils/math"
	dextypes "github.com/neutron-org/neutron/v8/x/dex/types"
)

type withdrawLimitOrderTestParams struct {
	// State Conditions
	SharedParams
	ExistingTokenAHolders string
	Filled                int
	WithdrawnCreator      bool
	WithdrawnOneOther     bool
	Expired               bool
	OrderType             int32 // JIT, GTT, GTC
}

func (p withdrawLimitOrderTestParams) printTestInfo(t *testing.T) {
	t.Logf(`
		Existing Shareholders: %s
		Filled: %v
		WithdrawnCreator: %v
		WithdrawnOneOther: %t
		Expired: %t
		OrderType: %v`,
		p.ExistingTokenAHolders,
		p.Filled,
		// Two fields define a state with a pre-withdrawn tranche
		p.WithdrawnCreator,
		p.WithdrawnOneOther,

		p.Expired,
		p.OrderType,
	)
}

func hydrateWithdrawLoTestCase(params map[string]string) withdrawLimitOrderTestParams {
	selltick, err := dextypes.CalcTickIndexFromPrice(math_utils.MustNewPrecDecFromStr(DefaultSellPrice))
	if err != nil {
		panic(err)
	}
	w := withdrawLimitOrderTestParams{
		ExistingTokenAHolders: params["ExistingTokenAHolders"],
		Filled:                parseInt(params["Filled"]),
		WithdrawnCreator:      parseBool(params["WithdrawnCreator"]),
		WithdrawnOneOther:     parseBool(params["WithdrawnOneOther"]),
		Expired:               parseBool(params["Expired"]),
		OrderType:             dextypes.LimitOrderType_value[params["OrderType"]],
	}
	w.SharedParams.Tick = selltick
	return w
}

func (s *DexStateTestSuite) setupWithdrawLimitOrderTest(params withdrawLimitOrderTestParams) *dextypes.LimitOrderTranche {
	coinA := sdk.NewCoin(params.PairID.Token0, BaseTokenAmountInt)
	coinB := sdk.NewCoin(params.PairID.Token1, BaseTokenAmountInt.MulRaw(10))
	s.FundAcc(s.creator, sdk.NewCoins(coinA))
	var expTime *time.Time
	if params.OrderType == int32(dextypes.LimitOrderType_GOOD_TIL_TIME) {
		t := time.Now()
		expTime = &t
	}
	res := s.makePlaceLOSuccess(s.creator, coinA, coinB.Denom, DefaultSellPrice, dextypes.LimitOrderType(params.OrderType), expTime)

	totalDeposited := BaseTokenAmountInt
	if params.ExistingTokenAHolders == OneOtherAndCreatorLO {
		totalDeposited = totalDeposited.MulRaw(2)
		s.FundAcc(s.alice, sdk.NewCoins(coinA))
		s.makePlaceLOSuccess(s.alice, coinA, coinB.Denom, DefaultSellPrice, dextypes.LimitOrderType(params.OrderType), expTime)
	}

	// withdraw in two steps: before and after pre-withdraw (if there are any)
	halfAmount := totalDeposited.MulRaw(int64(params.Filled)).QuoRaw(2 * 100)
	s.FundAcc(s.bob, sdk.NewCoins(coinB).MulInt(math.NewInt(10)))
	if params.Filled > 0 {
		_, err := s.makePlaceTakerLO(s.bob, coinB, coinA.Denom, DefaultBuyPriceTaker, dextypes.LimitOrderType_IMMEDIATE_OR_CANCEL, &halfAmount)
		s.NoError(err)
	}

	if params.WithdrawnCreator {
		s.makeWithdrawFilledSuccess(s.creator, res.TrancheKey)
	}

	if params.WithdrawnOneOther {
		s.makeWithdrawFilledSuccess(s.alice, res.TrancheKey)
	}

	if params.Filled > 0 {
		_, err := s.makePlaceTakerLO(s.bob, coinB, coinA.Denom, DefaultBuyPriceTaker, dextypes.LimitOrderType_IMMEDIATE_OR_CANCEL, &halfAmount)
		s.NoError(err)
	}

	if params.Expired {
		s.App.DexKeeper.PurgeExpiredLimitOrders(s.Ctx, time.Now())
	}
	tick, err := dextypes.CalcTickIndexFromPrice(DefaultStartPrice)
	s.NoError(err)

	req := dextypes.QueryGetLimitOrderTrancheRequest{
		PairId:     params.PairID.CanonicalString(),
		TickIndex:  tick,
		TokenIn:    params.PairID.Token0,
		TrancheKey: res.TrancheKey,
	}
	tranchResp, err := s.App.DexKeeper.LimitOrderTranche(s.Ctx, &req)
	s.NoError(err)

	return tranchResp.LimitOrderTranche
}

func hydrateAllWithdrawLoTestCases(paramsList []map[string]string) []withdrawLimitOrderTestParams {
	allTCs := make([]withdrawLimitOrderTestParams, 0)
	for i, paramsRaw := range paramsList {
		pairID := generatePairID(i)
		tc := hydrateWithdrawLoTestCase(paramsRaw)
		tc.PairID = pairID
		allTCs = append(allTCs, tc)
	}

	// return allTCs
	return removeRedundantWithdrawLOTests(allTCs)
}

func removeRedundantWithdrawLOTests(params []withdrawLimitOrderTestParams) []withdrawLimitOrderTestParams {
	newParams := make([]withdrawLimitOrderTestParams, 0)
	for _, p := range params {
		// it's impossible to withdraw 0 filled
		if p.Filled == 0 && (p.WithdrawnOneOther || p.WithdrawnCreator) {
			continue
		}
		if p.Expired && p.OrderType == int32(dextypes.LimitOrderType_GOOD_TIL_CANCELLED) {
			continue
		}
		if p.WithdrawnOneOther && p.ExistingTokenAHolders == CreatorLO {
			continue
		}
		if p.ExistingTokenAHolders == OneOtherAndCreatorLO && p.OrderType != int32(dextypes.LimitOrderType_GOOD_TIL_CANCELLED) {
			// user tranches combined into tranches only for LimitOrderType_GOOD_TIL_CANCELLED
			// it does not make any sense to create two tranches
			continue
		}
		newParams = append(newParams, p)
	}
	return newParams
}

func (s *DexStateTestSuite) handleWithdrawLimitOrderErrors(params withdrawLimitOrderTestParams, err error) {
	if params.Filled == 0 {
		if errors.Is(err, dextypes.ErrWithdrawEmptyLimitOrder) {
			s.T().Skip()
		}
	}
	s.NoError(err)
}

func (s *DexStateTestSuite) assertWithdrawFilledAmount(params withdrawLimitOrderTestParams, trancheKey string) {
	depositSize := BaseTokenAmountInt

	// expected balance: InitialBalance - depositSize + pre-withdrawn (filled/2 or 0) + withdrawn (filled/2 or filled)
	// pre-withdrawn (filled/2 or 0) + withdrawn (filled/2 or filled) === filled
	// converted to TokenB
	price := dextypes.MustCalcPrice(params.Tick)
	expectedBalanceB := price.MulInt(depositSize.MulRaw(int64(params.Filled)).QuoRaw(100)).Ceil().TruncateInt()
	expectedBalanceA := depositSize.Sub(depositSize.MulRaw(int64(params.Filled)).QuoRaw(100))
	// 1 - withdrawn amount
	s.assertBalanceWithPrecision(s.creator, params.PairID.Token1, expectedBalanceB, 3)

	ut, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.creator.String(), trancheKey)
	if params.Expired {
		// "canceled" amount
		s.assertBalance(s.creator, params.PairID.Token0, expectedBalanceA)
		s.False(found)
	} else {
		s.assertBalance(s.creator, params.PairID.Token0, math.ZeroInt())
		if params.Filled == 100 {
			s.False(found)
		} else {
			s.True(found)
			s.intsApproxEqual("", expectedBalanceA, ut.SharesOwned.Sub(ut.SharesWithdrawn), 1)
		}
	}
}

func TestWithdrawLimitOrder(t *testing.T) {
	testParams := []testParams{
		{field: "ExistingTokenAHolders", states: []string{CreatorLO, OneOtherAndCreatorLO}},
		{field: "Filled", states: []string{ZeroPCT, FiftyPCT, HundredPct}},
		{field: "WithdrawnCreator", states: []string{True, False}},
		{field: "WithdrawnOneOther", states: []string{True, False}},
		{field: "OrderType", states: []string{
			dextypes.LimitOrderType_name[int32(dextypes.LimitOrderType_GOOD_TIL_CANCELLED)],
			dextypes.LimitOrderType_name[int32(dextypes.LimitOrderType_GOOD_TIL_TIME)],
			dextypes.LimitOrderType_name[int32(dextypes.LimitOrderType_JUST_IN_TIME)],
		}},
		{field: "Expired", states: []string{True, False}},
	}
	testCasesRaw := generatePermutations(testParams)
	testCases := hydrateAllWithdrawLoTestCases(testCasesRaw)

	s := new(DexStateTestSuite)
	s.SetT(t)
	s.SetupTest()
	// totalExpectedToSwap := math.ZeroInt()

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s.SetT(t)
			tc.printTestInfo(t)

			initialTrancheKey := s.setupWithdrawLimitOrderTest(tc)
			fmt.Println(initialTrancheKey)

			resp, err := s.makeWithdrawFilled(s.creator, initialTrancheKey.Key.TrancheKey)
			s.handleWithdrawLimitOrderErrors(tc, err)
			fmt.Println("resp", resp)
			fmt.Println("err", err)
			s.assertWithdrawFilledAmount(tc, initialTrancheKey.Key.TrancheKey)
			/*
				   3. Assertions
					   1. (Value returned + remaining LO value)/ValueIn ~= LimitPrice
					   2. TakerDenom withdrawn == userOwnershipRatio * fillPercentage * takerReserves
					   3. If expired
						   1. MakerDenom withdraw ==  userOwnershipRatio * fillPercentage * makerReserves
			*/
		})
	}
}
