package v2

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v8/x/feerefunder/types"
)

// MigrateStore performs in-place store migrations.
// The migration adds execution stage for schedules.
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	return migrateParams(ctx, cdc, storeKey)
}

func migrateParams(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating feerefunder params...")

	store := ctx.KVStore(storeKey)

	paramsBz := store.Get(types.ParamsKey)
	var params types.Params
	// unmarshalling old type as new type, but it should be fine since it's same layout
	err := params.Unmarshal(paramsBz)
	if err != nil {
		return err
	}
	params.FeeEnabled = true

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)

	ctx.Logger().Info("Feerefunder params migrated")

	return nil
}
