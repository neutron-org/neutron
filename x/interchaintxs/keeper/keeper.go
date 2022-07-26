package keeper

import (
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/neutron-org/neutron/internal/sudo"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeKey     storetypes.StoreKey
		memKey       storetypes.StoreKey
		paramstore   paramtypes.Subspace
		scopedKeeper capabilitykeeper.ScopedKeeper

		icaControllerKeeper icacontrollerkeeper.Keeper
		wasmKeeper          *wasm.Keeper
		sudoHandler         sudo.SudoHandler
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

	wasmKeeper *wasm.Keeper,
	icaControllerKeeper icacontrollerkeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,

		icaControllerKeeper: icaControllerKeeper,
		scopedKeeper:        scopedKeeper,
		wasmKeeper:          wasmKeeper,
		sudoHandler:         sudo.NewSudoHandler(wasmKeeper, types.ModuleName),
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
