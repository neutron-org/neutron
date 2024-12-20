package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v5/x/tokenfactory/types"
)

// StoreEscrowAddress sets the total set of params.
func (k Keeper) StoreEscrowAddress(ctx sdk.Context, address sdk.AccAddress) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.EscrowAddressKey)

	prefixStore.Set(address.Bytes(), []byte{0})
}

func (k Keeper) isEscrowAddress(ctx sdk.Context, address sdk.AccAddress) bool {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.EscrowAddressKey)
	bz := prefixStore.Get(address.Bytes())

	return len(bz) != 0
}
