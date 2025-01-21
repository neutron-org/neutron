package harpoon

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v5/x/harpoon/keeper"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(_ sdk.Context, _ keeper.Keeper, _ types.GenesisState) {}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(_ sdk.Context, _ keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	return genesis
}
