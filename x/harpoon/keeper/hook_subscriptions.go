package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
	"golang.org/x/exp/maps"
)

// UpdateHookSubscription updates hook subscription for given contractAddress, removes it if `Hooks` passed were empty
func (k Keeper) UpdateHookSubscription(goCtx context.Context, subscriptionUpdate *types.HookSubscription) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))

	allHooks := maps.Values(types.HookType_name)

	hooksAdded, hooksRemoved := diff(allHooks, subscriptionUpdate.Hooks)

	for _, toAdd := range hooksAdded {
		key := types.GetHookSubscriptionKey(toAdd)
		subscriptions := types.HookSubscriptions{}
		if store.Has(key) {
			k.cdc.MustUnmarshal(store.Get(key), &subscriptions)
		}

		hasAddress := false
		for _, address := range subscriptions.ContractAddresses {
			if address == subscriptionUpdate.ContractAddress {
				hasAddress = true
			}
		}

		if !hasAddress {
			subscriptions.ContractAddresses = append(subscriptions.ContractAddresses, subscriptionUpdate.ContractAddress)
		}

		bz := k.cdc.MustMarshal(&subscriptions)
		store.Set(key, bz)
	}

	for _, toRemove := range hooksRemoved {
		key := types.GetHookSubscriptionKey(toRemove)
		subscriptions := types.HookSubscriptions{}
		if store.Has(key) {
			k.cdc.MustUnmarshal(store.Get(key), &subscriptions)
		}

		newContractAddresses := make([]string, 0)
		for _, address := range subscriptions.ContractAddresses {
			if address != subscriptionUpdate.ContractAddress {
				newContractAddresses = append(newContractAddresses, address)
			}
		}

		if len(newContractAddresses) == 0 {
			store.Delete(key)
		} else {
			subscriptions.ContractAddresses = newContractAddresses
			bz := k.cdc.MustMarshal(&subscriptions)
			store.Set(key, bz)
		}
	}

	return nil
}

//func (k Keeper) GetAllHookSubscriptions(ctx sdk.Context) []types.HookSubscription {
//	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
//	res := make([]types.HookSubscription, 0)
//
//	iterator := storetypes.KVStorePrefixIterator(store, types.GetHookSubscriptionKeyPrefix())
//	defer iterator.Close()
//
//	for ; iterator.Valid(); iterator.Next() {
//		var subscription types.HookSubscription
//		k.cdc.MustUnmarshal(iterator.Value(), &subscription)
//		res = append(res, subscription)
//	}
//	return res
//}

func (k Keeper) GetSubscribedAddressesForHookType(goCtx context.Context, hookType types.HookType) []string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))

	key := types.GetHookSubscriptionKey(hookType.String())
	if store.Has(key) {
		subscriptions := types.HookSubscriptions{}
		k.cdc.MustUnmarshal(store.Get(key), &subscriptions)
		return subscriptions.ContractAddresses
	} else {
		return []string{}
	}
}

func (k Keeper) CallSudoForSubscriptions(ctx context.Context, contractAddresses []string, msg any) error {
	if len(contractAddresses) == 0 {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	msgJsonBz, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal msg: %v", err)
	}

	// TODO: probably use custom gas meter here
	cacheCtx, writeFn := sdkCtx.CacheContext()

	// TODO: decide how errors are handled
	// TODO: decide how gas is handled: what is limit?
	for _, contractAddress := range contractAddresses {
		executeMsg := wasmtypes.MsgExecuteContract{
			Sender:   k.accountKeeper.GetModuleAddress(types.ModuleName).String(),
			Contract: contractAddress,
			Msg:      msgJsonBz,
			Funds:    sdk.NewCoins(),
		}
		_, err := k.WasmMsgServer.ExecuteContract(cacheCtx, &executeMsg)
		if err != nil {
			sdkCtx.Logger().Info("executeSchedule: failed to execute contract msg",
				"contract_address", contractAddress,
				"error", err,
			)
			// TODO: check that correct behaviour
			continue
		}
	}

	// only save state if all the messages in a schedule were executed successfully
	writeFn()
	return nil
}

// TODO: check
// diff calculates difference between set slice1 and slice2, returns (slice2 converted to []string, slicesDifference)
func diff(slice1 []string, slice2HookType []types.HookType) ([]string, []string) {
	// convert slice2HookType to string namings
	slice2 := make([]string, len(slice2HookType))
	for _, item := range slice2HookType {
		slice2 = append(slice2, types.HookType_name[int32(item)])
	}

	// diff
	var slice3 []string
	itemMap := make(map[string]bool)

	// Add all items from slice2 to the map
	for _, item := range slice2 {
		itemMap[item] = true
	}

	// Check each item in slice1 and add to slice3 if not in the map
	for _, item := range slice1 {
		if !itemMap[item] {
			slice3 = append(slice3, item)
		}
	}

	return slice2, slice3
}
