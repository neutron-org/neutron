package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/types"
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

// RemoveTickLiquidity removes a tickLiquidity from the store
func (k Keeper) RemovePoolReserves(ctx sdk.Context, poolReservesID *types.PoolReservesKey) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	store.Delete(poolReservesID.KeyMarshal())
}
