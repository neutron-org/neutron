package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	db "github.com/cometbft/cometbft-db"
	"github.com/cosmos/gogoproto/proto"

	dextypes "github.com/neutron-org/neutron/x/dex/types"
	epochtypes "github.com/neutron-org/neutron/x/epochs/types"
	"github.com/neutron-org/neutron/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetLastGaugeID returns the last used gauge ID.
func (k Keeper) GetLastGaugeID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastGaugeID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastGaugeID sets the last used gauge ID to the provided ID.
func (k Keeper) SetLastGaugeID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastGaugeID, sdk.Uint64ToBigEndian(id))
}

// getGaugesFromIterator iterates over everything in a gauge's iterator, until it reaches the end. Return all gauges iterated over.
func (k Keeper) getGaugesFromIterator(ctx sdk.Context, iterator db.Iterator) types.Gauges {
	gauges := []*types.Gauge{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		gaugeIDs := []uint64{}
		err := json.Unmarshal(iterator.Value(), &gaugeIDs)
		if err != nil {
			panic(err)
		}
		for _, gaugeID := range gaugeIDs {
			gauge, err := k.GetGaugeByID(ctx, gaugeID)
			if err != nil {
				panic(err)
			}
			gauges = append(gauges, gauge)
		}
	}
	return gauges
}

func (k Keeper) setGaugeRefs(ctx sdk.Context, gauge *types.Gauge) error {
	switch {
	case gauge.IsUpcomingGauge(ctx.BlockTime()):
		if err := k.addRefByKey(ctx, types.CombineKeys(types.KeyPrefixGaugeIndexUpcoming, types.GetTimeKey(gauge.StartTime)), gauge.Id); err != nil {
			return err
		}
		err := k.addRefByKey(
			ctx,
			types.GetKeyGaugeIndexByPair(gauge.DistributeTo.PairID.CanonicalString()),
			gauge.Id,
		)
		if err != nil {
			return err
		}
	case gauge.IsActiveGauge(ctx.BlockTime()):
		err := k.addRefByKey(ctx, types.CombineKeys(types.KeyPrefixGaugeIndexActive, types.GetTimeKey(gauge.StartTime)), gauge.Id)
		if err != nil {
			return err
		}
		err = k.addRefByKey(ctx, types.GetKeyGaugeIndexByPair(gauge.DistributeTo.PairID.CanonicalString()), gauge.Id)
		if err != nil {
			return err
		}
	default: // finished gauge
		err := k.addRefByKey(ctx, types.CombineKeys(types.KeyPrefixGaugeIndexFinished, types.GetTimeKey(gauge.StartTime)), gauge.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

// setGauge set the gauge inside store.
func (k Keeper) setGauge(ctx sdk.Context, gauge *types.Gauge) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(gauge)
	if err != nil {
		return err
	}
	store.Set(types.GetKeyGaugeStore(gauge.Id), bz)
	return nil
}

// CreateGauge creates a gauge and sends coins to the gauge.
func (k Keeper) CreateGauge(
	ctx sdk.Context,
	isPerpetual bool,
	owner sdk.AccAddress,
	coins sdk.Coins,
	distrTo types.QueryCondition,
	startTime time.Time,
	numEpochsPaidOver uint64,
	pricingTick int64,
) (*types.Gauge, error) {
	numGauges := k.GetLastGaugeID(ctx)
	params := k.GetParams(ctx)
	if params.MaxGauges < numGauges+1 {
		return nil, types.ErrMaxGaugesReached
	}

	// Perhaps overly defensive checks, these validations are also being performed
	// in the ValidateBasic() for CreateGaugeMsg.
	if dextypes.IsTickOutOfRange(pricingTick) {
		return nil, types.ErrGaugePricingTickOutOfRange
	}

	if dextypes.IsTickOutOfRange(distrTo.StartTick) {
		return nil, types.ErrGaugeDistrToTickOutOfRange
	}

	if dextypes.IsTickOutOfRange(distrTo.EndTick) {
		return nil, types.ErrGaugeDistrToTickOutOfRange
	}

	gauge := &types.Gauge{
		Id:                numGauges + 1,
		IsPerpetual:       isPerpetual,
		DistributeTo:      distrTo,
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
		// If this is outside the tick range then the distribution step will fail
		PricingTick: pricingTick,
	}

	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, gauge.Coins); err != nil {
		return nil, err
	}

	err := k.setGauge(ctx, gauge)
	if err != nil {
		return nil, err
	}

	k.SetLastGaugeID(ctx, gauge.Id)

	err = k.setGaugeRefs(ctx, gauge)
	if err != nil {
		return nil, err
	}

	k.hooks.AfterCreateGauge(ctx, gauge.Id)
	return gauge, nil
}

// AddToGaugeRewards adds coins to gauge.
func (k Keeper) AddToGaugeRewards(
	ctx sdk.Context,
	owner sdk.AccAddress,
	coins sdk.Coins,
	gaugeID uint64,
) error {
	gauge, err := k.GetGaugeByID(ctx, gaugeID)
	if err != nil {
		return err
	}
	if gauge.IsFinishedGauge(ctx.BlockTime()) {
		return errors.New("gauge is already completed")
	}
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return err
	}

	gauge.Coins = gauge.Coins.Add(coins...)
	err = k.setGauge(ctx, gauge)
	if err != nil {
		return err
	}
	k.hooks.AfterAddToGauge(ctx, gauge.Id)
	return nil
}

