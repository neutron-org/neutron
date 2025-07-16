package dex_state_test

import (
	"errors"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	dextypes "github.com/neutron-org/neutron/v7/x/dex/types"
)

type cancelLimitOrderTestParams struct {
	// State Conditions
	SharedParams
	ExistingTokenAHolders string
	Filled                int64
	WithdrawnCreator      bool
	WithdrawnOneOther     bool
	Expired               bool
	OrderType             dextypes.LimitOrderType // JIT, GTT, GTC
}

func (p cancelLimitOrderTestParams) printTestInfo(t *testing.T) {
	t.Logf(`
		Existing Shareholders: %s
		Filled: %v
		WithdrawnCreator: %v
		WithdrawnOneOther: %t
		Expired: %t
		OrderType: %v`,
		p.ExistingTokenAHolders,
		p.Filled,
		p.WithdrawnCreator,
		p.WithdrawnOneOther,
		p.Expired,
		p.OrderType.String(),
	)
}

func hydrateCancelLoTestCase(params map[string]string) cancelLimitOrderTestParams {
	selltick, err := dextypes.CalcTickIndexFromPrice(math_utils.MustNewPrecDecFromStr(DefaultSellPrice))
	if err != nil {
		panic(err)
	}
	c := cancelLimitOrderTestParams{
		ExistingTokenAHolders: params["ExistingTokenAHolders"],
		Filled:                int64(parseInt(params["Filled"])),
		WithdrawnCreator:      parseBool(params["WithdrawnCreator"]),
		WithdrawnOneOther:     parseBool(params["WithdrawnOneOther"]),
		Expired:               parseBool(params["Expired"]),
		OrderType:             dextypes.LimitOrderType(dextypes.LimitOrderType_value[params["OrderType"]]),
	}
	c.SharedParams.Tick = selltick
	return c
}

func (s *DexStateTestSuite) setupCancelTest(params cancelLimitOrderTestParams) (tranche *dextypes.LimitOrderTranche) {
	coinA := sdk.NewCoin(params.PairID.Token0, BaseTokenAmountInt)
	coinB := sdk.NewCoin(params.PairID.Token1, BaseTokenAmountInt.MulRaw(10))
	s.FundAcc(s.creator, sdk.NewCoins(coinA))
	var expTime *time.Time
	if params.OrderType.IsGoodTil() {
		t := time.Now()
		expTime = &t
	}
	res := s.makePlaceLOSuccess(s.creator, coinA, coinB.Denom, DefaultSellPrice, params.OrderType, expTime)

	totalDeposited := BaseTokenAmountInt
	if params.ExistingTokenAHolders == OneOtherAndCreatorLO {
		totalDeposited = totalDeposited.MulRaw(2)
		s.FundAcc(s.alice, sdk.NewCoins(coinA))
		s.makePlaceLOSuccess(s.alice, coinA, coinB.Denom, DefaultSellPrice, params.OrderType, expTime)
	}

	if params.Filled > 0 {
		s.FundAcc(s.bob, sdk.NewCoins(coinB))
		fillAmount := totalDeposited.MulRaw(params.Filled).QuoRaw(100)
		_, err := s.makePlaceTakerLO(s.bob, coinB, coinA.Denom, DefaultBuyPriceTaker, dextypes.LimitOrderType_IMMEDIATE_OR_CANCEL, &fillAmount)
		s.NoError(err)
	}

	if params.WithdrawnCreator {
		s.makeWithdrawFilledSuccess(s.creator, res.TrancheKey)
	}

	if params.WithdrawnOneOther {
		s.makeWithdrawFilledSuccess(s.alice, res.TrancheKey)
	}

	if params.Expired {
		s.App.DexKeeper.PurgeExpiredLimitOrders(s.Ctx, time.Now())
	}

	req := dextypes.QueryGetLimitOrderTrancheRequest{
		PairId:     params.PairID.CanonicalString(),
		TickIndex:  params.Tick,
		TokenIn:    params.PairID.Token0,
		TrancheKey: res.TrancheKey,
	}
	tranchResp, err := s.App.DexKeeper.LimitOrderTranche(s.Ctx, &req)
	s.NoError(err)

	return tranchResp.LimitOrderTranche
}

