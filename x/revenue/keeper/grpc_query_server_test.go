package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	appconfig "github.com/neutron-org/neutron/v6/app/config"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/revenue/types"
	testutil_keeper "github.com/neutron-org/neutron/v6/testutil/revenue/keeper"
	revenuekeeper "github.com/neutron-org/neutron/v6/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

func TestQueryParams(t *testing.T) {
	appconfig.GetDefaultConfig()

	k, ctx := testutil_keeper.RevenueKeeper(t, nil, nil, "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a")
	params := revenuetypes.DefaultParams()
	require.Nil(t, k.SetParams(ctx, params))
	queryServer := revenuekeeper.NewQueryServerImpl(k)

	// test default params
	paramsResp, err := queryServer.Params(ctx, &revenuetypes.QueryParamsRequest{})
	require.Nil(t, err)
	require.Equal(t, params, paramsResp.Params)

	params.RewardQuote.Amount = 11111
	paramsResp, err = queryServer.Params(ctx, &revenuetypes.QueryParamsRequest{})
	require.Nil(t, err)
	require.NotEqual(t, params, paramsResp.Params) // not set yet

	// test set params
	require.Nil(t, k.SetParams(ctx, params))
	paramsResp, err = queryServer.Params(ctx, &revenuetypes.QueryParamsRequest{})
	require.Nil(t, err)
	require.Equal(t, params, paramsResp.Params)
}

func TestQueryPaymentInfo(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	k, ctx := testutil_keeper.RevenueKeeper(t, bankKeeper, nil, "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a")
	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         10,
		CurrentPeriodStartBlock: 1,
	}
	require.Nil(t, k.SetPaymentScheduleI(ctx, ps))
	require.Nil(t, k.CalcNewRewardAssetPrice(ctx, math.LegacyMustNewDecFromStr("0.5"), ctx.BlockTime().Unix()))

	bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
		Base:       "untrn", Symbol: "NTRN",
	}, true).AnyTimes()

	queryServer := revenuekeeper.NewQueryServerImpl(k)

	paymentInfo, err := queryServer.PaymentInfo(ctx.WithBlockHeight(11), &revenuetypes.QueryPaymentInfoRequest{})
	require.Nil(t, err)
	require.Equal(t, ps, paymentInfo.
		PaymentSchedule.
		PaymentSchedule.(*revenuetypes.PaymentSchedule_BlockBasedPaymentSchedule).
		BlockBasedPaymentSchedule,
	)
	require.Equal(t, math.LegacyNewDec(1), paymentInfo.EffectivePeriodProgress)
	require.Equal(t, math.LegacyNewDecWithPrec(5, 1), paymentInfo.RewardAssetTwap)
	// RewardQuote.Amount / TWAP * exp = 2500 / 0.5 * 10^6 = 5 000 NTRN = 5 000 * 10^6 untrn
	require.Equal(t, math.NewInt(5000000000), paymentInfo.BaseRevenueAmount.Amount)
	require.Equal(t, "untrn", paymentInfo.BaseRevenueAmount.Denom)

	// query payment info in the middle of a period
	paymentInfo, err = queryServer.PaymentInfo(ctx.WithBlockHeight(6), &revenuetypes.QueryPaymentInfoRequest{})
	require.Nil(t, err)
	require.Equal(t, math.LegacyNewDecWithPrec(5, 1), paymentInfo.EffectivePeriodProgress)
}

