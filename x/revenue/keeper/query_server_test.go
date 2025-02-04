package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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

func TestQueryState(t *testing.T) {
	appconfig.GetDefaultConfig()

	k, ctx := testutil_keeper.RevenueKeeper(t, nil, nil, "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a")
	ps := revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         10,
		CurrentPeriodStartBlock: 1,
	}
	psAny, err := codectypes.NewAnyWithValue(&ps)
	require.Nil(t, err)
	require.Nil(t, k.SetState(ctx, revenuetypes.State{PaymentSchedule: psAny}))
	queryServer := revenuekeeper.NewQueryServerImpl(k)

	state, err := queryServer.State(ctx, &revenuetypes.QueryStateRequest{})
	require.Nil(t, err)
	require.Equal(t, psAny, state.State.PaymentSchedule)
}

func TestQueryValidatorStats(t *testing.T) {
	appconfig.GetDefaultConfig()

	k, ctx := testutil_keeper.RevenueKeeper(t, nil, nil, "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a")
	params := revenuetypes.DefaultParams()
	require.Nil(t, k.SetParams(ctx, params))
	queryServer := revenuekeeper.NewQueryServerImpl(k)

	ps := revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         500,
		CurrentPeriodStartBlock: 1,
	}
	psAny, err := codectypes.NewAnyWithValue(&ps)
	require.Nil(t, err)
	require.Nil(t, k.SetState(ctx, revenuetypes.State{PaymentSchedule: psAny}))

	// val 1 with 100/100 performance (ctx.WithBlockHeight(100))
	val1 := val1Info()
	val1.CommitedBlocksInPeriod = 100
	val1.CommitedOracleVotesInPeriod = 100
	require.Nil(t, k.SetValidatorInfo(ctx, mustConsAddressFromBech32(t, val1.ConsensusAddress), val1))
	// val 2 with 50/100 performance (ctx.WithBlockHeight(100))
	val2 := val2Info()
	val2.CommitedBlocksInPeriod = 50
	val2.CommitedOracleVotesInPeriod = 50
	require.Nil(t, k.SetValidatorInfo(ctx, mustConsAddressFromBech32(t, val2.ConsensusAddress), val2))

	val1Stats, err := queryServer.ValidatorStats(ctx.WithBlockHeight(100), &revenuetypes.QueryValidatorStatsRequest{
		ConsensusAddress: val1.ConsensusAddress,
	})
	require.Nil(t, err)
	require.Equal(t, uint64(100), val1Stats.Stats.ValidatorInfo.CommitedBlocksInPeriod)
	require.Equal(t, uint64(100), val1Stats.Stats.ValidatorInfo.CommitedOracleVotesInPeriod)
	require.Equal(t, math.LegacyOneDec(), val1Stats.Stats.PerformanceRating)

	val2Stats, err := queryServer.ValidatorStats(ctx.WithBlockHeight(100), &revenuetypes.QueryValidatorStatsRequest{
		ConsensusAddress: val2.ConsensusAddress,
	})
	require.Nil(t, err)
	require.Equal(t, uint64(50), val2Stats.Stats.ValidatorInfo.CommitedBlocksInPeriod)
	require.Equal(t, uint64(50), val2Stats.Stats.ValidatorInfo.CommitedOracleVotesInPeriod)
	require.Equal(t, math.LegacyZeroDec(), val2Stats.Stats.PerformanceRating)

	valsStats, err := queryServer.ValidatorsStats(ctx.WithBlockHeight(100), &revenuetypes.QueryValidatorsStatsRequest{})
	require.Nil(t, err)
	require.Equal(t, 2, len(valsStats.Stats))
	require.Equal(t, val1Stats.Stats, valsStats.Stats[1])
	require.Equal(t, val2Stats.Stats, valsStats.Stats[0])
}
