package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/types"
)

func NewLimitOrderExpiration(tranche *types.LimitOrderTranche) *types.LimitOrderExpiration {
	trancheExpiry := tranche.ExpirationTime
	if trancheExpiry == nil {
		panic("Cannot create LimitOrderExpiration from tranche with nil ExpirationTime")
	}

	return &types.LimitOrderExpiration{
		TrancheRef:     tranche.Key.KeyMarshal(),
		ExpirationTime: *tranche.ExpirationTime,
	}
}

// SetLimitOrderExpiration set a specific goodTilRecord in the store from its index
func (k Keeper) SetLimitOrderExpiration(
	ctx sdk.Context,
	goodTilRecord *types.LimitOrderExpiration,
) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefix(types.LimitOrderExpirationKeyPrefix),
	)
	b := k.cdc.MustMarshal(goodTilRecord)
	store.Set(types.LimitOrderExpirationKey(
		goodTilRecord.ExpirationTime,
		goodTilRecord.TrancheRef,
	), b)
}

// GetLimitOrderExpiration returns a goodTilRecord from its index
func (k Keeper) GetLimitOrderExpiration(
	ctx sdk.Context,
	goodTilDate time.Time,
	trancheRef []byte,
) (val *types.LimitOrderExpiration, found bool) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefix(types.LimitOrderExpirationKeyPrefix),
	)

	b := store.Get(types.LimitOrderExpirationKey(
		goodTilDate,
		trancheRef,
	))
	if b == nil {
		return val, false
	}

	val = &types.LimitOrderExpiration{}
	k.cdc.MustUnmarshal(b, val)

	return val, true
}

// RemoveLimitOrderExpiration removes a goodTilRecord from the store
func (k Keeper) RemoveLimitOrderExpiration(
	ctx sdk.Context,
	goodTilDate time.Time,
	trancheRef []byte,
) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefix(types.LimitOrderExpirationKeyPrefix),
	)
	store.Delete(types.LimitOrderExpirationKey(
		goodTilDate,
		trancheRef,
	))
}

func (k Keeper) RemoveLimitOrderExpirationByKey(ctx sdk.Context, key []byte) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefix(types.LimitOrderExpirationKeyPrefix),
	)
	store.Delete(key)
}

// GetAllLimitOrderExpiration returns all goodTilRecord
func (k Keeper) GetAllLimitOrderExpiration(ctx sdk.Context) (list []*types.LimitOrderExpiration) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefix(types.LimitOrderExpirationKeyPrefix),
	)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := &types.LimitOrderExpiration{}
		k.cdc.MustUnmarshal(iterator.Value(), val)
		list = append(list, val)
	}

	return
}

func (k Keeper) PurgeExpiredLimitOrders(ctx sdk.Context, curTime time.Time) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefix(types.LimitOrderExpirationKeyPrefix),
	)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	inGoodTilSegment := false

	archivedTranches := make(map[string]bool)
	defer iterator.Close()

	gasCutoff := ctx.BlockGasMeter().Limit() - types.GoodTilPurgeGasBuffer
	curBlockGas := ctx.BlockGasMeter().GasConsumed()

	for ; iterator.Valid(); iterator.Next() {
		var val types.LimitOrderExpiration
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if val.ExpirationTime.After(curTime) {
			return
		}

		inGoodTilSegment = inGoodTilSegment || val.ExpirationTime != types.JITGoodTilTime()
		gasConsumed := curBlockGas + ctx.GasMeter().GasConsumed()
		if inGoodTilSegment && gasConsumed >= gasCutoff {
			// If we hit our gas cutoff stop deleting so as not to timeout the block.
			// We can only do this if we are proccesing normal GT limitOrders
			// and not JIT limit orders, since there is not protection in place
			// to prevent JIT order from being traded on the next block.
			// This is ok since only GT limit orders pose a meaningful attack
			// vector since there is no upper bound on how many GT limit orders can be
			// canceled in a single block.
			ctx.EventManager().EmitEvent(types.GoodTilPurgeHitLimitEvent(gasConsumed))

			return
		}

		if _, ok := archivedTranches[string(val.TrancheRef)]; !ok {
			tranche, found := k.GetLimitOrderTrancheByKey(ctx, val.TrancheRef)
			if found {
				// Convert the tranche to an inactiveTranche
				k.SetInactiveLimitOrderTranche(ctx, tranche)
				k.RemoveLimitOrderTranche(ctx, tranche.Key)
				archivedTranches[string(val.TrancheRef)] = true
			}
		}

		k.RemoveLimitOrderExpirationByKey(ctx, iterator.Key())
		incExpiredOdrers()
	}
}
