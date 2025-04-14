package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// GetAllTickLiquidity returns all tickLiquidity
func (k Keeper) GetAllTickLiquidity(ctx sdk.Context) (list []*types.TickLiquidity) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := &types.TickLiquidity{}
		k.cdc.MustUnmarshal(iterator.Value(), val)
		list = append(list, val)
	}

	return
}

func (k Keeper) GetTickLiquidityIterator(ctx sdk.Context) storetypes.Iterator {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	return storetypes.KVStorePrefixIterator(store, []byte{})
}
