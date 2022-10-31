package keeper

import (
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"

	"github.com/neutron-org/neutron/internal/sudo"
	feekeeper "github.com/neutron-org/neutron/x/fee/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/controller/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

const (
	LabelSubmitTx                  = "submit_tx"
	LabelHandleAcknowledgment      = "handle_ack"
	LabelLabelHandleChanOpenAck    = "handle_chan_open_ack"
	LabelRegisterInterchainAccount = "register_interchain_account"
	LabelHandleTimeout             = "handle_timeout"
)

type (
	Keeper struct {
		Codec         codec.BinaryCodec
		storeKey      storetypes.StoreKey
		memKey        storetypes.StoreKey
		paramstore    paramtypes.Subspace
		scopedKeeper  capabilitykeeper.ScopedKeeper
		channelKeeper icatypes.ChannelKeeper
		feeKeeper     *feekeeper.Keeper

		icaControllerKeeper icacontrollerkeeper.Keeper
		wasmKeeper          *wasm.Keeper
		sudoHandler         sudo.Handler
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	paramstore paramtypes.Subspace,
	channelKeeper icatypes.ChannelKeeper,
	wasmKeeper *wasm.Keeper,
	icaControllerKeeper icacontrollerkeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	feeKeeper *feekeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Codec:         cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    paramstore,
		channelKeeper: channelKeeper,

		icaControllerKeeper: icaControllerKeeper,
		scopedKeeper:        scopedKeeper,
		wasmKeeper:          wasmKeeper,
		sudoHandler:         sudo.NewSudoHandler(wasmKeeper, types.ModuleName),
		feeKeeper:           feeKeeper,
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
