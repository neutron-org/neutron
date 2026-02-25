package v7

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v10/utils/math"
	"github.com/neutron-org/neutron/v10/x/dex/types"
)

// MigrateStore performs in-place store migrations.
// Add DecSharesWithdrawn field to all LimitOrderTrancheUser
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	if err := migrateLimitOrderTrancheUsers(ctx, cdc, storeKey); err != nil {
		return err
	}

	return nil
}

type migrationUpdate struct {
	key []byte
	val []byte
}

func migrateLimitOrderTrancheUsers(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating LimitOrderTrancheUser fields...")

	// Iterate through all LimitOrderTrancheUser
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.LimitOrderTrancheUserKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	trancheUserUpdates := make([]migrationUpdate, 0)

	for ; iterator.Valid(); iterator.Next() {
		var limitOrderTrancheUser types.LimitOrderTrancheUser
		cdc.MustUnmarshal(iterator.Value(), &limitOrderTrancheUser)

		limitOrderTrancheUser.DecSharesWithdrawn = math_utils.NewPrecDecFromInt(limitOrderTrancheUser.SharesWithdrawn)

		bz := cdc.MustMarshal(&limitOrderTrancheUser)
		trancheUserUpdates = append(trancheUserUpdates, migrationUpdate{key: iterator.Key(), val: bz})

	}

	err := iterator.Close()
	if err != nil {
		return errorsmod.Wrap(err, "iterator failed to close during migration")
	}

	// Store the updated LimitOrderTrancheUser
	for _, v := range trancheUserUpdates {
		store.Set(v.key, v.val)
	}

	ctx.Logger().Info("Finished migrating LimitOrderTrancheUser fields...")

	return nil
}
