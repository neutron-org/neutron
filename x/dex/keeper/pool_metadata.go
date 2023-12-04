package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v2/x/dex/types"
)

// SetPoolMetadata set a specific poolMetadata in the store
func (k Keeper) SetPoolMetadata(ctx sdk.Context, poolMetadata types.PoolMetadata) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PoolMetadataKeyPrefix))
	b := k.cdc.MustMarshal(&poolMetadata)
	store.Set(GetPoolMetadataIDBytes(poolMetadata.Id), b)
}

// GetPoolMetadata returns a poolMetadata from its id
func (k Keeper) GetPoolMetadata(ctx sdk.Context, id uint64) (val types.PoolMetadata, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PoolMetadataKeyPrefix))
	b := store.Get(GetPoolMetadataIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) GetPoolMetadataByDenom(
	ctx sdk.Context,
	denom string,
) (pm types.PoolMetadata, err error) {
	poolID, err := types.ParsePoolIDFromDenom(denom)
	if err != nil {
		return pm, err
	}
	pm, found := k.GetPoolMetadata(ctx, poolID)
	if !found {
		return pm, types.ErrInvalidPoolDenom
	}
	return pm, nil
}

// RemovePoolMetadata removes a poolMetadata from the store
func (k Keeper) RemovePoolMetadata(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PoolMetadataKeyPrefix))
	store.Delete(GetPoolMetadataIDBytes(id))
}

// GetAllPoolMetadata returns all poolMetadata
func (k Keeper) GetAllPoolMetadata(ctx sdk.Context) (list []types.PoolMetadata) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PoolMetadataKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PoolMetadata
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetPoolMetadataIDBytes returns the byte representation of the ID
func GetPoolMetadataIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}
