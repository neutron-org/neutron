package v3

import (
	"errors"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dex/types"
	v2types "github.com/neutron-org/neutron/v4/x/dex/types/v2"
)

// MigrateStore performs in-place store migrations.
// The migration adds new dex params -- GoodTilPurgeAllowance & MaxJITsPerBlock// for handling JIT orders.
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	return migrateParams(ctx, cdc, storeKey)
}

func migrateParams(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating dex params...")

	// fetch old params
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.KeyPrefix(types.ParamsKey))
	if bz == nil {
		return errors.New("cannot fetch dex params from KV store")
	}
	var oldParams v2types.Params
	cdc.MustUnmarshal(bz, &oldParams)

	// add new param values
	newParams := types.Params{
		GoodTilPurgeAllowance: types.DefaultGoodTilPurgeAllowance,
		Max_JITsPerBlock:      types.DefaultMaxJITsPerBlock,
		FeeTiers:              oldParams.FeeTiers,
		MaxTrueTakerSpread:    oldParams.MaxTrueTakerSpread,
	}

	// set params
	bz, err := cdc.Marshal(&newParams)
	if err != nil {
		return err
	}
	store.Set(types.KeyPrefix(types.ParamsKey), bz)

	ctx.Logger().Info("Finished migrating dex params")

	return nil
}
