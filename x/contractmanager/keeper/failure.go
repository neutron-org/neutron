package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

// AddContractFailure adds a specific failure to the store using address as the key
func (k Keeper) AddContractFailure(ctx sdk.Context, failure types.Failure) {
	nextFailureId := k.GetNextFailureIdKey(ctx, failure.GetAddress())

	store := ctx.KVStore(k.storeKey)

	failure.Offset = nextFailureId
	b := k.cdc.MustMarshal(&failure)
	store.Set(types.GetFailureKey(failure.GetAddress(), nextFailureId), b)

	nextFailureId++
	store.Set(types.GetNextFailureIdKey(failure.GetAddress()), sdk.Uint64ToBigEndian(nextFailureId))
}

func (k Keeper) GetNextFailureIdKey(ctx sdk.Context, address string) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.GetNextFailureIdKey(address))
	if bytes == nil {
		k.Logger(ctx).Debug("Unable to get last registered failure key for the address. Returns 0")
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}

// GetContractFailures returns failures of the specific contract
func (k Keeper) GetContractFailures(ctx sdk.Context, address string) (list []types.Failure) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetFailureKeyPrefix(address))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Failure
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetAllFailure returns all failures
func (k Keeper) GetAllFailures(ctx sdk.Context) (list []types.Failure) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ContractFailuresKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Failure
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
