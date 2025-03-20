package keeper

import (
	"fmt"
	stdmath "math"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"

	"github.com/neutron-org/neutron/v5/x/revenue/types"
)

const (
	// ntrnSlinkyDenom is the Slinky identifier of the reward denom.
	ntrnSlinkyDenom = "NTRN"
	// usdSlinkyDenom is the Slinky identifier of USD.
	usdSlinkyDenom = "USD"
)

// UpdateRewardAssetPrice stores fresh cumulative and absolute price of the reward asset and cleans
// out the cumulative prices ledger from outdated records (that are older than the TWAP window).
func (k *Keeper) UpdateRewardAssetPrice(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module params: %w", err)
	}

	pair := slinkytypes.CurrencyPair{
		Base:  ntrnSlinkyDenom,
		Quote: usdSlinkyDenom,
	}
	priceInt, err := k.oracleKeeper.GetPriceForCurrencyPair(ctx, pair)
	if err != nil {
		return fmt.Errorf("failed to get price for currency pair: %w", err)
	}
	// safecheck. Should never happen. Make sure the slinky price is valid
	if priceInt.Price.LTE(math.ZeroInt()) {
		return fmt.Errorf("price is invalid")
	}

	decimals, err := k.oracleKeeper.GetDecimalsForCurrencyPair(ctx, pair)
	if err != nil {
		return fmt.Errorf("failed to get decimals for currency pair: %w", err)
	}

	// the queried price is NTRN/USD, we need to convert it to untrn/USD
	ntrnPrice := math.LegacyNewDecFromIntWithPrec(priceInt.Price, int64(decimals))
	untrnPrice := ntrnPrice.QuoInt64(int64(stdmath.Pow(10, types.RewardDenomDecimals)))

	err = k.CalcNewRewardAssetPrice(ctx, untrnPrice, ctx.BlockTime().Unix())

	if err != nil {
		return fmt.Errorf("failed to save a new reward asset price: %w", err)
	}
	k.Logger(ctx).Debug("TWAP refresh", "price", untrnPrice.String())

	err = k.CleanOutdatedRewardAssetPrices(ctx, ctx.BlockTime().Unix()-params.TwapWindow)
	if err != nil {
		return fmt.Errorf("failed to clean outdated prices: %w", err)
	}

	return nil
}

// CalcNewRewardAssetPrice calculates and saves a new reward asset price. It accepts an absolute
// price and current timestamp as arguments and calculates a cumulative price taking into account
// time passed since the last saved cumulative price.
func (k *Keeper) CalcNewRewardAssetPrice(ctx sdk.Context, price math.LegacyDec, timestamp int64) error {
	rapPrevious, err := k.GetLastRewardAssetPrice(ctx)
	if err != nil {
		return err
	}

	rapNew := types.RewardAssetPrice{
		CumulativePrice: rapPrevious.AbsolutePrice.MulInt64(timestamp - rapPrevious.Timestamp).Add(rapPrevious.CumulativePrice),
		AbsolutePrice:   price,
		Timestamp:       timestamp,
	}

	return k.SaveRewardAssetPrice(ctx, &rapNew)
}

// SaveRewardAssetPrice saves a cumulative price.
func (k *Keeper) SaveRewardAssetPrice(ctx sdk.Context, price *types.RewardAssetPrice) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(price)
	if err != nil {
		return fmt.Errorf("failed to marshal reward asset price: %w", err)
	}

	err = store.Set(types.GetAccumulatedPriceKey(price.Timestamp), bz)
	if err != nil {
		return fmt.Errorf("failed to store reward asset price: %w", err)
	}

	return nil
}

// GetAllRewardAssetPrices get all stored reward price values.
func (k *Keeper) GetAllRewardAssetPrices(ctx sdk.Context) ([]*types.RewardAssetPrice, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.PrefixAccumulatedPriceKey, storetypes.PrefixEndBytes(types.PrefixAccumulatedPriceKey))
	if err != nil {
		return nil, fmt.Errorf("failed to iterate over reward asset prices: %w", err)
	}
	defer iter.Close()

	var prices []*types.RewardAssetPrice
	for ; iter.Valid(); iter.Next() {
		p := types.RewardAssetPrice{}
		if err = k.cdc.Unmarshal(iter.Value(), &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal a reward asset price: %w", err)
		}
		prices = append(prices, &p)
	}
	return prices, nil
}

