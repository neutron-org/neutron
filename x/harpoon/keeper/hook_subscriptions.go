package keeper

import (
	"context"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

// UpdateHookSubscription updates hook subscription for given contractAddress, removes it if `Hooks` passed were empty
func (k Keeper) UpdateHookSubscription(goCtx context.Context, subscription *types.HookSubscription) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))
	key := types.GetHookSubscriptionKey(subscription.ContractAddress)

	// remove if empty Hooks list passed
	if len(subscription.Hooks) == 0 {
		store.Delete(key)
	} else {
		// update if non-empty Hooks list passed
		bz, err := k.cdc.Marshal(subscription)
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}

	return nil
}

func (k Keeper) GetAllHookSubscriptions(ctx sdk.Context) []types.HookSubscription {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	res := make([]types.HookSubscription, 0)

	iterator := storetypes.KVStorePrefixIterator(store, types.GetHookSubscriptionKeyPrefix())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var subscription types.HookSubscription
		// TODO: test that it works
		k.cdc.MustUnmarshal(iterator.Value(), &subscription)
		res = append(res, subscription)
	}
	return res
}

// TODO: more efficient
func (k Keeper) GetSubscribedAddressesForHookType(goCtx sdk.Context, hookType types.HookType) []string {
	// TODO: implement efficiently
	return nil
}

func (k Keeper) CallSudoForSubscription() error {

}
