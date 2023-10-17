package keeper_test

import (
	"testing"
	"time"

	"github.com/neutron-org/neutron/testutil/apptesting"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

func (suite *IncentivesTestSuite) TestGaugeLifecycle() {
	addr0 := suite.SetupAddr(0)

	// setup dex deposit and stake of those shares
	suite.SetupDepositAndStake(depositStakeSpec{
		depositSpecs: []depositSpec{
			{
				addr:   addr0,
				token0: sdk.NewInt64Coin("TokenA", 10),
				token1: sdk.NewInt64Coin("TokenB", 10),
				tick:   0,
				fee:    1,
			},
		},
		stakeDistEpochOffset: -2,
	})

	// setup gauge starting 24 hours in the future
	suite.SetupGauge(gaugeSpec{
		startTime:   suite.Ctx.BlockTime().Add(24 * time.Hour),
		isPerpetual: false,
		rewards:     sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)),
		paidOver:    2,
		startTick:   -10,
		endTick:     10,
		pricingTick: 0,
	})

	// assert that the gauge is not in effect yet by triggering an epoch end before gauge start
	err := suite.App.IncentivesKeeper.AfterEpochEnd(suite.Ctx, "day")
	require.NoError(suite.T(), err)
	// no distribution yet
	require.Equal(
		suite.T(),
		"0foocoin",
		suite.App.BankKeeper.GetBalance(suite.Ctx, addr0, "foocoin").String(),
	)
	// assert that gauge state is well-managed
	require.Equal(suite.T(), len(suite.QueryServer.GetUpcomingGauges(suite.Ctx)), 1)
	require.Equal(suite.T(), len(suite.QueryServer.GetActiveGauges(suite.Ctx)), 0)
	require.Equal(suite.T(), len(suite.QueryServer.GetFinishedGauges(suite.Ctx)), 0)

	// advance time to epoch at or after the gauge starts, triggering distribution
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(24 * time.Hour))
	err = suite.App.IncentivesKeeper.AfterEpochEnd(suite.Ctx, "day")
	require.NoError(suite.T(), err)

	// assert that the gauge distributed
	require.Equal(
		suite.T(),
		"5foocoin",
		suite.App.BankKeeper.GetBalance(suite.Ctx, addr0, "foocoin").String(),
	)
	// assert that gauge state is well-managed
	require.Equal(suite.T(), len(suite.QueryServer.GetUpcomingGauges(suite.Ctx)), 0)
	require.Equal(suite.T(), len(suite.QueryServer.GetActiveGauges(suite.Ctx)), 1)
	require.Equal(suite.T(), len(suite.QueryServer.GetFinishedGauges(suite.Ctx)), 0)

	// advance to next epoch
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(24 * time.Hour))
	err = suite.App.IncentivesKeeper.AfterEpochEnd(suite.Ctx, "day")
	require.NoError(suite.T(), err)

	// assert new distribution
	require.Equal(
		suite.T(),
		"10foocoin",
		suite.App.BankKeeper.GetBalance(suite.Ctx, addr0, "foocoin").String(),
	)
	// assert that gauge state is well-managed
	require.Equal(suite.T(), len(suite.QueryServer.GetUpcomingGauges(suite.Ctx)), 0)
	require.Equal(suite.T(), len(suite.QueryServer.GetActiveGauges(suite.Ctx)), 0)
	require.Equal(suite.T(), len(suite.QueryServer.GetFinishedGauges(suite.Ctx)), 1)

	// repeat advancing to next epoch until gauge should be finished
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(24 * time.Hour))
	err = suite.App.IncentivesKeeper.AfterEpochEnd(suite.Ctx, "day")
	require.NoError(suite.T(), err)

	// assert no additional distribution from finished gauge
	require.Equal(
		suite.T(),
		"10foocoin",
		suite.App.BankKeeper.GetBalance(suite.Ctx, addr0, "foocoin").String(),
	)
	// assert that gauge state is well-managed
	require.Equal(suite.T(), len(suite.QueryServer.GetUpcomingGauges(suite.Ctx)), 0)
	require.Equal(suite.T(), len(suite.QueryServer.GetActiveGauges(suite.Ctx)), 0)
	require.Equal(suite.T(), len(suite.QueryServer.GetFinishedGauges(suite.Ctx)), 1)
	// fin.
}

