package v2

import (
	"errors"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v9/x/coinfactory/types"
)

// MigrateStore performs in-place store migrations.
// The migration sets the new coinfactory params TrackBeforeSendGasLimit to default value.
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	if err := migrateParams(ctx, cdc, storeKey); err != nil {
		return err
	}
	return nil
}

func migrateParams(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating coinfactory params...")

	// fetch old params
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return errors.New("cannot fetch coinfactory params from KV store")
	}
	var oldParams types.Params
	cdc.MustUnmarshal(bz, &oldParams)

	// add new param values
	newParams := types.Params{
		DenomCreationFee:        oldParams.DenomCreationFee,
		DenomCreationGasConsume: oldParams.DenomCreationGasConsume,
		FeeCollectorAddress:     oldParams.FeeCollectorAddress,
		WhitelistedHooks:        oldParams.WhitelistedHooks,
		TrackBeforeSendGasLimit: types.DefaultTrackBeforeSendGasLimit,
	}

	// set params
	bz, err := cdc.Marshal(&newParams)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)

	ctx.Logger().Info("Finished migrating coinfactory params")

	return nil
}
