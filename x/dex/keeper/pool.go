package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v2/x/dex/types"
	"github.com/neutron-org/neutron/v2/x/dex/utils"
)

func (k Keeper) GetOrInitPool(
	ctx sdk.Context,
	pairID *types.PairID,
	centerTickIndexNormalized int64,
	fee uint64,
) (*types.Pool, error) {
	pool, found := k.GetPool(ctx, pairID, centerTickIndexNormalized, fee)
	if found {
		return pool, nil
	}

	return k.InitPool(ctx, pairID, centerTickIndexNormalized, fee)
}

func (k Keeper) InitPool(
	ctx sdk.Context,
	pairID *types.PairID,
	centerTickIndexNormalized int64,
	fee uint64,
) (pool *types.Pool, err error) {
	poolID := k.initializePoolMetadata(ctx, pairID, centerTickIndexNormalized, fee)

	k.storePoolIDRef(ctx, poolID, pairID, centerTickIndexNormalized, fee)

	return types.NewPool(pairID, centerTickIndexNormalized, fee, poolID)
}

func (k Keeper) storePoolIDRef(
	ctx sdk.Context,
	poolID uint64,
	pairID *types.PairID,
	centerTickIndexNormalized int64,
	fee uint64,
) {
	poolIDBz := sdk.Uint64ToBigEndian(poolID)
	poolIDKey := types.PoolIDKey(pairID, centerTickIndexNormalized, fee)

	poolIDStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PoolIDKeyPrefix))
	poolIDStore.Set(poolIDKey, poolIDBz)
}

func (k Keeper) incrementPoolCount(ctx sdk.Context) {
	currentCount := k.GetPoolCount(ctx)
	k.SetPoolCount(ctx, currentCount+1)
}

// GetNextPoolId get ID for the next pool to be created
func (k Keeper) GetNextPoolID(ctx sdk.Context) uint64 {
	return k.GetPoolCount(ctx)
}

func (k Keeper) GetPool(
	ctx sdk.Context,
	pairID *types.PairID,
	centerTickIndexNormalized int64,
	fee uint64,
) (pool *types.Pool, found bool) {
	poolID, found := k.GetPoolIDByParams(ctx, pairID, centerTickIndexNormalized, fee)
	if !found {
		return nil, false
	}

	return k.getPoolByPoolID(ctx, poolID, pairID, centerTickIndexNormalized, fee)
}

func (k Keeper) getPoolByPoolID(
	ctx sdk.Context,
	poolID uint64,
	pairID *types.PairID,
	centerTickIndexNormalized int64,
	fee uint64,
) (pool *types.Pool, found bool) {
	feeInt64 := utils.MustSafeUint64ToInt64(fee)
	id0To1 := &types.PoolReservesKey{
		TradePairId:           types.NewTradePairIDFromMaker(pairID, pairID.Token1),
		TickIndexTakerToMaker: centerTickIndexNormalized + feeInt64,
		Fee:                   fee,
	}

	upperTick, upperTickFound := k.GetPoolReserves(ctx, id0To1)
	lowerTick, lowerTickFound := k.GetPoolReserves(ctx, id0To1.Counterpart())

	switch {
	case !lowerTickFound && upperTickFound:
		lowerTick = types.NewPoolReservesFromCounterpart(upperTick)
	case lowerTickFound && !upperTickFound:
		upperTick = types.NewPoolReservesFromCounterpart(lowerTick)
	case !lowerTickFound && !upperTickFound:
		// Pool has already been initialized before, пso we can safely assume that pool creation doesn't throw an error
		return types.MustNewPool(pairID, centerTickIndexNormalized, fee, poolID), true
	}

	return &types.Pool{
		Id:         poolID,
		LowerTick0: lowerTick,
		UpperTick1: upperTick,
	}, true
}

func (k Keeper) GetPoolByID(ctx sdk.Context, poolID uint64) (pool *types.Pool, found bool) {
	poolMetadata, found := k.GetPoolMetadata(ctx, poolID)
	if !found {
		return pool, false
	}

	return k.getPoolByPoolID(ctx, poolID, poolMetadata.PairId, poolMetadata.Tick, poolMetadata.Fee)
}

func (k Keeper) GetPoolIDByParams(
	ctx sdk.Context,
	pairID *types.PairID,
	centerTickIndexNormalized int64,
	fee uint64,
) (id uint64, found bool) {
	poolIDKey := types.PoolIDKey(pairID, centerTickIndexNormalized, fee)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PoolIDKeyPrefix))
	b := store.Get(poolIDKey)
	if b == nil {
		return 0, false
	}

	poolID := sdk.BigEndianToUint64(b)
	return poolID, true
}

func (k Keeper) SetPool(ctx sdk.Context, pool *types.Pool) {
	k.updatePoolReserves(ctx, pool.LowerTick0)
	k.updatePoolReserves(ctx, pool.UpperTick1)

	// TODO: this will create a bit of extra noise since not every Save is updating both ticks
	// This should be solved upstream by better tracking of dirty ticks
	ctx.EventManager().EmitEvent(types.CreateTickUpdatePoolReserves(*pool.LowerTick0))
	ctx.EventManager().EmitEvent(types.CreateTickUpdatePoolReserves(*pool.UpperTick1))
}

func (k Keeper) updatePoolReserves(ctx sdk.Context, reserves *types.PoolReserves) {
	if reserves.HasToken() {
		k.SetPoolReserves(ctx, reserves)
	} else {
		k.RemovePoolReserves(ctx, reserves.Key)
	}
}

// GetPoolCount get the total number of pools
func (k Keeper) GetPoolCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.PoolCountKeyPrefix)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetPoolCount set the total number of pools
func (k Keeper) SetPoolCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.PoolCountKeyPrefix)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}
