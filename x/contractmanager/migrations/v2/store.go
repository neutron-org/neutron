package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/contractmanager/types"
	v1 "github.com/neutron-org/neutron/x/contractmanager/types/v1"
)

// MigrateStore performs in-place store migrations.
// The migration rearranges and adds new fields to the failures
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	if err := migrateFailures(ctx, storeKey, cdc); err != nil {
		return err
	}

	return nil
}

func migrateFailures(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	ctx.Logger().Info("Migrating failures...")

	// fetch list of all old failures
	oldFailuresList := make([]v1.Failure, 0)
	iteratorStore := prefix.NewStore(ctx.KVStore(storeKey), types.ContractFailuresKey)
	iterator := sdk.KVStorePrefixIterator(iteratorStore, []byte{})

	for ; iterator.Valid(); iterator.Next() {
		var val v1.Failure
		cdc.MustUnmarshal(iterator.Value(), &val)
		oldFailuresList = append(oldFailuresList, val)
	}

	err := iterator.Close()
	if err != nil {
		return err
	}

	// migrate
	store := ctx.KVStore(storeKey)
	for _, oldItem := range oldFailuresList {
		failure := types.Failure{
			Address: oldItem.Address,
			Id:      oldItem.Id,
			AckType: oldItem.AckType,
			Packet:  nil,
			Ack:     nil,
		}
		bz := cdc.MustMarshal(&failure)
		store.Set(types.GetFailureKey(failure.Address, failure.Id), bz)
	}

	ctx.Logger().Info("Finished migrating failures")

	return nil
}
