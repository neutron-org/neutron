package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"

	"github.com/neutron-org/neutron/v6/x/revenue/types"
)

// UpdateRewardAssetPrice stores fresh cumulative and absolute price of the reward asset and cleans
// out the cumulative prices ledger from outdated records (that are older than the TWAP window).
func (k *Keeper) UpdateRewardAssetPrice(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module params: %w", err)
	}
	rewardAssetSymbol, err := k.getRewardAssetSymbol(ctx)
	if err != nil {
		return fmt.Errorf("failed to get reward asset symbol: %w", err)
	}

	pair := slinkytypes.CurrencyPair{
		Base:  rewardAssetSymbol,
		Quote: params.RewardQuote.Asset,
	}
	priceInt, err := k.oracleKeeper.GetPriceForCurrencyPair(ctx, pair)
	if err != nil {
		return fmt.Errorf("failed to get price for reward currency pair %s<>%s: %w", pair.Base, pair.Quote, err)
	}
	// safecheck. Should never happen. Make sure the slinky price is valid
	if priceInt.Price.LTE(math.ZeroInt()) {
		return fmt.Errorf("reward asset price %s is invalid", priceInt.Price.String())
	}

	decimals, err := k.oracleKeeper.GetDecimalsForCurrencyPair(ctx, pair)
	if err != nil {
		return fmt.Errorf("failed to get decimals for reward currency pair %s<>%s: %w", pair.Base, pair.Quote, err)
	}

	priceNormalised := math.LegacyNewDecFromIntWithPrec(priceInt.Price, int64(decimals)) //nolint:gosec
	err = k.CalcNewRewardAssetPrice(ctx, priceNormalised, ctx.BlockTime().Unix())
	if err != nil {
		return fmt.Errorf("failed to save a new reward asset price: %w", err)
	}
	k.Logger(ctx).Debug("TWAP refresh", "price", priceNormalised.String())

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
		return math.LegacyZeroDec(), fmt.Errorf("failed to get first reward asset price: %w", err)
	}
	if lastPrice.Timestamp == firstPrice.Timestamp {
		return lastPrice.AbsolutePrice, nil
	}
	return lastPrice.CumulativePrice.Sub(firstPrice.CumulativePrice).QuoInt64(lastPrice.Timestamp - firstPrice.Timestamp), nil
}

// getRewardAssetSymbol retrieves the reward asset symbol from its denom metadata.
func (k *Keeper) getRewardAssetSymbol(ctx sdk.Context) (string, error) {
	rewardAssetMd, err := k.getRewardAssetMetadata(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get reward asset metadata: %w", err)
	}
	return rewardAssetMd.Symbol, nil
}

// getRewardAssetSymbol retrieves the exponent of the reward asset's alias that corresponds to
// reward asset's symbol.
func (k *Keeper) getRewardAssetExponent(ctx sdk.Context) (uint32, error) {
	rewardAssetMd, err := k.getRewardAssetMetadata(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get reward asset metadata: %w", err)
	}

	for _, unit := range rewardAssetMd.DenomUnits {
		for _, alias := range unit.Aliases {
			if alias == rewardAssetMd.Symbol {
				return unit.Exponent, nil
			}
		}
	}
	return 0, fmt.Errorf("couldn't find exponent for reward asset alias %s in reward denom metadata", rewardAssetMd.Symbol)
}

// getRewardAssetMetadata retrieves the reward asset metadata from the bank module.
func (k *Keeper) getRewardAssetMetadata(ctx sdk.Context) (*banktypes.Metadata, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get module params: %w", err)
	}
	rewardAssetMd, ex := k.bankKeeper.GetDenomMetaData(ctx, params.RewardAsset)
	if !ex {
		return nil, fmt.Errorf("reward asset %s metadata doesn't exist", params.RewardAsset)
	}
	return &rewardAssetMd, nil
}
