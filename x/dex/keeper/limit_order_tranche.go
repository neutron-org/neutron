package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/types"
	"github.com/neutron-org/neutron/v3/x/dex/utils"
)

func NewLimitOrderTranche(
	limitOrderTrancheKey *types.LimitOrderTrancheKey,
	goodTil *time.Time,
) (*types.LimitOrderTranche, error) {
	priceTakerToMaker, err := limitOrderTrancheKey.PriceTakerToMaker()
	if err != nil {
		return nil, err
	}
	return &types.LimitOrderTranche{
		Key:                limitOrderTrancheKey,
		ReservesMakerDenom: math.ZeroInt(),
		ReservesTakerDenom: math.ZeroInt(),
		TotalMakerDenom:    math.ZeroInt(),
		TotalTakerDenom:    math.ZeroInt(),
		ExpirationTime:     goodTil,
		PriceTakerToMaker:  priceTakerToMaker,
	}, nil
}

func (k Keeper) FindLimitOrderTranche(
	ctx sdk.Context,
	limitOrderTrancheKey *types.LimitOrderTrancheKey,
) (val *types.LimitOrderTranche, fromFilled, found bool) {
	// Try to find the tranche in the active liq index
	tick := k.GetLimitOrderTranche(ctx, limitOrderTrancheKey)
	if tick != nil {
		return tick, false, true
	}
	// Look for filled limit orders
	tranche, found := k.GetInactiveLimitOrderTranche(ctx, limitOrderTrancheKey)
	if found {
		return tranche, true, true
	}

	return nil, false, false
}

func (k Keeper) SaveTranche(ctx sdk.Context, tranche *types.LimitOrderTranche) {
	if tranche.HasTokenIn() {
		k.SetLimitOrderTranche(ctx, tranche)
	} else {
		k.SetInactiveLimitOrderTranche(ctx, tranche)
		k.RemoveLimitOrderTranche(ctx, tranche.Key)
	}

	ctx.EventManager().EmitEvent(types.CreateTickUpdateLimitOrderTranche(tranche))
}

func (k Keeper) SetLimitOrderTranche(ctx sdk.Context, tranche *types.LimitOrderTranche) {
	// Wrap tranche back into TickLiquidity
	tick := types.TickLiquidity{
		Liquidity: &types.TickLiquidity_LimitOrderTranche{
			LimitOrderTranche: tranche,
		},
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	b := k.cdc.MustMarshal(&tick)
	store.Set(tranche.Key.KeyMarshal(), b)
}

func (k Keeper) GetLimitOrderTranche(
	ctx sdk.Context,
	limitOrderTrancheKey *types.LimitOrderTrancheKey,
) (tranche *types.LimitOrderTranche) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	b := store.Get(limitOrderTrancheKey.KeyMarshal())

	if b == nil {
		return nil
	}

	var tick types.TickLiquidity
	k.cdc.MustUnmarshal(b, &tick)

	return tick.GetLimitOrderTranche()
}

func (k Keeper) GetLimitOrderTrancheByKey(
	ctx sdk.Context,
	key []byte,
) (tranche *types.LimitOrderTranche, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	b := store.Get(key)

	if b == nil {
		return nil, false
	}

	var tick types.TickLiquidity
	k.cdc.MustUnmarshal(b, &tick)

	tranche = tick.GetLimitOrderTranche()
	if tranche != nil {
		return tranche, true
	}
	return nil, false
}

func (k Keeper) RemoveLimitOrderTranche(ctx sdk.Context, trancheKey *types.LimitOrderTrancheKey) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	store.Delete(trancheKey.KeyMarshal())
}

func (k Keeper) GetPlaceTranche(
	sdkCtx sdk.Context,
	tradePairID *types.TradePairID,
	tickIndexTakerToMaker int64,
) *types.LimitOrderTranche {
	prefixStore := prefix.NewStore(
		sdkCtx.KVStore(k.storeKey),
		types.TickLiquidityLimitOrderPrefix(tradePairID, tickIndexTakerToMaker),
	)
	iter := prefixStore.Iterator(nil, nil)

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var tick types.TickLiquidity
		k.cdc.MustUnmarshal(iter.Value(), &tick)
		tranche := tick.GetLimitOrderTranche()
		if tranche.IsPlaceTranche() {
			return tranche
		}
	}

	return nil
}

