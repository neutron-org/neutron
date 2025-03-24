package harpoon

import (
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/v6/x/harpoon/keeper"
	"github.com/neutron-org/neutron/v6/x/harpoon/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, state types.GenesisState) {
	for _, item := range state.HookSubscriptions {
		for _, contractAddr := range item.ContractAddresses {
			addr := sdk.MustAccAddressFromBech32(contractAddr)
			if !k.GetWasmKeeper().HasContractInfo(ctx, addr) {
				panic(errors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("contract address not found: %s", contractAddr)))
			}
		}
		k.SetHookSubscription(ctx, item)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.HookSubscriptions = k.GetAllSubscriptions(ctx)

	return genesis
}
