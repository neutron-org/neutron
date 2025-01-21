package keeper

import (
	"context"
	"cosmossdk.io/errors"
	errorsmod "cosmossdk.io/errors"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
	"golang.org/x/exp/maps"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) ManageHookSubscription(goCtx context.Context, req *types.MsgManageHookSubscription) (*types.MsgManageHookSubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errorsmod.Wrapf(err, "failed to validate manage hook subscription message")
	}

	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.UpdateHookSubscription(ctx, req.HookSubscription)

	return &types.MsgManageHookSubscriptionResponse{}, nil
}

// UpdateHookSubscription sets hook subscription for given contractAddress
// All previously subscribed hooks that are not in `subscriptionUpdate.hooks` will be removed.
func (k Keeper) UpdateHookSubscription(goCtx context.Context, subscriptionUpdate *types.HookSubscription) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))

	allHooks := maps.Values(types.HookType_name)

	hooksAdded, hooksRemoved := splitAddedAndRemovedHooks(allHooks, subscriptionUpdate.Hooks)

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
				break
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

// CallSudoForSubscriptionType calls sudo for all contracts subscribed to given `hookType`.
// Returns error in cases where marshalling error occurred (should never happen, since we control it) or
// when any error in contract happened.
// Important that because some calls are coming from BeginBlocker/EndBlocker, any errors in contracts can halt the chain.
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
		// As we're using ctx here (no cached context!), any errors will be returned from hooks as it is.
		// This means it can potentially halt the chain OR abort the transactions depending on where hook was called from.
		accContractAddress, err := sdk.AccAddressFromBech32(contractAddress)
		if err != nil {
			return errors.Wrapf(err, "could not parse acc address from bech32 for harpoon sudo call, contract_address=%s", contractAddress)
		}
		_, err = k.wasmKeeper.Sudo(sdkCtx, accContractAddress, msgJsonBz)
		if err != nil {
			return errors.Wrapf(err, "could not execute sudo call successfully for hook_type=%s, msg=%s, contract_address=%s", hookType.String(), string(msgJsonBz), contractAddress)
		}
	}

	return nil
}

// splitAddedAndRemovedHooks splits all hooks on which ones to add and which ones to remove.
func splitAddedAndRemovedHooks(allHooks []string, subscribedHooks []types.HookType) ([]string, []string) {
	// convert subscribedHooks to string namings
	slice2 := make([]string, len(subscribedHooks))
	for _, item := range subscribedHooks {
		slice2 = append(slice2, types.HookType_name[int32(item)])
	}

	// splitAddedAndRemovedHooks
	var slice3 []string
	itemMap := make(map[string]bool)

	// Add all items from slice2 to the map
	for _, item := range slice2 {
		itemMap[item] = true
	}

	// Check each item in allHooks and add to slice3 if not in the map
	for _, item := range allHooks {
		if !itemMap[item] {
			slice3 = append(slice3, item)
		}
	}

	return slice2, slice3
}
