package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"golang.org/x/exp/slices"

	"github.com/cosmos/cosmos-sdk/runtime"
	"golang.org/x/exp/maps"

	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

type (
	Keeper struct {
		wasmKeeper    types.WasmKeeper
		accountKeeper types.AccountKeeper

		cdc          codec.BinaryCodec
		storeService corestoretypes.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,
	accountKeeper types.AccountKeeper,
	wasmKeeper types.WasmKeeper,
	logger log.Logger,
	authority string,
) *Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return &Keeper{
		cdc:           cdc,
		storeService:  storeService,
		accountKeeper: accountKeeper,
		wasmKeeper:    wasmKeeper,
		authority:     authority,
		logger:        logger,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Hooks returns implemented StakingHooks that will be called by the staking module
func (k Keeper) Hooks() stakingtypes.StakingHooks {
	return Hooks{k}
}

func (k Keeper) SetHookSubscription(goCtx context.Context, hookSubscriptions types.HookSubscriptions) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))
	key := types.GetHookSubscriptionKey(hookSubscriptions.HookType)
	bz := k.cdc.MustMarshal(&hookSubscriptions)
	store.Set(key, bz)
}

// UpdateHookSubscription sets hook subscription for given contractAddress
// All previously subscribed hooks that are not in `subscriptionUpdate.hooks` will be removed.
func (k Keeper) UpdateHookSubscription(goCtx context.Context, update *types.HookSubscription) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))

	allHooks := maps.Keys(types.HookType_name)

	// First we understand which hooks should be added and which should be removed
	hooksToAdd, hooksToRemove := splitAddedAndRemovedHooks(allHooks, update.Hooks)

	// As the contract addresses are stored grouped by hooks, we need to iterate each hook
	// to understand which addresses to remove, which to add, and which are already there.

	for _, item := range hooksToAdd {
		key := types.GetHookSubscriptionKey(item)
		subscriptions := types.HookSubscriptions{
			HookType: types.HookType(item),
		}
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
		if store.Has(key) {
			subscriptions := types.HookSubscriptions{
				HookType: item,
			}
			k.cdc.MustUnmarshal(store.Get(key), &subscriptions)

			// remove contract address if it's present in the list
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

// GetSubscribedAddressesForHookType returns all subscribed contracts for a given `hookType`
func (k Keeper) GetSubscribedAddressesForHookType(goCtx context.Context, hookType types.HookType) []string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))

	key := types.GetHookSubscriptionKey(hookType)
	if store.Has(key) {
		subscriptions := types.HookSubscriptions{}
		k.cdc.MustUnmarshal(store.Get(key), &subscriptions)
		return subscriptions.ContractAddresses
	} else {
		return []string{}
	}
}

// GetAllSubscriptions returns subscriptions for all hooks
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
func splitAddedAndRemovedHooks(allHooks []int32, hooksToAdd []types.HookType) ([]types.HookType, []types.HookType) {
	// Calculate difference between allHooks and hooksToAdd. It will be hooksToRemove.
	var hooksToRemove []types.HookType
	itemMap := make(map[int32]bool)

	// Add all items from hooksToAdd to the map
	for _, item := range hooksToAdd {
		itemMap[int32(item)] = true
	}

	// Check each item in allHooks and add to hooksToRemove if not in the map
	for _, item := range allHooks {
		if !itemMap[item] {
			hooksToRemove = append(hooksToRemove, types.HookType(item))
		}
	}

	return hooksToAdd, hooksToRemove
}
