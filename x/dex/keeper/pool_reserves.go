package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (k Keeper) SetPoolReserves(ctx sdk.Context, poolReserves *types.PoolReserves) {
	tick := types.TickLiquidity{
		Liquidity: &types.TickLiquidity_PoolReserves{
			PoolReserves: poolReserves,
		},
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	b := k.cdc.MustMarshal(&tick)
	store.Set(poolReserves.Key.KeyMarshal(), b)
}

func (k Keeper) GetPoolReserves(
	ctx sdk.Context,
	poolReservesID *types.PoolReservesKey,
) (pool *types.PoolReserves, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	b := store.Get(poolReservesID.KeyMarshal())
	if b == nil {
		return nil, false
	}

	var tick types.TickLiquidity
	k.cdc.MustUnmarshal(b, &tick)

	return tick.GetPoolReserves(), true
}

// RemovePoolReserves removes a tickLiquidity from the store
func (k Keeper) RemovePoolReserves(ctx sdk.Context, poolReservesID *types.PoolReservesKey) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	store.Delete(poolReservesID.KeyMarshal())
}

// UpdatePoolReserves handles the logic for all updates to PoolReserves in the KV Store.
// NOTE: This method should always be called even if not all logic branches are applicable.
// It avoids unnecessary repetition of logic and provides a single place to attach update event handlers.
func (k Keeper) UpdatePoolReserves(ctx sdk.Context, reserves *types.PoolReserves, swapMetadata ...types.SwapMetadata) {
	if reserves.HasToken() {
		// The pool still has ReservesMakerDenom; save it as is
		k.SetPoolReserves(ctx, reserves)
	} else {
		ctx.EventManager().EmitEvents(types.GetEventsDecTotalPoolReserves(*reserves.Key.TradePairId.MustPairID()))
		// The pool is empty (ie. ReservesMakerDenom == 0); it can be safely deleted
		k.RemovePoolReserves(ctx, reserves.Key)
	}

	// TODO: This will create a bit of extra noise since UpdatePoolReserves is called for both sides of the pool,
	// but not in some cases only one side has been updated
	// This should be solved upstream by better tracking of dirty ticks
	ctx.EventManager().EmitEvent(types.CreateTickUpdatePoolReserves(*reserves, swapMetadata...))
}