func TestQueryValidatorStats(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	k, ctx := testutil_keeper.RevenueKeeper(t, bankKeeper, nil, "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a")
	params := revenuetypes.DefaultParams()
	require.Nil(t, k.SetParams(ctx, params))
	queryServer := revenuekeeper.NewQueryServerImpl(k)

	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         100,
		CurrentPeriodStartBlock: 1,
	}
	require.Nil(t, k.SetPaymentScheduleI(ctx, ps))
	require.Nil(t, k.CalcNewRewardAssetPrice(ctx, math.LegacyMustNewDecFromStr("0.5"), ctx.BlockTime().Unix()))

	bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
		Base:       "untrn", Symbol: "NTRN",
	}, true).AnyTimes()

	t.Run("EndOfPeriod", func(t *testing.T) {
		// val 1 with 100/100 performance (ctx.WithBlockHeight(101), CurrentPeriodStartBlock: 1)
		val1 := val1Info()
		val1.CommitedBlocksInPeriod = 100
		val1.CommitedOracleVotesInPeriod = 100
		val1.InActiveValsetForBlocksInPeriod = 100
		require.Nil(t, k.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1.ValOperAddress), val1))
		// val 2 with 50/100 performance (ctx.WithBlockHeight(101), CurrentPeriodStartBlock: 1)
		val2 := val2Info()
		val2.CommitedBlocksInPeriod = 50
		val2.CommitedOracleVotesInPeriod = 50
		val2.InActiveValsetForBlocksInPeriod = 100
		require.Nil(t, k.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val2.ValOperAddress), val2))

		val1Stats, err := queryServer.ValidatorStats(ctx.WithBlockHeight(101), &revenuetypes.QueryValidatorStatsRequest{
			ValOperAddress: val1.ValOperAddress,
		})
		require.Nil(t, err)
		require.Equal(t, uint64(100), val1Stats.Stats.TotalProducedBlocksInPeriod)
		require.Equal(t, uint64(100), val1Stats.Stats.ValidatorInfo.CommitedBlocksInPeriod)
		require.Equal(t, uint64(100), val1Stats.Stats.ValidatorInfo.CommitedOracleVotesInPeriod)
		require.Equal(t, uint64(100), val1Stats.Stats.ValidatorInfo.InActiveValsetForBlocksInPeriod)
		require.Equal(t, math.LegacyOneDec(), val1Stats.Stats.PerformanceRating)
		// RewardQuote.Amount / TWAP * exp = 2500 / 0.5 * 10^6 = 5 000 NTRN = 5 000 * 10^6 untrn
		require.Equal(t, math.NewIntFromUint64(5000000000), val1Stats.Stats.ExpectedRevenue.Amount)
		require.Equal(t, "untrn", val1Stats.Stats.ExpectedRevenue.Denom)

		val2Stats, err := queryServer.ValidatorStats(ctx.WithBlockHeight(101), &revenuetypes.QueryValidatorStatsRequest{
			ValOperAddress: val2.ValOperAddress,
		})
		require.Nil(t, err)
		require.Equal(t, uint64(100), val2Stats.Stats.TotalProducedBlocksInPeriod)
		require.Equal(t, uint64(50), val2Stats.Stats.ValidatorInfo.CommitedBlocksInPeriod)
		require.Equal(t, uint64(50), val2Stats.Stats.ValidatorInfo.CommitedOracleVotesInPeriod)
		require.Equal(t, uint64(100), val2Stats.Stats.ValidatorInfo.InActiveValsetForBlocksInPeriod)
		require.Equal(t, math.LegacyZeroDec(), val2Stats.Stats.PerformanceRating)
		require.Equal(t, math.ZeroInt(), val2Stats.Stats.ExpectedRevenue.Amount)
		require.Equal(t, "untrn", val2Stats.Stats.ExpectedRevenue.Denom)

		valsStats, err := queryServer.ValidatorsStats(ctx.WithBlockHeight(101), &revenuetypes.QueryValidatorsStatsRequest{})
		require.Nil(t, err)
		require.Equal(t, 2, len(valsStats.Stats))
		require.Equal(t, val1Stats.Stats, valsStats.Stats[1])
		require.Equal(t, val2Stats.Stats, valsStats.Stats[0])
	})

	t.Run("MiddleOfPeriod", func(t *testing.T) {
		// val 1 with 80/80 performance (ctx.WithBlockHeight(81), CurrentPeriodStartBlock: 1)
		val1 := val1Info()
		val1.CommitedBlocksInPeriod = 80
		val1.CommitedOracleVotesInPeriod = 80
		val1.InActiveValsetForBlocksInPeriod = 80
		require.Nil(t, k.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1.ValOperAddress), val1))
		// val 2 with 40/80 performance (ctx.WithBlockHeight(81), CurrentPeriodStartBlock: 1)
		val2 := val2Info()
		val2.CommitedBlocksInPeriod = 40
		val2.CommitedOracleVotesInPeriod = 40
		val2.InActiveValsetForBlocksInPeriod = 80
		require.Nil(t, k.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val2.ValOperAddress), val2))

		// expect the effective period progress to reflect the query context
		paymentInfo, err := queryServer.PaymentInfo(ctx.WithBlockHeight(81), &revenuetypes.QueryPaymentInfoRequest{})
		require.Nil(t, err)
		require.Equal(t, paymentInfo.EffectivePeriodProgress, math.LegacyNewDecWithPrec(8, 1))

		val1Stats, err := queryServer.ValidatorStats(ctx.WithBlockHeight(81), &revenuetypes.QueryValidatorStatsRequest{
			ValOperAddress: val1.ValOperAddress,
		})
		require.Nil(t, err)
		require.Equal(t, uint64(80), val1Stats.Stats.TotalProducedBlocksInPeriod)
		require.Equal(t, uint64(80), val1Stats.Stats.ValidatorInfo.CommitedBlocksInPeriod)
		require.Equal(t, uint64(80), val1Stats.Stats.ValidatorInfo.CommitedOracleVotesInPeriod)
		require.Equal(t, uint64(80), val1Stats.Stats.ValidatorInfo.InActiveValsetForBlocksInPeriod)
		require.Equal(t, math.LegacyOneDec(), val1Stats.Stats.PerformanceRating)
		// RewardQuote.Amount / TWAP * exp = 2500 / 0.5 * 10^6 = 5 000 NTRN = 5 000 * 10^6 untrn
		require.Equal(t, math.NewIntFromUint64(5000000000), val1Stats.Stats.ExpectedRevenue.Amount)
		require.Equal(t, "untrn", val1Stats.Stats.ExpectedRevenue.Denom)

		val2Stats, err := queryServer.ValidatorStats(ctx.WithBlockHeight(81), &revenuetypes.QueryValidatorStatsRequest{
			ValOperAddress: val2.ValOperAddress,
		})
		require.Nil(t, err)
		require.Equal(t, uint64(80), val2Stats.Stats.TotalProducedBlocksInPeriod)
		require.Equal(t, uint64(40), val2Stats.Stats.ValidatorInfo.CommitedBlocksInPeriod)
		require.Equal(t, uint64(40), val2Stats.Stats.ValidatorInfo.CommitedOracleVotesInPeriod)
		require.Equal(t, uint64(80), val2Stats.Stats.ValidatorInfo.InActiveValsetForBlocksInPeriod)
		require.Equal(t, math.LegacyZeroDec(), val2Stats.Stats.PerformanceRating)
		require.Equal(t, math.ZeroInt(), val2Stats.Stats.ExpectedRevenue.Amount)
		require.Equal(t, "untrn", val2Stats.Stats.ExpectedRevenue.Denom)

		valsStats, err := queryServer.ValidatorsStats(ctx.WithBlockHeight(81), &revenuetypes.QueryValidatorsStatsRequest{})
		require.Nil(t, err)
		require.Equal(t, 2, len(valsStats.Stats))
		require.Equal(t, val1Stats.Stats, valsStats.Stats[1])
		require.Equal(t, val2Stats.Stats, valsStats.Stats[0])
	})
}
