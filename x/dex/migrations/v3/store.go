package v3

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/types"
	v2types "github.com/neutron-org/neutron/v3/x/dex/types/v2"
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
	var params v2types.Params
	cdc.MustUnmarshal(bz, &params)

	newParams := types.Params{}
	// add new param values
	params.GoodTilPurgeAllowance = types.DefaultGoodTilPurgeAllowance
	params.Max_JITsPerBlock = types.DefaultMaxJITsPerBlock

	// set params
	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set([]byte(types.ParamsKey), bz)

	ctx.Logger().Info("Finished migrating dex params")

	return nil
}
