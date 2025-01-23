package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"cosmossdk.io/errors"
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/maps"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"
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
func (k Keeper) UpdateHookSubscription(goCtx context.Context, update *types.HookSubscription) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))

	allHooks := maps.Values(types.HookType_name)

	// First we understand which hooks should be added and which should be removed
	hooksToAdd, hooksToRemove := splitAddedAndRemovedHooks(allHooks, update.Hooks)

	// As the contract addresses are stored grouped by hooks, we need to iterate each hook
	// to understand which addresses to remove, which to add, and which are already there.

	for _, item := range hooksToAdd {
		key := types.GetHookSubscriptionKey(item)
		subscriptions := types.HookSubscriptions{}
		if store.Has(key) {
			k.cdc.MustUnmarshal(store.Get(key), &subscriptions)
		}

		// add contract address if it's not already in the list
		if !slices.Contains(subscriptions.ContractAddresses, update.ContractAddress) {
			subscriptions.ContractAddresses = append(subscriptions.ContractAddresses, update.ContractAddress)
		}

		bz := k.cdc.MustMarshal(&subscriptions)
		store.Set(key, bz)
	}

	for _, item := range hooksToRemove {
		key := types.GetHookSubscriptionKey(item)
		subscriptions := types.HookSubscriptions{}
		if store.Has(key) {
			k.cdc.MustUnmarshal(store.Get(key), &subscriptions)
		}

		// remove contract address if it's present in the list
		newContractAddresses := slices.DeleteFunc(subscriptions.ContractAddresses, func(addr string) bool {
			return addr == update.ContractAddress
		})

		subscriptions.ContractAddresses = newContractAddresses
		bz := k.cdc.MustMarshal(&subscriptions)
		store.Set(key, bz)
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
	// Convert subscribedHooks to string namings
	hooksToAdd := make([]string, len(subscribedHooks))
	for _, item := range subscribedHooks {
		hooksToAdd = append(hooksToAdd, types.HookType_name[int32(item)])
	}

	// Calculate difference between allHooks and hooksToAdd. It will be hooksToRemove.
	var hooksToRemove []string
	itemMap := make(map[string]bool)

	// Add all items from hooksToAdd to the map
	for _, item := range hooksToAdd {
		itemMap[item] = true
	}

	// Check each item in allHooks and add to hooksToRemove if not in the map
	for _, item := range allHooks {
		if !itemMap[item] {
			hooksToRemove = append(hooksToRemove, item)
		}
	}

	return hooksToAdd, hooksToRemove
}
