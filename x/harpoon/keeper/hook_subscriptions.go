package keeper

import (
	"context"
	"cosmossdk.io/errors"
	"encoding/json"
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
	"golang.org/x/exp/maps"
)

// UpdateHookSubscription sets hook subscription for given contractAddress
// All previously subscribed hooks that are not in `subscriptionUpdate.hooks` will be removed.
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

// CallSudoForSubscriptionType calls sudo for all contracts subscribed to given `hookType`.
// Returns error in cases where marshalling error occurred (should never happen, since we control it) or
// when out of gas on sudo call. That can happen when callback triggered during transaction (not BeginBlocker/EndBlocker).
// Ignores sudo contract errors.
func (k Keeper) CallSudoForSubscriptionType(ctx context.Context, hookType types.HookType, msg any) error {
	if err := k.DoCallSudoForSubscriptionType(ctx, hookType, msg); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", hookType)
	}

	return nil
}

func (k Keeper) DoCallSudoForSubscriptionType(ctx context.Context, hookType types.HookType, msg any) error {
	contractAddresses := k.GetSubscribedAddressesForHookType(ctx, hookType)

	if len(contractAddresses) == 0 {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	msgJsonBz, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal sudo subscription msg: %v", err)
	}

	for _, contractAddress := range contractAddresses {
		executeMsg := wasmtypes.MsgExecuteContract{
			Sender:   k.accountKeeper.GetModuleAddress(types.ModuleName).String(),
			Contract: contractAddress,
			Msg:      msgJsonBz,
			Funds:    sdk.NewCoins(),
		}
		// NOTE: as we're using sdkCtx here, all hooks that are triggered by Tx user actions such as Delegate.
		// will consume gas for executing this contract.
		// This also means it can breach gas limit if call was too heavy.
		// And if it happens, we may return prematurely to avoid more computations.
		// BUT: We also want to continue on other errors, since simple contract hook errors should not stop execution.
		// For EndBlocker, there is no counting gas.
		_, err := k.WasmMsgServer.ExecuteContract(sdkCtx, &executeMsg)
		if sdkCtx.GasMeter().IsPastLimit() {
			return errors.Wrapf(sdkerrors.ErrOutOfGas, "not enough gas when executed sudo contract: %v", err)
		}
		if err != nil {
			sdkCtx.Logger().Error("execute harpoon subscription hook error: failed to execute contract msg",
				"contract_address", contractAddress,
				"error", err,
			)
			continue
		}
	}

	return nil
}

// GetSubscribedAddressesForHookType returns all subscribed contracts for a given `hookType`
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
