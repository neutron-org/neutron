package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"

	"github.com/neutron-org/neutron/v5/x/revenue/types"
)

const (
	NTRNSlinkyDenom = "NTRN"
	USDSlinkyDenom  = "USD"
)

// TODO: We currently store prices under a single store prefix. We need to handle cases where DenomCompensation is changed.

// UpdateCumulativePrice updates cumulative prices and
// does the maintenance of the storage by removing outdated TWAP prices.
func (k *Keeper) UpdateCumulativePrice(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module params: %w", err)
	}

	pair := slinkytypes.CurrencyPair{
		Base:  NTRNSlinkyDenom,
		Quote: USDSlinkyDenom,
	}
	priceInt, err := k.oracleKeeper.GetPriceForCurrencyPair(ctx, pair)
	if err != nil {
		return fmt.Errorf("failed to get price for currency pair: %w", err)
	}

	decimals, err := k.oracleKeeper.GetDecimalsForCurrencyPair(ctx, pair)
	if err != nil {
		return fmt.Errorf("failed to get decimals for currency pair: %w", err)
	}

	price := math.LegacyNewDecFromIntWithPrec(priceInt.Price, int64(decimals))
	err = k.SaveCumulativePrice(ctx, price, ctx.BlockTime().Unix())
	if err != nil {
		return fmt.Errorf("failed to save cumulative price: %w", err)
	}
	k.Logger(ctx).Debug("TWAP refresh", "price", price.String())

	err = k.CleanOutdatedCumulativePrices(ctx, ctx.BlockTime().Unix()-params.TwapWindow)
	if err != nil {
		return fmt.Errorf("failed to clean outdated prices: %w", err)
	}

	return nil
}

// SaveCumulativePrice saves new cumulative price
// accepts price and current timestamp as arguments
// and calculates cumulative price taking into account
// a time passed since last saved cumulative price
func (k *Keeper) SaveCumulativePrice(ctx sdk.Context, price math.LegacyDec, timestamp int64) error {
	cumulativePrevious, err := k.GetLastCumulativePrice(ctx)
	if err != nil {
		return err
	}

	cumulativeNew := types.CumulativePrice{
		CumulativePrice: cumulativePrevious.LastPrice.MulInt64(timestamp - cumulativePrevious.Timestamp).Add(cumulativePrevious.CumulativePrice),
		LastPrice:       price,
		Timestamp:       timestamp,
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&cumulativeNew)
	if err != nil {
		return fmt.Errorf("failed to marshal cumulative price: %w", err)
	}

	err = store.Set(types.GetAccumulatedPriceKey(timestamp), bz)
	if err != nil {
		return fmt.Errorf("failed to store cumulative price: %w", err)
	}

	return nil
}

// GetAllCumulativePrices get all the TWAP values
func (k *Keeper) GetAllCumulativePrices(ctx sdk.Context) ([]*types.CumulativePrice, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.PrefixAccumulatedPriceKey, storetypes.PrefixEndBytes(types.PrefixAccumulatedPriceKey))
	if err != nil {
		return nil, fmt.Errorf("failed to iterate over cumulative price: %w", err)
	}
	defer iter.Close()

	var prices []*types.CumulativePrice
	for ; iter.Valid(); iter.Next() {
		p := types.CumulativePrice{}
		if err = k.cdc.Unmarshal(iter.Value(), &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal a cumulative price: %w", err)
		}
		prices = append(prices, &p)
	}
	return prices, nil
}

// GetLastCumulativePrice gets very last element of TWAP storage
func (k *Keeper) GetLastCumulativePrice(ctx sdk.Context) (types.CumulativePrice, error) {
	cmlt := types.CumulativePrice{
		CumulativePrice: math.LegacyZeroDec(),
		LastPrice:       math.LegacyZeroDec(),
		Timestamp:       0,
	}

	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.ReverseIterator(types.PrefixAccumulatedPriceKey, storetypes.PrefixEndBytes(types.PrefixAccumulatedPriceKey))
	if err != nil {
		return cmlt, fmt.Errorf("failed to iterate over cumulative price: %w", err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if err = k.cdc.Unmarshal(iter.Value(), &cmlt); err != nil {
			return cmlt, fmt.Errorf("failed to unmarshal a cumulative price: %w", err)
		}
		return cmlt, nil
	}
	k.Logger(ctx).Warn("TWAP storage is empty")
	return cmlt, nil
}

// GetFirstCumulativePriceAfter gets first element after `startAt` of the TWAP storage
func (k *Keeper) GetFirstCumulativePriceAfter(ctx sdk.Context, startAt int64) (types.CumulativePrice, error) {
	cmlt := types.CumulativePrice{
		CumulativePrice: math.LegacyZeroDec(),
		LastPrice:       math.LegacyZeroDec(),
		Timestamp:       0,
	}

	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.GetAccumulatedPriceKey(startAt), storetypes.PrefixEndBytes(types.PrefixAccumulatedPriceKey))
	if err != nil {
		return cmlt, fmt.Errorf("failed to iterate over cumulative price: %w", err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if err = k.cdc.Unmarshal(iter.Value(), &cmlt); err != nil {
			return cmlt, fmt.Errorf("failed to unmarshal a cumulative price: %w", err)
		}
		return cmlt, nil
	}
	k.Logger(ctx).Warn("TWAP storage is empty")
	return cmlt, nil
}

// CleanOutdatedCumulativePrices removes all the cumulative
// prices those are older than a threshold
func (k *Keeper) CleanOutdatedCumulativePrices(ctx sdk.Context, cleanUntil int64) error {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(
		types.PrefixAccumulatedPriceKey,
		types.GetAccumulatedPriceKey(cleanUntil),
	)
	if err != nil {
		return fmt.Errorf("failed to iterate over cumulative price: %w", err)
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

// GetTWAPStartFromTime returns a TWAP price for window
// from `startAt` till last saved value (saved at every block)
func (k *Keeper) GetTWAPStartFromTime(ctx sdk.Context, startAt int64) (math.LegacyDec, error) {
	lastPrice, err := k.GetLastCumulativePrice(ctx)
	if err != nil {
		return math.LegacyZeroDec(), fmt.Errorf("failed to get last cumulative price: %w", err)
	}

	firstPrice, err := k.GetFirstCumulativePriceAfter(ctx, startAt)
	if err != nil {
		return math.LegacyZeroDec(), fmt.Errorf("failed to get first cumulative price: %w", err)
	}
	if lastPrice.Timestamp == firstPrice.Timestamp {
		return lastPrice.LastPrice, nil
	}
	return lastPrice.CumulativePrice.Sub(firstPrice.CumulativePrice).QuoInt64(lastPrice.Timestamp - firstPrice.Timestamp), nil
}
