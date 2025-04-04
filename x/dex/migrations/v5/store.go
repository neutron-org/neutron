package v5

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// MigrateStore performs in-place store migrations.
// v5 adds a new field `MakerPrice` to all tickLiquidity. It must be calculated and added
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	if err := migrateTickLiquidityPrices(ctx, cdc, storeKey); err != nil {
		return err
	}

	if err := migrateInactiveTranchePrices(ctx, cdc, storeKey); err != nil {
		return err
	}

	return nil
}

type migrationUpdate struct {
	key []byte
	val []byte
}

func migrateTickLiquidityPrices(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating TickLiquidity Prices...")

	// Iterate through all tickLiquidity
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	ticksToUpdate := make([]migrationUpdate, 0)

	for ; iterator.Valid(); iterator.Next() {
		var tickLiq types.TickLiquidity
		var updatedTickLiq types.TickLiquidity
		cdc.MustUnmarshal(iterator.Value(), &tickLiq)
		// Add MakerPrice
		switch liquidity := tickLiq.Liquidity.(type) {
		case *types.TickLiquidity_LimitOrderTranche:
			liquidity.LimitOrderTranche.MakerPrice = types.MustCalcPrice(liquidity.LimitOrderTranche.Key.TickIndexTakerToMaker)
			updatedTickLiq = types.TickLiquidity{Liquidity: liquidity}
		case *types.TickLiquidity_PoolReserves:
			poolReservesKey := liquidity.PoolReserves.Key
			liquidity.PoolReserves.MakerPrice = types.MustCalcPrice(poolReservesKey.TickIndexTakerToMaker)
			updatedTickLiq = types.TickLiquidity{Liquidity: liquidity}

		default:
			panic("Tick does not contain valid liqudityType")
		}

		bz := cdc.MustMarshal(&updatedTickLiq)
		ticksToUpdate = append(ticksToUpdate, migrationUpdate{key: iterator.Key(), val: bz})

	}

	err := iterator.Close()
	if err != nil {
		return errorsmod.Wrap(err, "iterator failed to close during migration")
	}

	// Store the updated TickLiquidity
	for _, v := range ticksToUpdate {
		store.Set(v.key, v.val)
	}

	ctx.Logger().Info("Finished migrating TickLiquidity Prices...")

	return nil
}

func migrateInactiveTranchePrices(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating InactiveLimitOrderTranche Prices...")

	// Iterate through all InactiveTranches
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	ticksToUpdate := make([]migrationUpdate, 0)

	for ; iterator.Valid(); iterator.Next() {
		var tranche types.LimitOrderTranche
		cdc.MustUnmarshal(iterator.Value(), &tranche)
		// Add MakerPrice
		tranche.MakerPrice = types.MustCalcPrice(tranche.Key.TickIndexTakerToMaker)

		bz := cdc.MustMarshal(&tranche)
		ticksToUpdate = append(ticksToUpdate, migrationUpdate{key: iterator.Key(), val: bz})
	}

	err := iterator.Close()
	if err != nil {
		return errorsmod.Wrap(err, "iterator failed to close during migration")
	}

	// Store the updated InactiveTranches
	for _, v := range ticksToUpdate {
		store.Set(v.key, v.val)
	}

	ctx.Logger().Info("Finished migrating InactiveLimitOrderTranche Prices...")

	return nil
}
