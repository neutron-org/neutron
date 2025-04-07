package keeper

import (
	"encoding/binary"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
	"github.com/neutron-org/neutron/v6/x/dex/utils"
)

func NewLimitOrderTranche(
	limitOrderTrancheKey *types.LimitOrderTrancheKey,
	goodTil *time.Time,
) (*types.LimitOrderTranche, error) {
	price, err := limitOrderTrancheKey.Price()
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
		MakerPrice:         price,
		PriceTakerToMaker:  math_utils.OnePrecDec().Quo(price),
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

// UpdateTranche handles the logic for all updates to active LimitOrderTranches in the KV Store.
// NOTE: This method should always be called even if not all logic branches are applicable.
// It avoids unnecessary repetition of logic and provides a single place to attach update event handlers.
func (k Keeper) UpdateTranche(ctx sdk.Context, tranche *types.LimitOrderTranche, swapMetadata ...types.SwapMetadata) {
	switch {

	// Tranche still has TokenIn (ReservesMakerDenom) ==> Just save the update
	case tranche.HasTokenIn():
		k.SetLimitOrderTranche(ctx, tranche)

	// There is no TokenIn but there is still withdrawable TokenOut (ReservesTakerDenom) ==> Remove the active tranche and create a new inactive tranche
	case tranche.HasTokenOut():
		k.SetInactiveLimitOrderTranche(ctx, tranche)
		k.RemoveLimitOrderTranche(ctx, tranche.Key)
		// We are removing liquidity from the orderbook so we emit an event
		ctx.EventManager().EmitEvents(types.GetEventsDecTotalOrders(tranche.Key.TradePairId))

	// There is no TokenIn or Token Out ==> We can delete the tranche entirely
	default:
		k.RemoveLimitOrderTranche(ctx, tranche.Key)
		// We are removing liquidity from the orderbook so we emit an event
		ctx.EventManager().EmitEvents(types.GetEventsDecTotalOrders(tranche.Key.TradePairId))
	}

	ctx.EventManager().EmitEvent(types.CreateTickUpdateLimitOrderTranche(tranche, swapMetadata...))
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

func (k Keeper) GetGTCPlaceTranche(
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
		// Make sure tranche has not been traded through and is a GTC tranche
		if tranche.IsPlaceTranche() && !tranche.HasExpiration() {
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
	iter := storetypes.KVStorePrefixIterator(prefixStore, []byte{})

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
	iter := storetypes.KVStorePrefixIterator(prefixStore, []byte{})

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

	blockStr := utils.Uint64ToSortableString(uint64(blockHeight)) //nolint:gosec
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
		ctx.EventManager().EmitEvents(types.GetEventsIncTotalOrders(tradePairID))
	case types.LimitOrderType_GOOD_TIL_TIME:
		limitOrderTrancheKey := &types.LimitOrderTrancheKey{
			TradePairId:           tradePairID,
			TickIndexTakerToMaker: tickIndexTakerToMaker,
			TrancheKey:            NewTrancheKey(ctx),
		}
		placeTranche, err = NewLimitOrderTranche(limitOrderTrancheKey, goodTil)
		ctx.EventManager().EmitEvents(types.GetEventsIncExpiringOrders(tradePairID))
	default:
		placeTranche = k.GetGTCPlaceTranche(ctx, tradePairID, tickIndexTakerToMaker)
		if placeTranche == nil {
			limitOrderTrancheKey := &types.LimitOrderTrancheKey{
				TradePairId:           tradePairID,
				TickIndexTakerToMaker: tickIndexTakerToMaker,
				TrancheKey:            NewTrancheKey(ctx),
			}
			placeTranche, err = NewLimitOrderTranche(limitOrderTrancheKey, nil)
			ctx.EventManager().EmitEvents(types.GetEventsIncTotalOrders(tradePairID))
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

// GetJITsInBlockCount gets the total number of JIT LimitOrders placed in a block
func (k Keeper) GetJITsInBlockCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.TransientStore(k.tKey), []byte{})
	byteKey := types.KeyPrefix(types.JITsInBlockKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetJITsInBlockCount sets the total number of JIT LimitOrders placed in a block
func (k Keeper) SetJITsInBlockCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.TransientStore(k.tKey), []byte{})
	byteKey := types.KeyPrefix(types.JITsInBlockKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

func (k Keeper) IncrementJITsInBlock(ctx sdk.Context) {
	currentCount := k.GetJITsInBlockCount(ctx)
	k.SetJITsInBlockCount(ctx, currentCount+1)
}
