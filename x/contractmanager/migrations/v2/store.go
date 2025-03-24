package v2

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/contractmanager/types"
)

// MigrateStore performs in-place store migrations.
// The migration rearranges removes all old failures,
// since they do not have the necessary fields packet and ack for resubmission
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	return migrateFailures(ctx, storeKey)
}

func migrateFailures(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating failures...")

	// fetch list of all old failure keys
	failureKeys := make([][]byte, 0)
	iteratorStore := prefix.NewStore(ctx.KVStore(storeKey), types.ContractFailuresKey)
	iterator := storetypes.KVStorePrefixIterator(iteratorStore, []byte{})

	for ; iterator.Valid(); iterator.Next() {
		failureKeys = append(failureKeys, iterator.Key())
	}

	err := iterator.Close()
	if err != nil {
		return err
	}

	// remove failures
	store := prefix.NewStore(ctx.KVStore(storeKey), types.ContractFailuresKey)
	for _, key := range failureKeys {
		store.Delete(key)
	}

	ctx.Logger().Info("Finished migrating failures")

	return nil
}
