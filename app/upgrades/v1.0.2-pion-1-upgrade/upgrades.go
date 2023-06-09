package v102_pion

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/interchain-security/x/ccv/consumer/types"
	ccvtypes "github.com/cosmos/interchain-security/x/ccv/types"

	"github.com/neutron-org/neutron/app/upgrades"
)

type OldValidator struct {
	Key   []byte
	Value []byte
}

const OldCrossChainValidatorBytePrefix = 15

func MigratePendingDataPackets(ctx sdk.Context, consumerStoreKey store.Key) {
	store := ctx.KVStore(consumerStoreKey)

	oldData := store.Get([]byte{types.PendingDataPacketsByteKey})
	if len(oldData) == 0 {
		return
	}

	store.Set([]byte{types.PendingDataPacketsBytePrefix}, oldData)
	store.Delete([]byte{types.PendingDataPacketsByteKey}) // remove old data
}

// OldGetAllCCValidator reads validators under old keys
func OldGetAllCCValidator(ctx sdk.Context, consumerStoreKey store.Key) (validators []OldValidator) {
	store := ctx.KVStore(consumerStoreKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{OldCrossChainValidatorBytePrefix})

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		validators = append(validators, OldValidator{
			Key:   iterator.Key()[1:], // remove prefix from a key
			Value: iterator.Value(),
		})
	}

	// remove corrupted data
	for _, validator := range validators {
		store.Delete(append([]byte{OldCrossChainValidatorBytePrefix}, validator.Key...))
	}

	return validators
}

// RemoveOutdatedValidators removes outdated validators from corrupted storage
func RemoveOutdatedValidators(ctx sdk.Context, consumerStoreKey store.Key) {
	store := ctx.KVStore(consumerStoreKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{types.CrossChainValidatorBytePrefix})
	defer iterator.Close()

	var keys [][]byte
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, iterator.Key())
	}

	// remove corrupted data
	for _, key := range keys {
		store.Delete(key)
	}
}

// SetCCValidator saves a validator under a new proper key
func SetCCValidator(ctx sdk.Context, storeKey store.Key, key, bz []byte) {
	store := ctx.KVStore(storeKey)

	store.Set(types.CrossChainValidatorKey(key), bz)
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Migrating CrossChainValidators...")
		RemoveOutdatedValidators(ctx, keepers.ConsumerStoreKey)

		oldValidators := OldGetAllCCValidator(ctx, keepers.ConsumerStoreKey)
		for _, oldValidator := range oldValidators {
			SetCCValidator(ctx, keepers.ConsumerStoreKey, oldValidator.Key, oldValidator.Value)
		}

		ctx.Logger().Info("Migrating PendingDataPackets...")
		MigratePendingDataPackets(ctx, keepers.ConsumerStoreKey)

		ctx.Logger().Info("Migrating CCVConsumerKeeper Params...")

		paramsSubspace := keepers.ConsumerParamsSubspace

		var enabled bool
		paramsSubspace.Get(ctx, types.KeyEnabled, &enabled)
		var blocksPerDistributionTransmission int64
		paramsSubspace.Get(ctx, types.KeyBlocksPerDistributionTransmission, &blocksPerDistributionTransmission)
		var distributionTransmissionChannel string
		paramsSubspace.Get(ctx, types.KeyDistributionTransmissionChannel, &distributionTransmissionChannel)
		var providerFeePoolAddrStr string
		paramsSubspace.Get(ctx, types.KeyProviderFeePoolAddrStr, &providerFeePoolAddrStr)
		var ccvTimeoutPeriod time.Duration
		paramsSubspace.Get(ctx, ccvtypes.KeyCCVTimeoutPeriod, &ccvTimeoutPeriod)
		var transferTimeoutPeriod time.Duration
		paramsSubspace.Get(ctx, types.KeyTransferTimeoutPeriod, &transferTimeoutPeriod)
		var consumerRedistributionFrac string
		paramsSubspace.Get(ctx, types.KeyConsumerRedistributionFrac, &consumerRedistributionFrac)
		var historicalEntries int64
		paramsSubspace.Get(ctx, types.KeyHistoricalEntries, &historicalEntries)
		var unbondingPeriod time.Duration
		paramsSubspace.Get(ctx, types.KeyConsumerUnbondingPeriod, &unbondingPeriod)
		var softOptOutThreshold string
		paramsSubspace.Get(ctx, types.KeySoftOptOutThreshold, &softOptOutThreshold)

		// Recycle old params, set new params to default values
		newParams := types.NewParams(
			enabled,
			blocksPerDistributionTransmission,
			distributionTransmissionChannel,
			"", // because we will get an error about wrong prefix, we'll set it later
			ccvTimeoutPeriod,
			transferTimeoutPeriod,
			consumerRedistributionFrac,
			historicalEntries,
			unbondingPeriod,
			softOptOutThreshold,
			[]string{"untrn"},
			[]string{"uatom"},
		)

		// Persist new params
		paramsSubspace.SetParamSet(ctx, &newParams)

		keepers.ConsumerKeeper.SetProviderFeePoolAddrStr(ctx, providerFeePoolAddrStr)

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
