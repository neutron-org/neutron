package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type depositStakeSpec struct {
	depositSpecs         []depositSpec
	stakeDistEpochOffset int // used for simulating the time of staking
}

type depositSpec struct {
	addr   sdk.AccAddress
	token0 sdk.Coin
	token1 sdk.Coin
	tick   int64
	fee    uint64
}

type gaugeSpec struct {
	isPerpetual bool
	rewards     sdk.Coins
	paidOver    uint64
	startTick   int64
	endTick     int64
	pricingTick int64
	startTime   time.Time
}

// AddToGauge adds coins to the specified gauge.
func (suite *KeeperTestSuite) AddToGauge(coins sdk.Coins, gaugeID uint64) uint64 {
	addr := sdk.AccAddress([]byte("addrx---------------"))
	suite.FundAcc(addr, coins)
	err := suite.App.IncentivesKeeper.AddToGaugeRewards(suite.Ctx, addr, coins, gaugeID)
	suite.Require().NoError(err)
	return gaugeID
}

func (suite *KeeperTestSuite) SetupDeposit(ss []depositSpec) sdk.Coins {
	shares := sdk.NewCoins()
	for _, s := range ss {
		suite.FundAcc(s.addr, sdk.Coins{s.token0, s.token1})
		_, _, indivShares, err := suite.App.DexKeeper.DepositCore(
			sdk.WrapSDKContext(suite.Ctx),
			dextypes.MustNewPairID(s.token0.Denom, s.token1.Denom),
			s.addr,
			s.addr,
			[]math.Int{s.token0.Amount},
			[]math.Int{s.token1.Amount},
			[]int64{s.tick},
			[]uint64{s.fee},
			[]*dextypes.DepositOptions{{}},
		)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(indivShares)
		shares = shares.Add(indivShares...)
	}
	return shares
}

func (suite *KeeperTestSuite) SetupDepositAndStake(s depositStakeSpec) *types.Stake {
	shares := suite.SetupDeposit(s.depositSpecs)
	return suite.SetupStake(s.depositSpecs[0].addr, shares, s.stakeDistEpochOffset)
}

// StakeTokens stakes tokens for the specified duration
func (suite *KeeperTestSuite) SetupStake(
	addr sdk.AccAddress,
	shares sdk.Coins,
	distEpochOffset int,
) *types.Stake {
	params := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
	epoch := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, params.GetDistrEpochIdentifier())
	stake, err := suite.App.IncentivesKeeper.CreateStake(
		suite.Ctx,
		addr,
		shares,
		suite.Ctx.BlockTime(), // irrelevant now
		epoch.CurrentEpoch+int64(distEpochOffset),
	)
	suite.Require().NoError(err)
	return stake
}

// setupNewGauge creates a gauge with the specified duration.
func (suite *KeeperTestSuite) SetupGauge(s gaugeSpec) *types.Gauge {
	addr := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))

	// fund reward tokens
	suite.FundAcc(addr, s.rewards)

	// create gauge
	gauge, err := suite.App.IncentivesKeeper.CreateGauge(
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
	suite.Require().NoError(err)
	return gauge
}

func (suite *KeeperTestSuite) SetupGauges(specs []gaugeSpec) {
	for _, s := range specs {
		suite.SetupGauge(s)
	}
}
