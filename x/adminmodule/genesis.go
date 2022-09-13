package adminmodule

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/adminmodule/keeper"
	"github.com/neutron-org/neutron/x/adminmodule/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, admin := range genState.GetAdmins() {
		k.SetAdmin(ctx, admin)
	}
	k.SetProposalID(ctx, 1) // TODO add StartingProposalId to genesis
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Admins: k.GetAdmins(ctx),
	}
}
