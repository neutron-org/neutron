package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/testutil/apptesting"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/types"
)

var _ = suite.TestingSuite(nil)

type balanceAssertion struct {
	addr     sdk.AccAddress
	balances sdk.Coins
}

func (suite *IncentivesTestSuite) TestValueForShares() {
	addrs := apptesting.SetupAddrs(3)

	tests := []struct {
		name        string
		deposits    []depositSpec
		coin        sdk.Coin
		tick        int64
		expectation math.Int
		err         error
	}{
		// gauge 1 gives 3k coins. three stakes, all eligible. 1k coins per stake.
		// 1k should go to oneStakeUser and 2k to twoStakeUser.
		{
			name: "one deposit",
			deposits: []depositSpec{
				{
					addr:   addrs[0],
					token0: sdk.NewInt64Coin("TokenA", 1000),
					token1: sdk.NewInt64Coin("TokenB", 1000),
					tick:   0,
					fee:    1,
				},
			},
			coin:        sdk.NewInt64Coin(dextypes.NewPoolDenom(0), 20),
			tick:        1000,
			expectation: math.NewInt(21),
		},
		{
			name: "one deposit: no adjustment",
			deposits: []depositSpec{
				{
					addr:   addrs[0],
					token0: sdk.NewInt64Coin("TokenA", 1000),
					token1: sdk.NewInt64Coin("TokenB", 1000),
					tick:   0,
					fee:    1,
				},
			},
			coin:        sdk.NewInt64Coin(dextypes.NewPoolDenom(0), 20),
			tick:        0,
			expectation: math.NewInt(20),
		},
		{
			name: "two deposits: one extraneous",
			deposits: []depositSpec{
				{
					addr:   addrs[0],
					token0: sdk.NewInt64Coin("TokenA", 1000),
					token1: sdk.NewInt64Coin("TokenB", 1000),
					tick:   0,
					fee:    1,
				},
				{
					addr:   addrs[0],
					token0: sdk.NewInt64Coin("TokenA", 1000),
					token1: sdk.NewInt64Coin("TokenB", 1000),
					tick:   0,
					fee:    2,
				},
			},
			coin:        sdk.NewInt64Coin(dextypes.NewPoolDenom(0), 20),
			tick:        1000,
			expectation: math.NewInt(21),
		},
		{
			name: "two deposits: both relevant",
			deposits: []depositSpec{
				{
					addr:   addrs[0],
					token0: sdk.NewInt64Coin("TokenA", 1000),
					token1: sdk.NewInt64Coin("TokenB", 1000),
					tick:   0,
					fee:    1,
				},
				{
					addr:   addrs[0],
					token0: sdk.NewInt64Coin("TokenA", 1000),
					token1: sdk.NewInt64Coin("TokenB", 1000),
					tick:   0,
					fee:    1,
				},
			},
			coin:        sdk.NewInt64Coin(dextypes.NewPoolDenom(0), 20),
			tick:        1000,
			expectation: math.NewInt(21),
		},
	}
	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			suite.SetupTest()
			_ = suite.SetupDeposit(tc.deposits)
			value, err := suite.App.IncentivesKeeper.ValueForShares(suite.Ctx, tc.coin, tc.tick)
			if tc.err == nil {
				require.NoError(t, err)
				require.Equal(t, tc.expectation, value)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestDistribute tests that when the distribute command is executed on a provided gauge
// that the correct amount of rewards is sent to the correct stake owners.
func (suite *IncentivesTestSuite) TestDistribute() {
	addrs := apptesting.SetupAddrs(3)
	tests := []struct {
		name              string
		addrs             []sdk.AccAddress
		depositStakeSpecs []depositStakeSpec
		gaugeSpecs        []gaugeSpec
		assertions        []balanceAssertion
	}{
		{
			name: "one gauge",
			depositStakeSpecs: []depositStakeSpec{
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[0],
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
							tick:   0,
							fee:    1,
						},
					},
					stakeDistEpochOffset: -2,
				},
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[1],
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
							tick:   0,
							fee:    1,
						},
					},
					stakeDistEpochOffset: -2,
				},
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[1],
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
							tick:   0,
							fee:    1,
						},
					},
					stakeDistEpochOffset: -2,
				},
			},
			gaugeSpecs: []gaugeSpec{
				{
					isPerpetual: false,
					rewards:     sdk.Coins{sdk.NewInt64Coin("reward", 3000)},
					startTick:   -10,
					endTick:     10,
					paidOver:    1,
					pricingTick: 0,
				},
			},
			assertions: []balanceAssertion{
				{addr: addrs[0], balances: sdk.Coins{sdk.NewInt64Coin("reward", 1000)}},
				{addr: addrs[1], balances: sdk.Coins{sdk.NewInt64Coin("reward", 2000)}},
			},
		},
		{
			name: "two gauges",
			depositStakeSpecs: []depositStakeSpec{
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[0],
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
							tick:   0,
							fee:    1,
						},
					},
					stakeDistEpochOffset: -2,
				},
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[1],
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
							tick:   0,
							fee:    1,
						},
					},
					stakeDistEpochOffset: -2,
				},
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[1],
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
							tick:   0,
							fee:    1,
						},
					},
					stakeDistEpochOffset: -2,
				},
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[0],
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
							tick:   0,
							fee:    1,
						},
					},
					stakeDistEpochOffset: -1,
				},
			},
			gaugeSpecs: []gaugeSpec{
				{
					isPerpetual: false,
					rewards:     sdk.Coins{sdk.NewInt64Coin("reward", 3000)},
					startTick:   -10,
					endTick:     10,
					paidOver:    1,
					pricingTick: 0,
				},
				{
					isPerpetual: false,
					rewards:     sdk.Coins{sdk.NewInt64Coin("reward", 3000)},
					startTick:   -10,
					endTick:     10,
					paidOver:    2,
					pricingTick: 0,
				},
			},
			assertions: []balanceAssertion{
				{addr: addrs[0], balances: sdk.Coins{sdk.NewInt64Coin("reward", 1500)}},
				{addr: addrs[1], balances: sdk.Coins{sdk.NewInt64Coin("reward", 3000)}},
			},
		},
		{
			name: "one stake with adjustment",
			depositStakeSpecs: []depositStakeSpec{
				{
					depositSpecs: []depositSpec{
						{
							addr:   addrs[0],
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
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
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
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
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
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
							token0: sdk.NewInt64Coin("TokenA", 1000),
							token1: sdk.NewInt64Coin("TokenB", 1000),
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
					pricingTick: 0,
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
			gauges := make(types.Gauges, len(tc.gaugeSpecs))
			for i, gaugeSpec := range tc.gaugeSpecs {
				gauge := suite.SetupGauge(gaugeSpec)
				gauges[i] = gauge
			}
			_, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, gauges)
			require.NoError(t, err)
			// check expected rewards against actual rewards received
			for i, assertion := range tc.assertions {
				bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, assertion.addr)
				assert.Equal(
					t,
					assertion.balances.String(),
					bal.String(),
					"test %v, person %d",
					tc.name,
					i,
				)
			}
		})
	}
}
