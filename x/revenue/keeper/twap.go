package keeper

import (
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v5/x/revenue/types"
)

func (k *Keeper) SaveCumulativePrice(ctx sdk.Context, price math.LegacyDec, timestamp uint64) error {
	cumulativeFirst, err := k.GetLastCumulativePrice(ctx)
	if err != nil {
		return err
	}

	cumulativeNew := types.CumulativePrice{
		CumulativePrice: cumulativeFirst.LastPrice.MulInt64(int64(timestamp - cumulativeFirst.Timestamp)).Add(cumulativeFirst.CumulativePrice),
		LastPrice:       price,
		Timestamp:       timestamp,
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&cumulativeNew)
	if err != nil {
		return err
	}

	err = store.Set(types.GetAccumulatedPriceKey(timestamp), bz)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keeper) GetAllCumulativePrices(ctx sdk.Context) ([]*types.CumulativePrice, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.ReverseIterator(types.PrefixAccumulatedPriceKey, storetypes.PrefixEndBytes(types.PrefixAccumulatedPriceKey))
	if err != nil {
		return nil, fmt.Errorf("failed to iterate over accumulated price: %w", err)
	}
	defer iter.Close()

	prices := []*types.CumulativePrice{}
	for ; iter.Valid(); iter.Next() {
		p := types.CumulativePrice{}
		if err = k.cdc.Unmarshal(iter.Value(), &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal a accumulated price: %w", err)
		}
		prices = append(prices, &p)
	}
	k.Logger(ctx).Error("TWAP storage is empty")
	return prices, nil
}

func (k *Keeper) GetLastCumulativePrice(ctx sdk.Context) (types.CumulativePrice, error) {
	cmlt := types.CumulativePrice{
		CumulativePrice: math.LegacyZeroDec(),
		LastPrice:       math.LegacyZeroDec(),
		Timestamp:       0,
	}

	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.ReverseIterator(types.PrefixAccumulatedPriceKey, storetypes.PrefixEndBytes(types.PrefixAccumulatedPriceKey))
	if err != nil {
		return cmlt, fmt.Errorf("failed to iterate over accumulated price: %w", err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if err = k.cdc.Unmarshal(iter.Value(), &cmlt); err != nil {
			return cmlt, fmt.Errorf("failed to unmarshal a accumulated price: %w", err)
		}
		return cmlt, nil
	}
	k.Logger(ctx).Error("TWAP storage is empty")
	return cmlt, nil
}

func (k *Keeper) GetFirstCumulativePriceAfter(ctx sdk.Context, startAt uint64) (types.CumulativePrice, error) {
	cmlt := types.CumulativePrice{
		CumulativePrice: math.LegacyZeroDec(),
		LastPrice:       math.LegacyZeroDec(),
		Timestamp:       0,
	}

	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.GetAccumulatedPriceKey(startAt), storetypes.PrefixEndBytes(types.PrefixAccumulatedPriceKey))
	if err != nil {
		return cmlt, fmt.Errorf("failed to iterate over validator info: %w", err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if err = k.cdc.Unmarshal(iter.Value(), &cmlt); err != nil {
			return cmlt, fmt.Errorf("failed to unmarshal a validator info: %w", err)
		}
		return cmlt, nil
	}
	k.Logger(ctx).Error("TWAP storage is empty")
	return cmlt, nil
}

func (k *Keeper) CleanOutdatedCumulativePrices(ctx sdk.Context) error {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(
		types.PrefixAccumulatedPriceKey,
		storetypes.PrefixEndBytes(types.GetAccumulatedPriceKey(uint64(ctx.BlockTime().Unix())-types.MaxTWAPWindow)),
	)
	if err != nil {
		return fmt.Errorf("failed to iterate over validator info: %w", err)
	}
	defer iter.Close()

	var keysToRemove [][]byte
	for ; iter.Valid(); iter.Next() {
		keysToRemove = append(keysToRemove, iter.Key())
	}

	for _, key := range keysToRemove {
		err = store.Delete(key)
		if err != nil {
			return fmt.Errorf("failed to remove key {%s} from the store: %w", key, err)
		}
	}
	return nil
}

//func getFirstIterValue[T proto.Marshaler](iter dbm.Iterator) ([]byte, T, error) {
//	var value T
//	for ; iter.Valid(); iter.Next() {
//		twap := sdk.DecCoin{}
//		if err = k.cdc.Unmarshal(iter.Value(), &twap); err != nil {
//			return nil, 0, fmt.Errorf("failed to unmarshal a validator info: %w", err)
//		}
//		return &twap, types.GetTimeFromAccumulatedPriceKey(iter.Key()), nil
//	}
//}

func (k *Keeper) GetTWAPForTime(ctx sdk.Context, startAt uint64) (math.LegacyDec, error) {
	lastPrice, err := k.GetLastCumulativePrice(ctx)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	firstPrice, err := k.GetFirstCumulativePriceAfter(ctx, startAt)
	if err != nil {
		return math.LegacyZeroDec(), err
	}
	if lastPrice == firstPrice {
		return lastPrice.LastPrice, nil
	}
	return lastPrice.CumulativePrice.Sub(firstPrice.CumulativePrice).QuoInt64(int64(lastPrice.Timestamp - firstPrice.Timestamp)), nil
}
