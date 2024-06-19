package v2

import (
	"errors"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v4/x/tokenfactory/types"
	v1beta1types "github.com/neutron-org/neutron/v4/x/tokenfactory/types/v1beta1"
)

// MigrateStore performs in-place store migrations.
// The migration adds the new tokenfactory params WhitelistedHooks
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	return migrateParams(ctx, cdc, storeKey)

}

var WhitelistedHooks = []*types.HookWhitelist{

	{ // xASTRO balances tracker
		CodeID:       944,
		DenomCreator: "neutron1zlf3hutsa4qnmue53lz2tfxrutp8y2e3rj4nkghg3rupgl4mqy8s5jgxsn",
	},
	{ // NFA.zoneV1
		CodeID:       1265,
		DenomCreator: "neutron1pwjn3tsumm3j7v7clzqhjsaukv4tdjlclhdytawhet68fwlz84fqcrdyf5",
	},
}

func migrateParams(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating tokenfactory params...")

	// fetch old params
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return errors.New("cannot fetch tokenfactory params from KV store")
	}
	var oldParams v1beta1types.Params
	cdc.MustUnmarshal(bz, &oldParams)

	// add new param values
	newParams := types.Params{
		DenomCreationFee:        oldParams.DenomCreationFee,
		DenomCreationGasConsume: oldParams.DenomCreationGasConsume,
		FeeCollectorAddress:     oldParams.FeeCollectorAddress,
		WhitelistedHooks:        WhitelistedHooks,
	}

	// set params
	bz, err := cdc.Marshal(&newParams)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)

	ctx.Logger().Info("Finished migrating tokenfactory params")

	return nil
}