// GetGaugeByID returns gauge from gauge ID.
func (k Keeper) GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*types.Gauge, error) {
	gauge := types.Gauge{}
	store := ctx.KVStore(k.storeKey)
	gaugeKey := types.GetKeyGaugeStore(gaugeID)
	if !store.Has(gaugeKey) {
		return nil, fmt.Errorf("gauge with ID %d does not exist", gaugeID)
	}
	bz := store.Get(gaugeKey)
	if err := proto.Unmarshal(bz, &gauge); err != nil {
		return nil, err
	}
	return &gauge, nil
}

// GetGaugeQualifyingValue returns gauge qualifying value from gauge ID.
func (k Keeper) GetGaugeQualifyingValue(ctx sdk.Context, gaugeID uint64) (uint64, error) {
	gauge := types.Gauge{}
	store := ctx.KVStore(k.storeKey)
	gaugeKey := types.GetKeyGaugeStore(gaugeID)
	if !store.Has(gaugeKey) {
		return 0, fmt.Errorf("gauge with ID %d does not exist", gaugeID)
	}
	bz := store.Get(gaugeKey)
	if err := proto.Unmarshal(bz, &gauge); err != nil {
		return 0, err
	}
	var value uint64
	stakes := k.GetStakesByQueryCondition(ctx, &gauge.DistributeTo)
	for _, stake := range stakes {
		stakeCoins := k.StakeCoinsPassingQueryCondition(ctx, stake, gauge.DistributeTo)
		for _, stakeCoin := range stakeCoins {
			adjustedPositionValue, err := k.ValueForShares(ctx, stakeCoin, gauge.PricingTick)
			if err != nil {
				return 0, err
			}
			value += value + adjustedPositionValue.Uint64()
		}
	}
	return value, nil
}

// GetGauges returns upcoming, active, and finished gauges.
func (k Keeper) GetGauges(ctx sdk.Context) types.Gauges {
	return k.getGaugesFromIterator(ctx, k.iterator(ctx, types.KeyPrefixGaugeIndex))
}

// GetNotFinishedGauges returns both upcoming and active gauges.
func (k Keeper) GetNotFinishedGauges(ctx sdk.Context) types.Gauges {
	return append(k.GetActiveGauges(ctx), k.GetUpcomingGauges(ctx)...)
}

// GetEpochInfo returns EpochInfo struct given context.
func (k Keeper) GetEpochInfo(ctx sdk.Context) epochtypes.EpochInfo {
	params := k.GetParams(ctx)
	return k.ek.GetEpochInfo(ctx, params.DistrEpochIdentifier)
}

// moveUpcomingGaugeToActiveGauge moves a gauge that has reached it's start time from an upcoming to an active status.
func (k Keeper) moveUpcomingGaugeToActiveGauge(ctx sdk.Context, gauge *types.Gauge) error {
	// validation for current time and distribution start time
	if ctx.BlockTime().Before(gauge.StartTime) {
		return fmt.Errorf(
			"gauge is not able to start distribution yet: %s >= %s",
			ctx.BlockTime().String(),
			gauge.StartTime.String(),
		)
	}

	timeKey := types.GetTimeKey(gauge.StartTime)
	if err := k.deleteRefByKey(ctx, types.CombineKeys(types.KeyPrefixGaugeIndexUpcoming, timeKey), gauge.Id); err != nil {
		return err
	}

	err := k.addRefByKey(ctx, types.CombineKeys(types.KeyPrefixGaugeIndexActive, timeKey), gauge.Id)
	return err
}

// moveActiveGaugeToFinishedGauge moves a gauge that has completed its distribution from an active to a finished status.
func (k Keeper) moveActiveGaugeToFinishedGauge(ctx sdk.Context, gauge *types.Gauge) error {
	timeKey := types.GetTimeKey(gauge.StartTime)
	if err := k.deleteRefByKey(ctx, types.CombineKeys(types.KeyPrefixGaugeIndexActive, timeKey), gauge.Id); err != nil {
		return err
	}
	if err := k.addRefByKey(ctx, types.CombineKeys(types.KeyPrefixGaugeIndexFinished, timeKey), gauge.Id); err != nil {
		return err
	}
	err := k.deleteRefByKey(
		ctx,
		types.GetKeyGaugeIndexByPair(gauge.DistributeTo.PairID.CanonicalString()),
		gauge.Id,
	)
	if err != nil {
		return err
	}
	k.hooks.AfterFinishDistribution(ctx, gauge.Id)
	return nil
}

// GetActiveGauges returns active gauges.
func (k Keeper) GetActiveGauges(ctx sdk.Context) types.Gauges {
	return k.getGaugesFromIterator(ctx, k.iterator(ctx, types.KeyPrefixGaugeIndexActive))
}

// GetUpcomingGauges returns upcoming gauges.
func (k Keeper) GetUpcomingGauges(ctx sdk.Context) types.Gauges {
	return k.getGaugesFromIterator(ctx, k.iterator(ctx, types.KeyPrefixGaugeIndexUpcoming))
}

// GetFinishedGauges returns finished gauges.
func (k Keeper) GetFinishedGauges(ctx sdk.Context) types.Gauges {
	return k.getGaugesFromIterator(ctx, k.iterator(ctx, types.KeyPrefixGaugeIndexFinished))
}

func (k Keeper) GetGaugesByPair(ctx sdk.Context, pair *dextypes.PairID) []*types.Gauge {
	return k.getGaugesFromIterator(
		ctx,
		k.iterator(ctx, types.GetKeyGaugeIndexByPair(pair.CanonicalString())),
	)
}
