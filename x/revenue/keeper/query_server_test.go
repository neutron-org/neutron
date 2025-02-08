package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	appconfig "github.com/neutron-org/neutron/v5/app/config"
	testutil_keeper "github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	revenuekeeper "github.com/neutron-org/neutron/v5/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/require"
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

	params.BaseCompensation = 11111
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

	k, ctx := testutil_keeper.RevenueKeeper(t, nil, nil, "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a")
	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         10,
		CurrentPeriodStartBlock: 1,
	}
	require.Nil(t, k.SetPaymentScheduleI(ctx, ps))
	require.Nil(t, k.CalcNewCumulativePrice(ctx, math.LegacyMustNewDecFromStr("0.5"), ctx.BlockTime().Unix()))

	queryServer := revenuekeeper.NewQueryServerImpl(k)

	paymentInfo, err := queryServer.PaymentInfo(ctx, &revenuetypes.QueryPaymentInfoRequest{})
	require.Nil(t, err)
	require.Equal(t, ps, paymentInfo.
		PaymentSchedule.
		PaymentSchedule.(*revenuetypes.PaymentSchedule_BlockBasedPaymentSchedule).
		BlockBasedPaymentSchedule,
	)
	require.Equal(t, revenuetypes.RewardDenom, paymentInfo.RewardDenom)
	require.Equal(t, math.LegacyNewDecWithPrec(5, 1), paymentInfo.RewardDenomTwap)
	require.Equal(t, math.NewInt(5000), paymentInfo.BaseRevenueAmount)
}

func TestQueryValidatorStats(t *testing.T) {
	appconfig.GetDefaultConfig()

	k, ctx := testutil_keeper.RevenueKeeper(t, nil, nil, "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a")
	params := revenuetypes.DefaultParams()
	require.Nil(t, k.SetParams(ctx, params))
	queryServer := revenuekeeper.NewQueryServerImpl(k)

	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         500,
		CurrentPeriodStartBlock: 1,
	}
	require.Nil(t, k.SetPaymentScheduleI(ctx, ps))
	require.Nil(t, k.CalcNewCumulativePrice(ctx, math.LegacyMustNewDecFromStr("0.5"), ctx.BlockTime().Unix()))

	// val 1 with 100/100 performance (ctx.WithBlockHeight(100))
	val1 := val1Info()
	val1.CommitedBlocksInPeriod = 100
	val1.CommitedOracleVotesInPeriod = 100
	require.Nil(t, k.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1.ValOperAddress), val1))
	// val 2 with 50/100 performance (ctx.WithBlockHeight(100))
	val2 := val2Info()
	val2.CommitedBlocksInPeriod = 50
	val2.CommitedOracleVotesInPeriod = 50
	require.Nil(t, k.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val2.ValOperAddress), val2))

	val1Stats, err := queryServer.ValidatorStats(ctx.WithBlockHeight(100), &revenuetypes.QueryValidatorStatsRequest{
		ValOperAddress: val1.ValOperAddress,
	})
	require.Nil(t, err)
	require.Equal(t, uint64(100), val1Stats.Stats.ValidatorInfo.CommitedBlocksInPeriod)
	require.Equal(t, uint64(100), val1Stats.Stats.ValidatorInfo.CommitedOracleVotesInPeriod)
	require.Equal(t, math.LegacyOneDec(), val1Stats.Stats.PerformanceRating)
	// only 1 price in TWAP storage, take it as TWAP
	// TWAP = 0.5 USD/NTRN
	// total NTRN = 2500/0.5
	require.Equal(t, math.NewIntFromUint64(5000), val1Stats.Stats.ExpectedRevenue)

	val2Stats, err := queryServer.ValidatorStats(ctx.WithBlockHeight(100), &revenuetypes.QueryValidatorStatsRequest{
		ValOperAddress: val2.ValOperAddress,
	})
	require.Nil(t, err)
	require.Equal(t, uint64(50), val2Stats.Stats.ValidatorInfo.CommitedBlocksInPeriod)
	require.Equal(t, uint64(50), val2Stats.Stats.ValidatorInfo.CommitedOracleVotesInPeriod)
	require.Equal(t, math.LegacyZeroDec(), val2Stats.Stats.PerformanceRating)
	require.Equal(t, math.ZeroInt(), val2Stats.Stats.ExpectedRevenue)

	valsStats, err := queryServer.ValidatorsStats(ctx.WithBlockHeight(100), &revenuetypes.QueryValidatorsStatsRequest{})
	require.Nil(t, err)
	require.Equal(t, 2, len(valsStats.Stats))
	require.Equal(t, val1Stats.Stats, valsStats.Stats[1])
	require.Equal(t, val2Stats.Stats, valsStats.Stats[0])
}
