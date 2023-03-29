package cron

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/cron/keeper"
	"github.com/neutron-org/neutron/x/cron/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	ctx.Logger().Error(fmt.Sprintf("InitGenesis: SetParams: %+v", genState.Params))
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
