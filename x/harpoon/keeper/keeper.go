package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"golang.org/x/exp/maps"

	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/neutron-org/neutron/v6/x/harpoon/types"
)

type (
	Keeper struct {
		wasmKeeper types.WasmKeeper

		cdc          codec.BinaryCodec
		storeService corestoretypes.KVStoreService
		logger       log.Logger

		authority string
	}
)

func (k *Keeper) GetWasmKeeper() types.WasmKeeper {
	return k.wasmKeeper
}

// NewKeeper creates a new keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,
	wasmKeeper types.WasmKeeper,
	logger log.Logger,
	authority string,
) *Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return &Keeper{
		cdc:          cdc,
		storeService: storeService,
		wasmKeeper:   wasmKeeper,
		authority:    authority,
		logger:       logger,
	}
}

// GetAuthority returns the authority of the module.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns the logger specific to the module.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Hooks returns the implemented StakingHooks to be called by the staking module.
func (k Keeper) Hooks() stakingtypes.StakingHooksBeforeValidatorSlashedHasTokensToBurn {
	return Hooks{&k}
}

// SetHookSubscription configures hook subscriptions for the specified hook type.
func (k Keeper) SetHookSubscription(goCtx context.Context, hookSubscriptions types.HookSubscriptions) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))
	key := types.GetHookSubscriptionKey(hookSubscriptions.HookType)
	bz := k.cdc.MustMarshal(&hookSubscriptions)
	store.Set(key, bz)
}

// UpdateHookSubscription updates the hook subscription for the given contractAddress.
// Previously subscribed hooks not listed in `subscriptionUpdate.hooks` will be removed.
func (k Keeper) UpdateHookSubscription(goCtx context.Context, update *types.HookSubscription) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))

	allHooks := maps.Keys(types.HookType_name)
	// Sort to avoid any potential non-determinism of `maps.Keys`
	slices.Sort(allHooks)

	// First we understand which hooks should be removed
	hooksToRemove := findHooksToRemove(allHooks, update.Hooks)

	// Since contract addresses are stored grouped by hooks, we need to iterate through each hook
	// to determine which addresses to add, remove, or retain.

	for _, item := range update.Hooks {
		key := types.GetHookSubscriptionKey(item)
		subscriptions := types.HookSubscriptions{
			HookType: item,
		}
		if store.Has(key) {
			k.cdc.MustUnmarshal(store.Get(key), &subscriptions)
		}

		// Add contract address if it's not already in the list
		if !slices.Contains(subscriptions.ContractAddresses, update.ContractAddress) {
			subscriptions.ContractAddresses = append(subscriptions.ContractAddresses, update.ContractAddress)
		}

		bz := k.cdc.MustMarshal(&subscriptions)
		store.Set(key, bz)
	}

	for _, item := range hooksToRemove {
		key := types.GetHookSubscriptionKey(item)
		if store.Has(key) {
			subscriptions := types.HookSubscriptions{}
			k.cdc.MustUnmarshal(store.Get(key), &subscriptions)

			// Remove contract address if it's present in the list
			newContractAddresses := slices.DeleteFunc(subscriptions.ContractAddresses, func(addr string) bool {
				return addr == update.ContractAddress
			})

			subscriptions.ContractAddresses = newContractAddresses

			if len(subscriptions.ContractAddresses) == 0 {
				store.Delete(key)
			} else {
				bz := k.cdc.MustMarshal(&subscriptions)
				store.Set(key, bz)
			}
		}
	}
}

// GetSubscribedAddressesForHookType retrieves all contracts subscribed to the specified `hookType`.
func (k Keeper) GetSubscribedAddressesForHookType(goCtx context.Context, hookType types.HookType) []string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))

	key := types.GetHookSubscriptionKey(hookType)
	if store.Has(key) {
		subscriptions := types.HookSubscriptions{}
		k.cdc.MustUnmarshal(store.Get(key), &subscriptions)
		return subscriptions.ContractAddresses
	}

	return []string{}
}

// GetAllSubscriptions retrieves subscriptions for all hooks.
func (k Keeper) GetAllSubscriptions(goCtx context.Context) (res []types.HookSubscriptions) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx)), types.GetHookSubscriptionKeyPrefix())

	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var s types.HookSubscriptions
		k.cdc.MustUnmarshal(iterator.Value(), &s)

		res = append(res, s)
	}

	return res
}

// CallSudoForSubscriptionType executes the sudo method for all contracts subscribed to the specified `hookType`.
// Returns an error if a marshalling issue occurs (unlikely, as it's controlled) or if any contract-related error occurs.
// Note: Errors in contracts, especially from BeginBlocker/EndBlocker calls, can halt the chain.
func (k Keeper) CallSudoForSubscriptionType(ctx context.Context, hookType types.HookType, msg any) error {
	if err := k.doCallSudoForSubscriptionType(ctx, hookType, msg); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", hookType)
	}

	return nil
}

func (k Keeper) doCallSudoForSubscriptionType(ctx context.Context, hookType types.HookType, msg any) error {
	contractAddresses := k.GetSubscribedAddressesForHookType(ctx, hookType)

	if len(contractAddresses) == 0 {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	msgJSONBz, err := json.Marshal(msg)
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
		_, err = k.wasmKeeper.Sudo(sdkCtx, accContractAddress, msgJSONBz)
		if err != nil {
			return errors.Wrapf(err, "could not execute sudo call successfully for hook_type=%s, msg=%s, contract_address=%s", hookType.String(), string(msgJSONBz), contractAddress)
		}
	}

	return nil
}

// findHooksToRemove finds what hooks are to be removed.
func findHooksToRemove(allHooks []int32, hooksToAdd []types.HookType) []types.HookType {
	var res []types.HookType
	// Calculate difference between allHooks and hooksToAdd.
	itemMap := make(map[int32]bool)

	// Add all items from hooksToAdd to the map
	for _, item := range hooksToAdd {
		itemMap[int32(item)] = true
	}

	// Check each item in allHooks and add to hooksToRemove if not in the map
	for _, item := range allHooks {
		if !itemMap[item] {
			res = append(res, types.HookType(item))
		}
	}

	return res
}
