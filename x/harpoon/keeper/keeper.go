package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
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
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
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
