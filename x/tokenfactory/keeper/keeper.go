package keeper

import (
	"fmt"
	"sort"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/tokenfactory/types"
)

type (
	Keeper struct {
		storeKey       storetypes.StoreKey
		permAddrs      map[string]authtypes.PermissionsForAddress
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
	permAddrs := make(map[string]authtypes.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = authtypes.NewPermissionsForAddress(name, perms)
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		permAddrs:      permAddrs,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		contractKeeper: contractKeeper,
		authority:      authority,
	}
}

// BlockedAddr checks if a given address is restricted
func (k Keeper) BlockedAddr(ctx sdk.Context, addr sdk.AccAddress) bool {
	sortedPermAddrs := make([]string, 0, len(k.permAddrs))
	for moduleName := range k.permAddrs {
		sortedPermAddrs = append(sortedPermAddrs, moduleName)
	}
	sort.Strings(sortedPermAddrs)

	for _, moduleName := range sortedPermAddrs {
		// tokenfactory itself is allowed to receive and send funds, otherwise Mint/Burn don't work
		if moduleName == types.ModuleName {
			continue
		}

		account := k.accountKeeper.GetModuleAccount(ctx, moduleName)
		if account == nil {
			return true
		}

		if account.GetAddress().Equals(addr) {
			return true
		}
	}
	return false
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
func (k Keeper) GetDenomPrefixStore(ctx sdk.Context, denom string) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetDenomPrefixStore(denom))
}

// GetCreatorPrefixStore returns the substore for a specific creator address
func (k Keeper) GetCreatorPrefixStore(ctx sdk.Context, creator string) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorPrefix(creator))
}

// GetCreatorsPrefixStore returns the substore that contains a list of creators
func (k Keeper) GetCreatorsPrefixStore(ctx sdk.Context) sdk.KVStore {
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
