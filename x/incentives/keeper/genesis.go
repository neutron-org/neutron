package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/x/incentives/types"
)

// InitGenesis initializes the incentives module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	if err := k.InitializeAllStakes(ctx, genState.Stakes); err != nil {
		panic(err)
	}
	if err := k.InitializeAllGauges(ctx, genState.Gauges); err != nil {
		panic(err)
	}
	k.SetLastStakeID(ctx, genState.LastStakeId)
	k.SetLastGaugeID(ctx, genState.LastGaugeId)
	for _, accountHistory := range genState.AccountHistories {
		err := k.SetAccountHistory(ctx, accountHistory)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the x/incentives module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:           k.GetParams(ctx),
		Gauges:           k.GetNotFinishedGauges(ctx),
		LastGaugeId:      k.GetLastGaugeID(ctx),
		LastStakeId:      k.GetLastStakeID(ctx),
		Stakes:           k.GetStakes(ctx),
		AccountHistories: k.GetAllAccountHistory(ctx),
	}
}

// InitializeAllStakes takes a set of stakes, and initializes state to be storing
// them all correctly.
func (k Keeper) InitializeAllStakes(ctx sdk.Context, stakes types.Stakes) error {
	for i, stake := range stakes {
		if i%25000 == 0 {
			msg := fmt.Sprintf("Reset %d stake refs, cur stake ID %d", i, stake.ID)
			ctx.Logger().Info(msg)
		}
		err := k.setStake(ctx, stake)
		if err != nil {
			return err
		}

		err = k.addStakeRefs(ctx, stake)
		if err != nil {
			return err
		}
	}

	return nil
}

// InitializeAllGauges takes a set of gauges, and initializes state to be storing
// them all correctly.
func (k Keeper) InitializeAllGauges(ctx sdk.Context, gauges types.Gauges) error {
	for _, gauge := range gauges {
		err := k.setGauge(ctx, gauge)
		if err != nil {
			return err
		}
		err = k.setGaugeRefs(ctx, gauge)
		if err != nil {
			return err
		}
	}
	return nil
}
