package keeper

import (
	"context"
	"fmt"
	"sort"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

type (
	Keeper struct {
		storeKey       storetypes.StoreKey
		knownModules   []string
		cdc            codec.Codec
		accountKeeper  types.AccountKeeper
		bankKeeper     types.BankKeeper
		contractKeeper types.ContractKeeper
		authority      string
	}
)

// NewKeeper returns a new instance of the x/tokenfactory keeper
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	maccPerms map[string][]string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	contractKeeper types.ContractKeeper,
	authority string,
) Keeper {
	sortedKnownModules := make([]string, 0, len(maccPerms))
	for moduleName := range maccPerms {
		sortedKnownModules = append(sortedKnownModules, moduleName)
	}
	sort.Strings(sortedKnownModules)

	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		knownModules:   sortedKnownModules,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		contractKeeper: contractKeeper,
		authority:      authority,
	}
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns an authority for the x/tokenfactory module
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetDenomPrefixStore returns the substore for a specific denom
func (k Keeper) GetDenomPrefixStore(ctx context.Context, denom string) storetypes.KVStore {
	c := sdk.UnwrapSDKContext(ctx)

	store := c.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetDenomPrefixStore(denom))
}

// GetCreatorPrefixStore returns the substore for a specific creator address
func (k Keeper) GetCreatorPrefixStore(ctx sdk.Context, creator string) storetypes.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorPrefix(creator))
}

// GetCreatorsPrefixStore returns the substore that contains a list of creators
func (k Keeper) GetCreatorsPrefixStore(ctx sdk.Context) storetypes.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorsPrefix())
}

// Set the wasm keeper.
func (k *Keeper) SetContractKeeper(contractKeeper types.ContractKeeper) {
	k.contractKeeper = contractKeeper
}

// CreateModuleAccount creates a module account with minting and burning capabilities
// This account isn't intended to store any coins,
// it purely mints and burns them on behalf of the admin of respective denoms,
// and sends to the relevant address.
func (k Keeper) CreateModuleAccount(ctx sdk.Context) {
	// GetModuleAccount creates new module account if not present under the hood
	k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
}
