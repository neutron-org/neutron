package keeper

import (
	epochstypes "github.com/neutron-org/neutron/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeforeEpochStart is the epoch start hook.
func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string) error {
	params := k.GetParams(ctx)
	if epochIdentifier == params.DistrEpochIdentifier {
		// begin distribution if it's start time
		gauges := k.GetUpcomingGauges(ctx)
		for _, gauge := range gauges {
			if gauge.IsActiveGauge(ctx.BlockTime()) {
				if err := k.moveUpcomingGaugeToActiveGauge(ctx, gauge); err != nil {
					return err
				}
			}
		}

		// distribute due to epoch event
		gauges = k.GetActiveGauges(ctx)
		_, err := k.Distribute(ctx, gauges)
		if err != nil {
			return err
		}
	}
	return nil
}

// ___________________________________________________________________________________________________

// Hooks is the wrapper struct for the incentives keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Hooks returns the hook wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// BeforeEpochStart is the epoch start hook.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier)
}

// AfterEpochEnd is the epoch end hook.
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier)
}
