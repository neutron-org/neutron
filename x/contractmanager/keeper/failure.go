package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/contractmanager/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
)

// SetFailure set a specific failure in the store from its index
func (k Keeper) SetFailure(ctx sdk.Context, failure types.Failure) {
	store :=  prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FailureKeyPrefix))
	b := k.cdc.MustMarshal(&failure)
	store.Set(types.FailureKey(
        failure.Index,
    ), b)
}

// GetFailure returns a failure from its index
func (k Keeper) GetFailure(
    ctx sdk.Context,
    index string,
    
) (val types.Failure, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FailureKeyPrefix))

	b := store.Get(types.FailureKey(
        index,
    ))
    if b == nil {
        return val, false
    }

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveFailure removes a failure from the store
func (k Keeper) RemoveFailure(
    ctx sdk.Context,
    index string,
    
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FailureKeyPrefix))
	store.Delete(types.FailureKey(
	    index,
    ))
}

// GetAllFailure returns all failure
func (k Keeper) GetAllFailure(ctx sdk.Context) (list []types.Failure) {
    store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FailureKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Failure
		k.cdc.MustUnmarshal(iterator.Value(), &val)
        list = append(list, val)
	}

    return
}
