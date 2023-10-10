package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/incentives/types"
)

type DistributorKeeper interface {
	ValueForShares(ctx sdk.Context, coin sdk.Coin, tick int64) (math.Int, error)
	GetStakesByQueryCondition(ctx sdk.Context, distrTo *types.QueryCondition) types.Stakes
	StakeCoinsPassingQueryCondition(ctx sdk.Context, stake *types.Stake, distrTo types.QueryCondition) sdk.Coins
}

type Distributor struct {
	keeper DistributorKeeper
}

func NewDistributor(keeper DistributorKeeper) Distributor {
	return Distributor{
		keeper: keeper,
	}
}

func (d Distributor) Distribute(
	ctx sdk.Context,
	gauge *types.Gauge,
	filterStakes types.Stakes,
) (types.DistributionSpec, error) {
	if !gauge.IsActiveGauge(ctx.BlockTime()) {
		return nil, types.ErrGaugeNotActive
	}

	distSpec := types.DistributionSpec{}

	rewardsNextEpoch := gauge.RewardsNextEpoch()

	adjustedGaugeTotal := math.ZeroInt()

	gaugeStakes := d.keeper.GetStakesByQueryCondition(ctx, &gauge.DistributeTo)
	if filterStakes == nil {
		filterStakes = gaugeStakes
	}

	stakeSumCache := make(map[uint64]math.Int, len(gaugeStakes))
	for _, stake := range gaugeStakes {
		stakeCoins := d.keeper.StakeCoinsPassingQueryCondition(ctx, stake, gauge.DistributeTo)
		stakeTotal := math.ZeroInt()
		for _, stakeCoin := range stakeCoins {
			adjustedPositionValue, err := d.keeper.ValueForShares(ctx, stakeCoin, gauge.PricingTick)
			if err != nil {
				return nil, err
			}
			stakeTotal = stakeTotal.Add(adjustedPositionValue)
		}
		adjustedGaugeTotal = adjustedGaugeTotal.Add(stakeTotal)
		stakeSumCache[stake.ID] = stakeTotal
	}

	if adjustedGaugeTotal.IsZero() {
		return distSpec, nil
	}

	for _, stake := range filterStakes {
		stakeAmt := stakeSumCache[stake.ID]
		distCoins := sdk.Coins{}
		for _, epochRewards := range rewardsNextEpoch {
			// distribution amount = gauge_size * denom_stake_amount / (total_denom_stake_amount * remain_epochs)
			amount := sdk.NewDecFromInt(epochRewards.Amount).
				Mul(sdk.NewDecFromInt(stakeAmt)).
				Quo(sdk.NewDecFromInt(adjustedGaugeTotal)).
				TruncateInt()
			reward := sdk.Coin{Denom: epochRewards.Denom, Amount: amount}
			distCoins = distCoins.Add(reward)
		}

		// update the amount for that address
		if distCoins.Empty() {
			continue
		}

		if spec, ok := distSpec[stake.Owner]; ok {
			distSpec[stake.Owner] = spec.Add(distCoins...)
		} else {
			distSpec[stake.Owner] = distCoins
		}
	}

	gauge.DistributedCoins = gauge.DistributedCoins.Add(rewardsNextEpoch...)
	gauge.FilledEpochs++
	return distSpec, nil
}
