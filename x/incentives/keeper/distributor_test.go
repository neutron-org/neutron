package keeper_test

import (
	"testing"
	time "time"

	"cosmossdk.io/math"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/testutil"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	. "github.com/neutron-org/neutron/x/incentives/keeper"
	"github.com/neutron-org/neutron/x/incentives/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ DistributorKeeper = MockKeeper{}

type MockKeeper struct {
	stakes types.Stakes
	keeper DistributorKeeper
}

func NewMockKeeper(keeper DistributorKeeper, stakes types.Stakes) MockKeeper {
	return MockKeeper{
		stakes: stakes,
		keeper: keeper,
	}
}

func (k MockKeeper) ValueForShares(_ sdk.Context, coin sdk.Coin, _ int64) (math.Int, error) {
	return coin.Amount.Mul(math.NewInt(2)), nil
}

func (k MockKeeper) GetStakesByQueryCondition(
	_ sdk.Context,
	_ *types.QueryCondition,
) types.Stakes {
	return k.stakes
}

func (k MockKeeper) StakeCoinsPassingQueryCondition(ctx sdk.Context, stake *types.Stake, distrTo types.QueryCondition) sdk.Coins {
	return k.keeper.StakeCoinsPassingQueryCondition(ctx, stake, distrTo)
}

func TestDistributor(t *testing.T) {
	app := testutil.Setup(t)
	ctx := app.BaseApp.NewContext(
		false,
		tmtypes.Header{Height: 1, ChainID: "duality-1", Time: time.Now().UTC()},
	)

	gauge := types.NewGauge(
		1,
		false,
		types.QueryCondition{
			PairID: &dextypes.PairID{
				Token0: "TokenA",
				Token1: "TokenB",
			},
			StartTick: -10,
			EndTick:   10,
		},
		sdk.Coins{sdk.NewCoin("coin1", math.NewInt(100))},
		ctx.BlockTime(),
		10,
		0,
		sdk.Coins{},
		0,
	)
	rewardPool, _ := app.DexKeeper.GetOrInitPool(ctx, &dextypes.PairID{Token0: "TokenA", Token1: "TokenB"}, 5, 1)
	rewardedDenom := rewardPool.GetPoolDenom()
	nonRewardPool, _ := app.DexKeeper.GetOrInitPool(ctx, &dextypes.PairID{Token0: "TokenA", Token1: "TokenB"}, 12, 1)
	nonRewardedDenom := nonRewardPool.GetPoolDenom()
	addr1 := sdk.AccAddress("addr1")
	addr2 := sdk.AccAddress("addr2")
	addr3 := sdk.AccAddress("addr3")
	allStakes := types.Stakes{
		types.NewStake(1, addr1, sdk.Coins{sdk.NewCoin(rewardedDenom, math.NewInt(50))}, ctx.BlockTime(), 0),
		types.NewStake(2, addr2, sdk.Coins{sdk.NewCoin(rewardedDenom, math.NewInt(25))}, ctx.BlockTime(), 0),
		types.NewStake(3, addr2, sdk.Coins{sdk.NewCoin(rewardedDenom, math.NewInt(25))}, ctx.BlockTime(), 0),
		types.NewStake(4, addr3, sdk.Coins{sdk.NewCoin(nonRewardedDenom, math.NewInt(50))}, ctx.BlockTime(), 0),
	}

	distributor := NewDistributor(NewMockKeeper(app.IncentivesKeeper, allStakes))

	testCases := []struct {
		name         string
		timeOffset   time.Duration
		filterStakes types.Stakes
		expected     types.DistributionSpec
		expectedErr  error
	}{
		{
			name:         "Error case: gauge not active",
			timeOffset:   -1 * time.Minute,
			filterStakes: allStakes,
			expected:     nil,
			expectedErr:  types.ErrGaugeNotActive,
		},
		{
			name:         "Successful case: distribute to all stakes",
			timeOffset:   0,
			filterStakes: allStakes,
			expected: types.DistributionSpec{
				addr1.String(): sdk.Coins{sdk.NewCoin("coin1", math.NewInt(5))},
				addr2.String(): sdk.Coins{sdk.NewCoin("coin1", math.NewInt(4))},
			},
			expectedErr: nil,
		},
		{
			name:         "Successful case: distribute to one stake",
			timeOffset:   0,
			filterStakes: types.Stakes{allStakes[0]},
			expected: types.DistributionSpec{
				addr1.String(): sdk.Coins{sdk.NewCoin("coin1", math.NewInt(5))},
			},
			expectedErr: nil,
		},
		{
			name:         "No distribution: empty filterStakes",
			filterStakes: types.Stakes{},
			expected:     types.DistributionSpec{},
			expectedErr:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			distSpec, err := distributor.Distribute(
				ctx.WithBlockTime(ctx.BlockTime().Add(tc.timeOffset)),
				&gauge,
				tc.filterStakes,
			)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expected, distSpec)
		})
	}
}
