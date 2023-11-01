package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/incentives/types"
	"github.com/stretchr/testify/suite"
)

var _ = suite.TestingSuite(nil)

func (suite *IncentivesTestSuite) TestGetFutureRewardEstimate() {
	addr1 := suite.SetupAddr(0)
	suite.SetupDepositAndStake(depositStakeSpec{
		depositSpecs: []depositSpec{
			{
				addr:   addr1,
				token0: sdk.NewInt64Coin("TokenA", 1000),
				token1: sdk.NewInt64Coin("TokenB", 1000),
				tick:   0,
				fee:    1,
			},
		},
		stakeDistEpochOffset: -2,
	})
	addr2 := suite.SetupAddr(1)
	suite.SetupDepositAndStake(depositStakeSpec{
		depositSpecs: []depositSpec{
			{
				addr:   addr2,
				token0: sdk.NewInt64Coin("TokenA", 1000),
				token1: sdk.NewInt64Coin("TokenB", 1000),
				tick:   0,
				fee:    1,
			},
		},
		stakeDistEpochOffset: -2,
	})
	suite.SetupGauge(gaugeSpec{
		isPerpetual: false,
		rewards:     sdk.NewCoins(sdk.NewInt64Coin("foocoin", 1000)),
		paidOver:    100,
		startTick:   -10,
		endTick:     10,
		pricingTick: 0,
		startTime:   suite.Ctx.BlockTime(),
	})
	suite.SetupGauge(gaugeSpec{
		isPerpetual: false,
		rewards:     sdk.NewCoins(sdk.NewInt64Coin("foocoin", 1000)),
		paidOver:    100,
		startTick:   -10,
		endTick:     10,
		pricingTick: 0,
		startTime:   suite.Ctx.BlockTime().Add(315 * 24 * time.Hour),
	})
	estimate, err := suite.QueryServer.GetFutureRewardEstimate(
		suite.GoCtx,
		&types.GetFutureRewardEstimateRequest{
			Owner:     addr1.String(),
			StakeIds:  nil,
			NumEpochs: 365,
		},
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 750)), estimate.Coins)
}

func (suite *IncentivesTestSuite) TestGetGauges() {
	addr1 := suite.SetupAddr(0)
	suite.SetupDepositAndStake(depositStakeSpec{
		depositSpecs: []depositSpec{
			{
				addr:   addr1,
				token0: sdk.NewInt64Coin("TokenA", 1000),
				token1: sdk.NewInt64Coin("TokenB", 1000),
				tick:   0,
				fee:    1,
			},
		},
		stakeDistEpochOffset: -2,
	})
	addr2 := suite.SetupAddr(1)
	suite.SetupDepositAndStake(depositStakeSpec{
		depositSpecs: []depositSpec{
			{
				addr:   addr2,
				token0: sdk.NewInt64Coin("TokenA", 1000),
				token1: sdk.NewInt64Coin("TokenB", 1000),
				tick:   0,
				fee:    1,
			},
		},
		stakeDistEpochOffset: -2,
	})
	gauge1 := suite.SetupGauge(gaugeSpec{
		isPerpetual: false,
		rewards:     sdk.NewCoins(sdk.NewInt64Coin("foocoin", 1000)),
		paidOver:    100,
		startTick:   -10,
		endTick:     10,
		pricingTick: 0,
		startTime:   suite.Ctx.BlockTime(),
	})
	gauge2 := suite.SetupGauge(gaugeSpec{
		isPerpetual: false,
		rewards:     sdk.NewCoins(sdk.NewInt64Coin("foocoin", 1000)),
		paidOver:    100,
		startTick:   -10,
		endTick:     10,
		pricingTick: 0,
		startTime:   suite.Ctx.BlockTime().Add(315 * 24 * time.Hour),
	})

	response, err := suite.QueryServer.GetGauges(suite.GoCtx, &types.GetGaugesRequest{
		Status: types.GaugeStatus_ACTIVE_UPCOMING,
		Denom:  "",
	})

	suite.Require().NoError(err)
	suite.Require().Equal([]*types.Gauge{
		gauge2,
		gauge1,
	}, response.Gauges)
}
