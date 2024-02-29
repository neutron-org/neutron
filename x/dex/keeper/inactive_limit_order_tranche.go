package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/types"
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
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := &types.LimitOrderTranche{}
		k.cdc.MustUnmarshal(iterator.Value(), val)
		list = append(list, val)
	}

	return
}

func (k Keeper) SaveInactiveTranche(sdkCtx sdk.Context, tranche *types.LimitOrderTranche) {
	if tranche.HasTokenIn() || tranche.HasTokenOut() {
		k.SetInactiveLimitOrderTranche(sdkCtx, tranche)
	} else {
		k.RemoveInactiveLimitOrderTranche(
			sdkCtx,
			tranche.Key,
		)
	}
}
