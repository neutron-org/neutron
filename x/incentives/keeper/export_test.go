package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/x/incentives/types"
)

// AddGaugeRefByKey appends the provided gauge ID into an array associated with the provided key.
func (k Keeper) AddRefByKey(ctx sdk.Context, key []byte, gaugeID uint64) error {
	return k.addRefByKey(ctx, key, gaugeID)
}

// DeleteGaugeRefByKey removes the provided gauge ID from an array associated with the provided key.
func (k Keeper) DeleteRefByKey(ctx sdk.Context, key []byte, guageID uint64) error {
	return k.deleteRefByKey(ctx, key, guageID)
}

// GetGaugeRefs returns the gauge IDs specified by the provided key.
func (k Keeper) GetRefs(ctx sdk.Context, key []byte) []uint64 {
	return k.getRefs(ctx, key)
}

// MoveUpcomingGaugeToActiveGauge moves a gauge that has reached it's start time from an upcoming to an active status.
func (k Keeper) MoveUpcomingGaugeToActiveGauge(ctx sdk.Context, gauge *types.Gauge) error {
	return k.moveUpcomingGaugeToActiveGauge(ctx, gauge)
}

// MoveActiveGaugeToFinishedGauge moves a gauge that has completed its distribution from an active to a finished status.
func (k Keeper) MoveActiveGaugeToFinishedGauge(ctx sdk.Context, gauge *types.Gauge) error {
	return k.moveActiveGaugeToFinishedGauge(ctx, gauge)
}

func (k Keeper) GetStakeRefKeys(ctx sdk.Context, stake *types.Stake) ([][]byte, error) {
	return k.getStakeRefKeys(ctx, stake)
}

func RemoveValue(ids []uint64, id uint64) ([]uint64, int) {
	return removeValue(ids, id)
}

func FindIndex(ids []uint64, id uint64) int {
	return findIndex(ids, id)
}
