package v3

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dex/types"
	v2types "github.com/neutron-org/neutron/v4/x/dex/types/v2"
)

// MigrateStore performs in-place store migrations.
// The migration adds new dex params -- GoodTilPurgeAllowance & MaxJITsPerBlock// for handling JIT orders.
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	if err := migrateParams(ctx, cdc, storeKey); err != nil {
		return err
	}

	if err := migrateLimitOrderExpirations(ctx, cdc, storeKey); err != nil {
		return err
	}

	if err := migrateTickLiquidityPrices(ctx, cdc, storeKey); err != nil {
		return err
	}

	if err := migrateInactiveTranchePrices(ctx, cdc, storeKey); err != nil {
		return err
	}

	return nil
}

func migrateParams(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating dex params...")

	// fetch old params
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.KeyPrefix(types.ParamsKey))
	if bz == nil {
		return errors.New("cannot fetch dex params from KV store")
	}
	var oldParams v2types.Params
	cdc.MustUnmarshal(bz, &oldParams)

	// add new param values
	newParams := types.Params{
		Paused:                types.DefaultPaused,
		FeeTiers:              oldParams.FeeTiers,
		GoodTilPurgeAllowance: types.DefaultGoodTilPurgeAllowance,
		MaxJitsPerBlock:       types.DefaultMaxJITsPerBlock,
	}

	// set params
	bz, err := cdc.Marshal(&newParams)
	if err != nil {
		return err
	}
	store.Set(types.KeyPrefix(types.ParamsKey), bz)

	ctx.Logger().Info("Finished migrating dex params")

	return nil
}

func migrateLimitOrderExpirations(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating dex LimitOrderExpirations...")

	// fetch list of all old limit order expirations
	expirationKeys := make([][]byte, 0)
	expirationVals := make([]*types.LimitOrderExpiration, 0)
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.LimitOrderExpirationKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	for ; iterator.Valid(); iterator.Next() {
		expirationKeys = append(expirationKeys, iterator.Key())
		var expiration types.LimitOrderExpiration
		cdc.MustUnmarshal(iterator.Value(), &expiration)
		expirationVals = append(expirationVals, &expiration)
	}

	err := iterator.Close()
	if err != nil {
		return errorsmod.Wrap(err, "iterator failed to close during migration")
	}

	for i, key := range expirationKeys {
		// re-save expiration with new key

		expiration := expirationVals[i]
		b := cdc.MustMarshal(expiration)
		store.Set(types.LimitOrderExpirationKey(
			expiration.ExpirationTime,
			expiration.TrancheRef,
		), b)
		// Delete record with old key
		store.Delete(key)
	}

	ctx.Logger().Info("Finished migrating dex LimitOrderExpirations")

	return nil
}

type migrationUpdate struct {
	key []byte
	val []byte
}

func migrateTickLiquidityPrices(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	// Due to change in precision of PrecDec between v2 and v3 we need to recompute all PrecDecs in the kvstore
	ctx.Logger().Info("Migrating TickLiquidity Prices...")

	// Iterate through all tickLiquidity
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	ticksToUpdate := make([]migrationUpdate, 0)

	for ; iterator.Valid(); iterator.Next() {
		var tickLiq types.TickLiquidity
		var updatedTickLiq types.TickLiquidity
		cdc.MustUnmarshal(iterator.Value(), &tickLiq)
		// Recalculate all prices
		switch liquidity := tickLiq.Liquidity.(type) {
		case *types.TickLiquidity_LimitOrderTranche:
			liquidity.LimitOrderTranche.PriceTakerToMaker = types.MustCalcPrice(liquidity.LimitOrderTranche.Key.TickIndexTakerToMaker)
			updatedTickLiq = types.TickLiquidity{Liquidity: liquidity}
		case *types.TickLiquidity_PoolReserves:
			poolReservesKey := liquidity.PoolReserves.Key
			liquidity.PoolReserves.PriceTakerToMaker = types.MustCalcPrice(poolReservesKey.TickIndexTakerToMaker)
			liquidity.PoolReserves.PriceOppositeTakerToMaker = poolReservesKey.Counterpart().MustPriceTakerToMaker()
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
	// Due to change in precision of PrecDec between v2 and v3 we need to recompute all PrecDecs in the kvstore
	ctx.Logger().Info("Migrating InactiveLimitOrderTranche Prices...")

	// Iterate through all InactiveTranches
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	ticksToUpdate := make([]migrationUpdate, 0)

	for ; iterator.Valid(); iterator.Next() {
		var tranche types.LimitOrderTranche
		cdc.MustUnmarshal(iterator.Value(), &tranche)
		// Recalculate price
		tranche.PriceTakerToMaker = types.MustCalcPrice(tranche.Key.TickIndexTakerToMaker)

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
