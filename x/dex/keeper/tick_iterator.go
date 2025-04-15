package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

type TickIterator struct {
	iter storetypes.Iterator
	cdc  codec.BinaryCodec
}

func (k Keeper) NewTickIterator(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
) TickIterator {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.TickLiquidityPrefix(tradePairID))

	return TickIterator{
		iter: prefixStore.Iterator(nil, nil),
		cdc:  k.cdc,
	}
}

func (ti TickIterator) Valid() bool {
	return ti.iter.Valid()
}

func (ti TickIterator) Close() error {
	return ti.iter.Close()
}

func (ti TickIterator) Value() (tick types.TickLiquidity) {
	ti.cdc.MustUnmarshal(ti.iter.Value(), &tick)
	return tick
}

func (ti TickIterator) Next() {
	ti.iter.Next()
}
