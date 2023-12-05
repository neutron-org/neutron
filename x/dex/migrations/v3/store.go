package v2

import (
	"bytes"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v2/x/dex/types"
	typesv2 "github.com/neutron-org/neutron/v2/x/dex/types/v2"
)

func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	err := migrateLimitOrders(ctx, cdc, storeKey)
	if err != nil {
		return err
	}

	err = migrateInactiveLimitOrders(ctx, cdc, storeKey)
	if err != nil {
		return err
	}

	return migrateLimitOrderExpirations(ctx, cdc, storeKey)
}

func tickLiquidityV2ToV3(tick typesv2.TickLiquidity) types.TickLiquidity {
	oldTranche := tick.GetLimitOrderTranche()
	newTranche := limitOrderTrancheV2ToV3(*oldTranche)

	return types.TickLiquidity{
		Liquidity: &types.TickLiquidity_LimitOrderTranche{
			LimitOrderTranche: newTranche,
		},
	}
}

func limitOrderTrancheV2ToV3(tranche typesv2.LimitOrderTranche) *types.LimitOrderTranche {
	var expirationTimeInt64 int64
	var orderType types.LimitOrderType
	switch tranche.ExpirationTime {
	case nil:
		expirationTimeInt64 = 0
		orderType = types.LimitOrderType_GOOD_TIL_CANCELLED
	case &time.Time{}:
		expirationTimeInt64 = 0
		orderType = types.LimitOrderType_JUST_IN_TIME
	default:
		expirationTimeInt64 = tranche.ExpirationTime.Unix()
		orderType = types.LimitOrderType_GOOD_TIL_TIME
	}
	return &types.LimitOrderTranche{
		Key:                (*types.LimitOrderTrancheKey)(tranche.Key),
		ReservesMakerDenom: tranche.ReservesMakerDenom,
		ReservesTakerDenom: tranche.ReservesTakerDenom,
		TotalMakerDenom:    tranche.TotalMakerDenom,
		TotalTakerDenom:    tranche.TotalTakerDenom,
		PriceTakerToMaker:  tranche.PriceTakerToMaker,
		ExpirationTime:     expirationTimeInt64,
		Type:               orderType,
	}
}

func limitOrderExpirationV2ToV3(loExpiration typesv2.LimitOrderExpiration) types.LimitOrderExpiration {
	var expirationTimeInt64 int64
	switch loExpiration.ExpirationTime {
	case time.Time{}:
		expirationTimeInt64 = 0
	default:
		expirationTimeInt64 = loExpiration.ExpirationTime.Unix()
	}
	return types.LimitOrderExpiration{
		ExpirationTime: expirationTimeInt64,
		TrancheRef:     loExpiration.TrancheRef,
	}
}

func migrateLimitOrders(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating limitOrders ticks...")

	// fetch list of all LimitOrder ticks
	tickPrefix := types.KeyPrefix(types.TickLiquidityKeyPrefix)
	oldLOTicks := make([]typesv2.TickLiquidity, 0)
	loKeys := make([][]byte, 0)
	tickStore := prefix.NewStore(ctx.KVStore(storeKey), tickPrefix)

	iterator := sdk.KVStorePrefixIterator(tickStore, []byte{})

	for ; iterator.Valid(); iterator.Next() {
		if bytes.Contains(iterator.Key(), []byte(types.LiquidityTypeLimitOrder)) {
			var tick typesv2.TickLiquidity
			cdc.MustUnmarshal(iterator.Value(), &tick)
			oldLOTicks = append(oldLOTicks, tick)
			loKeys = append(loKeys, iterator.Key())
		}
	}

	err := iterator.Close()
	if err != nil {
		return err
	}

	// migrate tranches
	for i, oldTick := range oldLOTicks {
		newTick := tickLiquidityV2ToV3(oldTick)
		key := loKeys[i]
		bz := cdc.MustMarshal(&newTick)
		tickStore.Set(key, bz)
	}

	ctx.Logger().Info("Finished migrating limitOrders")

	return nil
}

func migrateInactiveLimitOrders(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating InactivelimitOrders...")

	// fetch list of all InactiveLimitOrders
	inactiveLimitOrderPrefix := types.KeyPrefix(types.InactiveLimitOrderTrancheKeyPrefix)
	oldInactiveLOs := make([]typesv2.LimitOrderTranche, 0)
	inactiveLOKeys := make([][]byte, 0)
	inactiveLOStore := prefix.NewStore(ctx.KVStore(storeKey), inactiveLimitOrderPrefix)
	iterator := sdk.KVStorePrefixIterator(inactiveLOStore, []byte{})

	for ; iterator.Valid(); iterator.Next() {
		var inactiveLO typesv2.LimitOrderTranche
		cdc.MustUnmarshal(iterator.Value(), &inactiveLO)
		oldInactiveLOs = append(oldInactiveLOs, inactiveLO)
		inactiveLOKeys = append(inactiveLOKeys, iterator.Key())
	}

	err := iterator.Close()
	if err != nil {
		return err
	}

	// migrate InactiveLimitOrders
	for i, oldInactiveLO := range oldInactiveLOs {
		newInactiveLO := limitOrderTrancheV2ToV3(oldInactiveLO)
		key := inactiveLOKeys[i]
		bz := cdc.MustMarshal(newInactiveLO)
		inactiveLOStore.Set(key, bz)
	}

	ctx.Logger().Info("Finished migrating InactiveLimitOrders")

	return nil
}

func migrateLimitOrderExpirations(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating LimitOrderExpirations...")

	// fetch list of all LimitOrderExpirations
	limitOrderExpirationPrefix := types.KeyPrefix(types.LimitOrderExpirationKeyPrefix)
	oldLOExpirations := make([]typesv2.LimitOrderExpiration, 0)
	expirationKeys := make([][]byte, 0)
	loExpirationStore := prefix.NewStore(ctx.KVStore(storeKey), limitOrderExpirationPrefix)
	iterator := sdk.KVStorePrefixIterator(loExpirationStore, []byte{})

	for ; iterator.Valid(); iterator.Next() {
		var loExpiration typesv2.LimitOrderExpiration
		cdc.MustUnmarshal(iterator.Value(), &loExpiration)
		oldLOExpirations = append(oldLOExpirations, loExpiration)
		expirationKeys = append(expirationKeys, iterator.Key())
	}

	err := iterator.Close()
	if err != nil {
		return err
	}

	// migrate InactiveLimitOrders
	for i, oldInactiveLO := range oldLOExpirations {
		oldKey := expirationKeys[i]
		loExpirationStore.Delete(oldKey)

		newLOExpiration := limitOrderExpirationV2ToV3(oldInactiveLO)
		newKey := types.LimitOrderExpirationKey(newLOExpiration.ExpirationTime, newLOExpiration.TrancheRef)
		bz := cdc.MustMarshal(&newLOExpiration)
		loExpirationStore.Set(newKey, bz)
	}

	ctx.Logger().Info("Finished migrating LimitOrderExpirations")

	return nil
}
