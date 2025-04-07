package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// SetInactiveLimitOrderTranche set a specific inactiveLimitOrderTranche in the store from its index
func (k Keeper) SetInactiveLimitOrderTranche(ctx sdk.Context, limitOrderTranche *types.LimitOrderTranche) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix))
	b := k.cdc.MustMarshal(limitOrderTranche)
	store.Set(limitOrderTranche.Key.KeyMarshal(), b)
}

// GetInactiveLimitOrderTranche returns a inactiveLimitOrderTranche from its index
func (k Keeper) GetInactiveLimitOrderTranche(
	ctx sdk.Context,
	limitOrderTrancheKey *types.LimitOrderTrancheKey,
) (val *types.LimitOrderTranche, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix))

	b := store.Get(limitOrderTrancheKey.KeyMarshal())
	if b == nil {
		return val, false
	}

	val = &types.LimitOrderTranche{}
	k.cdc.MustUnmarshal(b, val)

	return val, true
}

// RemoveInactiveLimitOrderTranche removes a inactiveLimitOrderTranche from the store
func (k Keeper) RemoveInactiveLimitOrderTranche(
	ctx sdk.Context,
	limitOrderTrancheKey *types.LimitOrderTrancheKey,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix))
	store.Delete(limitOrderTrancheKey.KeyMarshal())
}

// GetAllInactiveLimitOrderTranche returns all inactiveLimitOrderTranche
func (k Keeper) GetAllInactiveLimitOrderTranche(ctx sdk.Context) (list []*types.LimitOrderTranche) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := &types.LimitOrderTranche{}
		k.cdc.MustUnmarshal(iterator.Value(), val)
		list = append(list, val)
	}

	return
}

// UpdateInactiveTranche handles the logic for all updates to InactiveLimitOrderTranches
// It will delete an InactiveTranche if there is no remaining MakerReserves or TakerReserves
func (k Keeper) UpdateInactiveTranche(sdkCtx sdk.Context, tranche *types.LimitOrderTranche) {
	if tranche.HasTokenIn() || tranche.HasTokenOut() {
		// There are still reserves to be removed. Save the tranche as is.
		k.SetInactiveLimitOrderTranche(sdkCtx, tranche)
	} else {
		// There is nothing left to remove, we can safely remove the tranche entirely.
		k.RemoveInactiveLimitOrderTranche(sdkCtx, tranche.Key)
	}
}