func (suite *IncentivesTestSuite) TestGaugeLimit() {
	// We set the gauge limit to 20. On the 21st gauge, we should encounter an error.
	params := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
	params.MaxGauges = 20
	err := suite.App.IncentivesKeeper.SetParams(suite.Ctx, params)
	suite.Require().NoError(err)

	addr0 := suite.SetupAddr(0)

	// setup dex deposit and stake of those shares
	suite.SetupDepositAndStake(depositStakeSpec{
		depositSpecs: []depositSpec{
			{
				addr:   addr0,
				token0: sdk.NewInt64Coin("TokenA", 10),
				token1: sdk.NewInt64Coin("TokenB", 10),
				tick:   0,
				fee:    1,
			},
		},
		stakeDistEpochOffset: -2,
	})

	for i := 0; i < 20; i++ {
		// setup gauge starting 24 hours in the future
		suite.SetupGauge(gaugeSpec{
			startTime:   suite.Ctx.BlockTime().Add(24 * time.Hour),
			isPerpetual: false,
			rewards:     sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)),
			paidOver:    2,
			startTick:   -10,
			endTick:     10,
			pricingTick: 0,
		})
	}

	addr := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))

	// fund reward tokens
	suite.FundAcc(addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))

	// create gauge
	_, err = suite.App.IncentivesKeeper.CreateGauge(
		suite.Ctx,
		false,
		addr,
		sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)),
		types.QueryCondition{
			PairID: &dextypes.PairID{
				Token0: "TokenA",
				Token1: "TokenB",
			},
			StartTick: -10,
			EndTick:   10,
		},
		suite.Ctx.BlockTime().Add(24*time.Hour),
		2,
		0,
	)
	suite.Require().Error(err)
}

// TestGaugeCreateFails tests that when the distribute command is executed on a provided bad gauge
// that the step fails gracefully.
func (suite *IncentivesTestSuite) TestGaugeCreateFails() {
	addrs := apptesting.SetupAddrs(3)
	tests := []struct {
		name              string
		addrs             []sdk.AccAddress
		depositStakeSpecs []depositStakeSpec
		gaugeSpecs        []gaugeSpec
		assertions        []balanceAssertion
	}{
		{
			name: "one stake with bad gauge",
			depositStakeSpecs: []depositStakeSpec{
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[0],
							token0: sdk.NewInt64Coin("TokenA", 10),
							token1: sdk.NewInt64Coin("TokenB", 10),
							tick:   999,
							fee:    1,
						},
					},
					stakeDistEpochOffset: -2,
				},
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[1],
							token0: sdk.NewInt64Coin("TokenA", 10),
							token1: sdk.NewInt64Coin("TokenB", 10),
							tick:   999,
							fee:    1,
						},
					},
					stakeDistEpochOffset: -2,
				},
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[1],
							token0: sdk.NewInt64Coin("TokenA", 10),
							token1: sdk.NewInt64Coin("TokenB", 10),
							tick:   999,
							fee:    50,
						},
					},
					stakeDistEpochOffset: -2,
				},
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[1],
							token0: sdk.NewInt64Coin("TokenA", 10),
							token1: sdk.NewInt64Coin("TokenB", 10),
							tick:   999,
							fee:    50,
						},
					},
					stakeDistEpochOffset: -1,
				},
			},
			gaugeSpecs: []gaugeSpec{
				{
					isPerpetual: false,
					rewards:     sdk.Coins{sdk.NewInt64Coin("reward", 3000)},
					startTick:   -1000,
					endTick:     1000,
					paidOver:    1,
					pricingTick: 9999999,
				},
			},
			assertions: []balanceAssertion{
				{addr: addrs[0], balances: sdk.Coins{sdk.NewInt64Coin("reward", 1500)}},
				{addr: addrs[1], balances: sdk.Coins{sdk.NewInt64Coin("reward", 1500)}},
			},
		},
	}
	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			suite.SetupTest()
			for _, depositSpec := range tc.depositStakeSpecs {
				suite.SetupDepositAndStake(depositSpec)
			}
			for _, s := range tc.gaugeSpecs {
				addr := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))

				// fund reward tokens
				suite.FundAcc(addr, s.rewards)

				// create gauge
				_, err := suite.App.IncentivesKeeper.CreateGauge(
					suite.Ctx,
					s.isPerpetual,
					addr,
					s.rewards,
					types.QueryCondition{
						PairID: &dextypes.PairID{
							Token0: "TokenA",
							Token1: "TokenB",
						},
						StartTick: s.startTick,
						EndTick:   s.endTick,
					},
					s.startTime,
					s.paidOver,
					s.pricingTick,
				)
				require.Error(t, err)
			}
		})
	}
}