func hydrateAllCancelLoTestCases(paramsList []map[string]string) []cancelLimitOrderTestParams {
	allTCs := make([]cancelLimitOrderTestParams, 0)
	for i, paramsRaw := range paramsList {
		tc := hydrateCancelLoTestCase(paramsRaw)

		pairID := generatePairID(i)
		tc.PairID = pairID

		allTCs = append(allTCs, tc)
	}

	return removeRedundantCancelLOTests(allTCs)
}

func removeRedundantCancelLOTests(params []cancelLimitOrderTestParams) []cancelLimitOrderTestParams {
	newParams := make([]cancelLimitOrderTestParams, 0)
	for _, p := range params {
		// it's impossible to withdraw 0 filled
		// error checks is not in a scope of the testcase (see withdraw filled test)
		if p.Filled == 0 && (p.WithdrawnOneOther || p.WithdrawnCreator) {
			continue
		}
		if p.Expired && p.OrderType.IsGTC() {
			continue
		}
		if p.WithdrawnOneOther && p.ExistingTokenAHolders == CreatorLO {
			continue
		}
		if p.ExistingTokenAHolders == OneOtherAndCreatorLO && !p.OrderType.IsGTC() {
			// user tranches combined into tranches only for LimitOrderType_GOOD_TIL_CANCELLED
			// it does not make any sense to create two tranches
			continue
		}

		if p.Filled == 100 &&
			(p.ExistingTokenAHolders == OneOtherAndCreatorLO && p.WithdrawnCreator && p.WithdrawnOneOther) ||
			(p.ExistingTokenAHolders == CreatorLO && p.WithdrawnCreator) {
			// if limit order is filled and withdrawn by all shareholders, it's impossible to cancel it
			continue
		}

		newParams = append(newParams, p)
	}
	return newParams
}

func (s *DexStateTestSuite) handleCancelErrors(params cancelLimitOrderTestParams, err error) {
	if params.Filled == 100 && params.WithdrawnCreator {
		if errors.Is(err, dextypes.ErrValidLimitOrderTrancheNotFound) {
			s.T().Skip()
		}
	}
	s.NoError(err)
}

func (s *DexStateTestSuite) assertCancelAmount(params cancelLimitOrderTestParams) {
	depositSize := BaseTokenAmountInt

	// expected balance: InitialBalance - depositSize + pre-withdrawn (filled/2 or 0) + withdrawn (filled/2 or filled)
	// pre-withdrawn (filled/2 or 0) + withdrawn (filled/2 or filled) === filled
	// converted to TokenB
	price := dextypes.MustCalcPrice(params.Tick)
	expectedBalanceB := price.MulInt(depositSize.MulRaw(params.Filled).QuoRaw(100)).TruncateInt()
	expectedBalanceA := depositSize.Sub(depositSize.MulRaw(params.Filled).QuoRaw(100))
	// 1 - withdrawn amount
	s.assertBalance(s.creator, params.PairID.Token1, expectedBalanceB)

	s.assertBalance(s.creator, params.PairID.Token0, expectedBalanceA)
}

func TestCancel(t *testing.T) {
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
	testCases := hydrateAllCancelLoTestCases(testCasesRaw)

	s := new(DexStateTestSuite)
	s.SetT(t)
	s.SetupTest()

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s.SetT(t)
			tc.printTestInfo(t)

			tranche := s.setupCancelTest(tc)

			_, err := s.makeCancel(s.creator, tranche.Key.TrancheKey)
			s.handleCancelErrors(tc, err)
			_, found := s.App.DexKeeper.GetLimitOrderTrancheUser(s.Ctx, s.creator.String(), tranche.Key.TrancheKey)
			s.False(found)
			s.assertCancelAmount(tc)
		})
	}
}
