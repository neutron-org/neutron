package keeper

import (
	"fmt"

	"cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		Codec               codec.BinaryCodec
		storeKey            storetypes.StoreKey
		memKey              storetypes.StoreKey
		channelKeeper       types.ChannelKeeper
		feeKeeper           types.FeeRefunderKeeper
		icaControllerKeeper types.ICAControllerKeeper
		sudoKeeper          types.WasmKeeper
		bankKeeper          types.BankKeeper
		feeBurnerKeeper     types.FeeBurnerKeeper
		authority           string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	channelKeeper types.ChannelKeeper,
	icaControllerKeeper types.ICAControllerKeeper,
	sudoKeeper types.WasmKeeper,
	feeKeeper types.FeeRefunderKeeper,
	bankKeeper types.BankKeeper,
	feeBurnerKeeper types.FeeBurnerKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		Codec:               cdc,
		storeKey:            storeKey,
		memKey:              memKey,
		channelKeeper:       channelKeeper,
		icaControllerKeeper: icaControllerKeeper,
		sudoKeeper:          sudoKeeper,
		feeKeeper:           feeKeeper,
		bankKeeper:          bankKeeper,
		feeBurnerKeeper:     feeBurnerKeeper,
		authority:           authority,
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ChargeFee(ctx sdk.Context, payer sdk.AccAddress, fee sdk.Coins) error {
	k.Logger(ctx).Debug("Trying to change fees", "payer", payer, "fee", fee)

	params := k.GetParams(ctx)

	if !fee.IsAnyGTE(params.RegisterFee) {
		return errors.Wrapf(sdkerrors.ErrInsufficientFee, "provided fee is less than min governance set ack fee: %s < %s", fee, params.RegisterFee)
	}

	treasury := k.feeBurnerKeeper.GetParams(ctx).TreasuryAddress
	treasuryAddress, err := sdk.AccAddressFromBech32(treasury)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to convert treasury, bech32 to AccAddress: %s: %s", treasury, err.Error())
	}

	err = k.bankKeeper.SendCoins(ctx, payer, treasuryAddress, fee)
	if err != nil {
		return errors.Wrapf(err, "failed send fee(%s) from %s to %s", fee, payer, treasury)
	}
	return nil
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) GetLastFreeRegisterICACodeID(ctx sdk.Context) (codeID uint64) {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastFreeRegisterICACodeID)

	if bytes == nil {
		k.Logger(ctx).Debug("Last code id key don't exists, GetLastCodeID returns 0")
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}
