package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/utils/math"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/types"
)

var _ DistributorKeeper = Keeper{}

func (k Keeper) ValueForShares(ctx sdk.Context, coin sdk.Coin, tick int64) (math.Int, error) {
	totalShares := k.bk.GetSupply(ctx, coin.Denom).Amount
	poolMetadata, err := k.dk.GetPoolMetadataByDenom(ctx, coin.Denom)
	if err != nil {
		return math.ZeroInt(), err
	}

	pool, err := k.dk.GetOrInitPool(
		ctx,
		poolMetadata.PairID,
		poolMetadata.Tick,
		poolMetadata.Fee,
	)
	if err != nil {
		return math.ZeroInt(), err
	}
	amount0, amount1 := pool.RedeemValue(coin.Amount, totalShares)
	price1To0Center, err := dextypes.CalcPrice(-1 * tick)
	if err != nil {
		return math.ZeroInt(), err
	}
	return math_utils.NewPrecDecFromInt(amount0).Add(price1To0Center.MulInt(amount1)).TruncateInt(), nil
}

// Distribute distributes coins from an array of gauges to all eligible stakes.
func (k Keeper) Distribute(ctx sdk.Context, gauges types.Gauges) (types.DistributionSpec, error) {
	distSpec := types.DistributionSpec{}
	for _, gauge := range gauges {
		gaugeDistSpec, err := k.distributor.Distribute(ctx, gauge, nil)
		if err != nil {
			return nil, err
		}
		distSpec = distSpec.Add(gaugeDistSpec)

		err = k.setGauge(ctx, gauge)
		if err != nil {
			return nil, err
		}
		if gauge.IsFinishedGauge(ctx.BlockTime()) {
			if err := k.moveActiveGaugeToFinishedGauge(ctx, gauge); err != nil {
				return nil, err
			}
		}
	}

	ctx.Logger().Debug(fmt.Sprintf("Beginning distribution to %d users", len(distSpec)))
	for addr, rewards := range distSpec {
		decodedAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, err
		}
		err = k.bk.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			decodedAddr,
			rewards)
		if err != nil {
			return nil, err
		}

		// Accumulate to account history
		accHistory, found := k.GetAccountHistory(ctx, addr)
		if found {
			accHistory.Coins = accHistory.Coins.Add(rewards...)
		} else {
			accHistory = NewAccountHistory(addr, rewards)
		}
		if err := k.SetAccountHistory(ctx, accHistory); err != nil {
			return nil, err
		}

		// Emit events
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.TypeEvtDistribution,
				sdk.NewAttribute(types.AttributeReceiver, addr),
				sdk.NewAttribute(types.AttributeAmount, rewards.String()),
			),
		})
	}
	ctx.Logger().Debug(fmt.Sprintf("Finished Distributing to %d users", len(distSpec)))
	k.hooks.AfterEpochDistribution(ctx)
	return distSpec, nil
}

// GetModuleCoinsToBeDistributed returns sum of coins yet to be distributed for all of the module.
func (k Keeper) GetModuleCoinsToBeDistributed(ctx sdk.Context) sdk.Coins {
	activeGaugesDistr := k.GetActiveGauges(ctx).GetCoinsRemaining()
	upcomingGaugesDistr := k.GetUpcomingGauges(ctx).GetCoinsRemaining()
	return activeGaugesDistr.Add(upcomingGaugesDistr...)
}

// GetModuleDistributedCoins returns sum of coins that have been distributed so far for all of the module.
func (k Keeper) GetModuleDistributedCoins(ctx sdk.Context) sdk.Coins {
	activeGaugesDistr := k.GetActiveGauges(ctx).GetCoinsDistributed()
	finishedGaugesDistr := k.GetFinishedGauges(ctx).GetCoinsDistributed()
	return activeGaugesDistr.Add(finishedGaugesDistr...)
}

// GetRewardsEstimate returns rewards estimation at a future specific time (by epoch)
// If stakes are nil, it returns the rewards between now and the end epoch associated with address.
// If stakes are not nil, it returns all the rewards for the given stakes between now and end epoch.
func (k Keeper) GetRewardsEstimate(
	ctx sdk.Context,
	addr sdk.AccAddress,
	filterStakes types.Stakes,
	numEpochs int64,
) (sdk.Coins, error) {
	// if stakes are nil, populate with all stakes associated with the address
	if len(filterStakes) == 0 {
		filterStakes = k.GetStakesByAccount(ctx, addr)
	}

	// for each specified stake get associated pairs
	pairSet := map[dextypes.PairID]bool{}
	for _, l := range filterStakes {
		for _, c := range l.Coins {
			poolMetadata, err := k.dk.GetPoolMetadataByDenom(ctx, c.Denom)
			if err != nil {
				panic("all stakes should be valid deposit denoms")
			}
			pairSet[*poolMetadata.PairID] = true
		}
	}

	// for each pair get associated gauges
	gauges := types.Gauges{}
	for s := range pairSet {
		gauges = append(gauges, k.GetGaugesByPair(ctx, &s)...)
	}

	// estimate rewards
	estimatedRewards := sdk.Coins{}
	epochInfo := k.GetEpochInfo(ctx)

	// ensure we don't change storage while doing estimation
	cacheCtx, _ := ctx.CacheContext()
	for _, gauge := range gauges {
		distrBeginEpoch := epochInfo.CurrentEpoch
		endEpoch := epochInfo.CurrentEpoch + numEpochs
		bstakeTime := ctx.BlockTime()
		if gauge.StartTime.After(bstakeTime) {
			distrBeginEpoch = epochInfo.CurrentEpoch + 1 + int64(
				gauge.StartTime.Sub(bstakeTime)/epochInfo.Duration,
			)
		}

		// TODO: Make more efficient by making it possible to call distribute with this
		// gaugeStakes := k.GetStakesByQueryCondition(cacheCtx, &gauge.DistributeTo)
		gaugeRewards := sdk.Coins{}
		for epoch := distrBeginEpoch; epoch <= endEpoch; epoch++ {
			epochTime := epochInfo.StartTime.Add(
				time.Duration(epoch-epochInfo.CurrentEpoch) * epochInfo.Duration,
			)
			if !gauge.IsActiveGauge(epochTime) {
				break
			}

			futureCtx := cacheCtx.WithBlockTime(epochTime)
			distSpec, err := k.distributor.Distribute(futureCtx, gauge, filterStakes)
			if err != nil {
				return nil, err
			}

			gaugeRewards = gaugeRewards.Add(distSpec.GetTotal()...)
		}
		estimatedRewards = estimatedRewards.Add(gaugeRewards...)
	}

	return estimatedRewards, nil
}