// GetLastRewardAssetPrice gets the freshest reward asset price stored in the module's state.
func (k *Keeper) GetLastRewardAssetPrice(ctx sdk.Context) (types.RewardAssetPrice, error) {
	cmlt := types.RewardAssetPrice{
		CumulativePrice: math.LegacyZeroDec(),
		AbsolutePrice:   math.LegacyZeroDec(),
		Timestamp:       0,
	}

	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.ReverseIterator(types.PrefixAccumulatedPriceKey, storetypes.PrefixEndBytes(types.PrefixAccumulatedPriceKey))
	if err != nil {
		return cmlt, fmt.Errorf("failed to iterate over reward asset prices: %w", err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if err = k.cdc.Unmarshal(iter.Value(), &cmlt); err != nil {
			return cmlt, fmt.Errorf("failed to unmarshal a reward asset price: %w", err)
		}
		return cmlt, nil //nolint:staticcheck
	}
	k.Logger(ctx).Warn("TWAP storage is empty")
	return cmlt, nil
}

// GetFirstRewardAssetPriceAfter gets the oldest reward asset price stored in the module's state
// stored after a given timestamp.
func (k *Keeper) GetFirstRewardAssetPriceAfter(ctx sdk.Context, startAt int64) (types.RewardAssetPrice, error) {
	cmlt := types.RewardAssetPrice{
		CumulativePrice: math.LegacyZeroDec(),
		AbsolutePrice:   math.LegacyZeroDec(),
		Timestamp:       0,
	}

	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.GetAccumulatedPriceKey(startAt), storetypes.PrefixEndBytes(types.PrefixAccumulatedPriceKey))
	if err != nil {
		return cmlt, fmt.Errorf("failed to iterate over reward asset prices: %w", err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if err = k.cdc.Unmarshal(iter.Value(), &cmlt); err != nil {
			return cmlt, fmt.Errorf("failed to unmarshal a reward asset price: %w", err)
		}
		return cmlt, nil //nolint:staticcheck
	}
	k.Logger(ctx).Warn("TWAP storage is empty")
	return cmlt, nil
}

// CleanOutdatedRewardAssetPrices removes all the reward asset prices those are older than a
// threshold.
func (k *Keeper) CleanOutdatedRewardAssetPrices(ctx sdk.Context, cleanUntil int64) error {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(
		types.PrefixAccumulatedPriceKey,
		types.GetAccumulatedPriceKey(cleanUntil),
	)
	if err != nil {
		return fmt.Errorf("failed to iterate over reward asset prices: %w", err)
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

// GetTWAP returns a TWAP for a whole observation period.
// Observation period is limited with `params.TwapWindow`
func (k *Keeper) GetTWAP(ctx sdk.Context) (math.LegacyDec, error) {
	return k.GetTWAPStartingFromTime(ctx, 0)
}

// GetTWAPStartingFromTime returns a TWAP for window from `startAt` till the last saved value.
func (k *Keeper) GetTWAPStartingFromTime(ctx sdk.Context, startAt int64) (math.LegacyDec, error) {
	lastPrice, err := k.GetLastRewardAssetPrice(ctx)
	if err != nil {
		return math.LegacyZeroDec(), fmt.Errorf("failed to get last reward asset price: %w", err)
	}

	firstPrice, err := k.GetFirstRewardAssetPriceAfter(ctx, startAt)
	if err != nil {
		return math.LegacyZeroDec(), fmt.Errorf("failed to get first first reward asset price: %w", err)
	}
	if lastPrice.Timestamp == firstPrice.Timestamp {
		return lastPrice.AbsolutePrice, nil
	}
	return lastPrice.CumulativePrice.Sub(firstPrice.CumulativePrice).QuoInt64(lastPrice.Timestamp - firstPrice.Timestamp), nil
}
