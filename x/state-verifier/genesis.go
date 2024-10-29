package stateverifier

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v5/x/state-verifier/keeper"
	"github.com/neutron-org/neutron/v5/x/state-verifier/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	for _, state := range genState.States {
		if err := k.WriteConsensusState(ctx, state.Height, *state.Cs); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	allCs, err := k.GetAllConsensusStates(ctx)
	if err != nil {
		panic(err)
	}

	genesis.States = append(genesis.States, allCs...)

	return genesis
}
