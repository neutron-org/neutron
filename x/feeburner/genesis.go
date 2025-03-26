package feeburner

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/feeburner/keeper"
	"github.com/neutron-org/neutron/v6/x/feeburner/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetTotalBurnedNeutronsAmount(ctx, genState.TotalBurnedNeutronsAmount)

	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.TotalBurnedNeutronsAmount = k.GetTotalBurnedNeutronsAmount(ctx)

	return genesis
}