func (k Keeper) GetFillTranche(
	sdkCtx sdk.Context,
	tradePairID *types.TradePairID,
	tickIndexTakerToMaker int64,
) (*types.LimitOrderTranche, bool) {
	prefixStore := prefix.NewStore(
		sdkCtx.KVStore(k.storeKey),
		types.TickLiquidityLimitOrderPrefix(tradePairID, tickIndexTakerToMaker),
	)
	iter := sdk.KVStorePrefixIterator(prefixStore, []byte{})

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var tick types.TickLiquidity
		k.cdc.MustUnmarshal(iter.Value(), &tick)

		return tick.GetLimitOrderTranche(), true
	}

	return &types.LimitOrderTranche{}, false
}

func (k Keeper) GetAllLimitOrderTrancheAtIndex(
	sdkCtx sdk.Context,
	tradePairID *types.TradePairID,
	tickIndexTakerToMaker int64,
) (trancheList []types.LimitOrderTranche) {
	prefixStore := prefix.NewStore(
		sdkCtx.KVStore(k.storeKey),
		types.TickLiquidityLimitOrderPrefix(tradePairID, tickIndexTakerToMaker),
	)
	iter := sdk.KVStorePrefixIterator(prefixStore, []byte{})

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var tick types.TickLiquidity
		k.cdc.MustUnmarshal(iter.Value(), &tick)
		maybeTranche := tick.GetLimitOrderTranche()
		if maybeTranche != nil {
			trancheList = append(trancheList, *maybeTranche)
		}
	}

	return trancheList
}

func NewTrancheKey(ctx sdk.Context) string {
	blockHeight := ctx.BlockHeight()
	txGas := ctx.GasMeter().GasConsumed()
	blockGas := utils.MustGetBlockGasUsed(ctx)
	totalGas := blockGas + txGas

	blockStr := utils.Uint64ToSortableString(uint64(blockHeight))
	gasStr := utils.Uint64ToSortableString(totalGas)

	return fmt.Sprintf("%s%s", blockStr, gasStr)
}

func (k Keeper) GetOrInitPlaceTranche(ctx sdk.Context,
	tradePairID *types.TradePairID,
	tickIndexTakerToMaker int64,
	goodTil *time.Time,
	orderType types.LimitOrderType,
) (placeTranche *types.LimitOrderTranche, err error) {
	// NOTE: Right now we are not indexing by goodTil date so we can't easily check if there's already a tranche
	// with the same goodTil date so instead we create a new tranche for each goodTil order
	// if there is a large number of limitOrders with the same goodTilTime (most likely JIT)
	// aggregating might be more efficient particularly for deletion, but if they are relatively sparse
	// it will incur fewer lookups to just create a new limitOrderTranche
	// Also trying to cancel aggregated good_til orders will be a PITA
	JITGoodTilTime := types.JITGoodTilTime()
	switch orderType {
	case types.LimitOrderType_JUST_IN_TIME:
		limitOrderTrancheKey := &types.LimitOrderTrancheKey{
			TradePairId:           tradePairID,
			TickIndexTakerToMaker: tickIndexTakerToMaker,
			TrancheKey:            NewTrancheKey(ctx),
		}
		placeTranche, err = NewLimitOrderTranche(limitOrderTrancheKey, &JITGoodTilTime)
	case types.LimitOrderType_GOOD_TIL_TIME:
		limitOrderTrancheKey := &types.LimitOrderTrancheKey{
			TradePairId:           tradePairID,
			TickIndexTakerToMaker: tickIndexTakerToMaker,
			TrancheKey:            NewTrancheKey(ctx),
		}
		placeTranche, err = NewLimitOrderTranche(limitOrderTrancheKey, goodTil)
	default:
		placeTranche = k.GetPlaceTranche(ctx, tradePairID, tickIndexTakerToMaker)
		if placeTranche == nil {
			limitOrderTrancheKey := &types.LimitOrderTrancheKey{
				TradePairId:           tradePairID,
				TickIndexTakerToMaker: tickIndexTakerToMaker,
				TrancheKey:            NewTrancheKey(ctx),
			}
			placeTranche, err = NewLimitOrderTranche(limitOrderTrancheKey, nil)
			if err != nil {
				return nil, err
			}
		}
	}
	if err != nil {
		return nil, err
	}

	return placeTranche, nil
}
