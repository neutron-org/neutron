package incentives

import (
	"github.com/CosmWasm/wasmd/x/wasm/simulation"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the incentives module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns nil for governance proposals contents.
// Should eventually be deleted in a future update.
func (AppModule) ProposalContents(
	_ module.SimulationState,
) []simtypes.WeightedProposalMsg {
	return nil
}

// RegisterStoreDecoder has an unknown purpose. Should eventually be deleted in a future update.
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {
}

// WeightedOperations returns the all the module's operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	return operations
}
