package v8

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v10/x/dex/types"
	v7types "github.com/neutron-org/neutron/v10/x/dex/types/v7"
)

// MigrateStore performs in-place store migrations.
// The migration adds new dex params
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	if err := migrateParams(ctx, cdc, storeKey); err != nil {
		return err
	}
	return nil
}

func migrateParams(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating dex params...")

	// fetch old params
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.KeyPrefix(types.ParamsKey))
	if bz == nil {
		return errors.New("cannot fetch dex params from KV store")
	}
	var oldParams v7types.Params
	cdc.MustUnmarshal(bz, &oldParams)

	// add new param values
	newParams := types.Params{
		Paused:                types.DefaultPaused,
		FeeTiers:              oldParams.FeeTiers,
		GoodTilPurgeAllowance: types.DefaultGoodTilPurgeAllowance,
		MaxJitsPerBlock:       types.DefaultMaxJITsPerBlock,
		WithdrawOnly:          true,
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
