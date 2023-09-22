package keeper

import (
	"encoding/json"

	db "github.com/cometbft/cometbft-db"
	"github.com/neutron-org/neutron/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// iterator returns an iterator over all gauges in the {prefix} space of state.
func (k Keeper) iterator(ctx sdk.Context, prefix []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, prefix)
}

// iterator returns an iterator over all gauges in the {prefix} space of state.
func (k Keeper) iteratorStartEnd(ctx sdk.Context, start []byte, end []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(start, end)
}

func UnmarshalRefArray(bz []byte) []uint64 {
	ids := []uint64{}
	err := json.Unmarshal(bz, &ids)
	if err != nil {
		panic(err)
	}
	return ids
}

// getStakesFromIterator returns an array of single stake units by period defined by the x/stake module.
func (k Keeper) getStakesFromIterator(ctx sdk.Context, iterator db.Iterator) types.Stakes {
	stakes := types.Stakes{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		stakeIDs := UnmarshalRefArray(iterator.Value())
		for _, stakeID := range stakeIDs {
			stake, err := k.GetStakeByID(ctx, stakeID)
			if err != nil {
				panic(err)
			}
			stakes = append(stakes, stake)
		}
	}
	return stakes
}

func (k Keeper) getIDsFromIterator(iterator db.Iterator) []uint64 {
	allIds := []uint64{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		ids := UnmarshalRefArray(iterator.Value())
		allIds = append(allIds, ids...)
	}
	return allIds
}
