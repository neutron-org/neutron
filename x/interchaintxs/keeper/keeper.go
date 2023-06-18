package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

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
		Codec                 codec.BinaryCodec
		storeKey              storetypes.StoreKey
		memKey                storetypes.StoreKey
		paramstore            paramtypes.Subspace
		scopedKeeper          types.ScopedKeeper
		channelKeeper         types.ChannelKeeper
		feeKeeper             types.FeeRefunderKeeper
		icaControllerKeeper   types.ICAControllerKeeper
		contractManagerKeeper types.ContractManagerKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	paramstore paramtypes.Subspace,
	channelKeeper types.ChannelKeeper,
	icaControllerKeeper types.ICAControllerKeeper,
	scopedKeeper types.ScopedKeeper,
	contractManagerKeeper types.ContractManagerKeeper,
	feeKeeper types.FeeRefunderKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Codec:                 cdc,
		storeKey:              storeKey,
		memKey:                memKey,
		paramstore:            paramstore,
		channelKeeper:         channelKeeper,
		icaControllerKeeper:   icaControllerKeeper,
		scopedKeeper:          scopedKeeper,
		contractManagerKeeper: contractManagerKeeper,
		feeKeeper:             feeKeeper,
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
