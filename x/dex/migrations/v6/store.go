package v6

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v8/utils/math"
	"github.com/neutron-org/neutron/v8/x/dex/types"
)

// MigrateStore performs in-place store migrations.
// Add DecXXX fields to all PoolReserves and LimitOrderTranche and InactiveLimitOrderTranche
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	if err := migrateTickLiquidity(ctx, cdc, storeKey); err != nil {
		return err
	}

	if err := migrateInactiveTranches(ctx, cdc, storeKey); err != nil {
		return err
	}

	return nil
}

type migrationUpdate struct {
	key []byte
	val []byte
}

func migrateTickLiquidity(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating TickLiquidity fields...")

	// Iterate through all tickLiquidity
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	ticksToUpdate := make([]migrationUpdate, 0)

	for ; iterator.Valid(); iterator.Next() {
		var tickLiq types.TickLiquidity
		var updatedTickLiq types.TickLiquidity
		cdc.MustUnmarshal(iterator.Value(), &tickLiq)

		// Add DecXXX fields
		switch liquidity := tickLiq.Liquidity.(type) {
		case *types.TickLiquidity_LimitOrderTranche:
			liquidity.LimitOrderTranche.DecReservesMakerDenom = math_utils.NewPrecDecFromInt(liquidity.LimitOrderTranche.ReservesMakerDenom)
			liquidity.LimitOrderTranche.DecReservesTakerDenom = math_utils.NewPrecDecFromInt(liquidity.LimitOrderTranche.ReservesTakerDenom)
			liquidity.LimitOrderTranche.DecTotalTakerDenom = math_utils.NewPrecDecFromInt(liquidity.LimitOrderTranche.TotalTakerDenom)
			updatedTickLiq = types.TickLiquidity{Liquidity: liquidity}
		case *types.TickLiquidity_PoolReserves:
			liquidity.PoolReserves.DecReservesMakerDenom = math_utils.NewPrecDecFromInt(liquidity.PoolReserves.ReservesMakerDenom)

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

	ctx.Logger().Info("Finished migrating TickLiquidity fields...")

	return nil
}

func migrateInactiveTranches(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating InactiveLimitOrderTranche fields...")

	// Iterate through all InactiveLimitOrderTranche
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	inactiveTranchesToUpdate := make([]migrationUpdate, 0)

	for ; iterator.Valid(); iterator.Next() {
		var tranche types.LimitOrderTranche
		cdc.MustUnmarshal(iterator.Value(), &tranche)

		tranche.DecReservesMakerDenom = math_utils.NewPrecDecFromInt(tranche.ReservesMakerDenom)
		tranche.DecReservesTakerDenom = math_utils.NewPrecDecFromInt(tranche.ReservesTakerDenom)
		tranche.DecTotalTakerDenom = math_utils.NewPrecDecFromInt(tranche.TotalTakerDenom)

		bz := cdc.MustMarshal(&tranche)
		inactiveTranchesToUpdate = append(inactiveTranchesToUpdate, migrationUpdate{key: iterator.Key(), val: bz})

	}

	err := iterator.Close()
	if err != nil {
		return errorsmod.Wrap(err, "iterator failed to close during migration")
	}

	// Store the updated TickLiquidity
	for _, v := range inactiveTranchesToUpdate {
		store.Set(v.key, v.val)
	}

	ctx.Logger().Info("Finished migrating InactiveLimitOrderTranche fields...")

	return nil
}
