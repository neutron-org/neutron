package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/interchaintxs/types"
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
		Codec                  codec.BinaryCodec
		storeKey               storetypes.StoreKey
		memKey                 storetypes.StoreKey
		channelKeeper          types.ChannelKeeper
		feeKeeper              types.FeeRefunderKeeper
		icaControllerKeeper    types.ICAControllerKeeper
		icaControllerMsgServer types.ICAControllerMsgServer
		sudoKeeper             types.WasmKeeper
		bankKeeper             types.BankKeeper
		getFeeCollectorAddr    types.GetFeeCollectorAddr
		authority              string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	channelKeeper types.ChannelKeeper,
	icaControllerKeeper types.ICAControllerKeeper,
	icaControllerMsgServer types.ICAControllerMsgServer,
	sudoKeeper types.WasmKeeper,
	feeKeeper types.FeeRefunderKeeper,
	bankKeeper types.BankKeeper,
	getFeeCollectorAddr types.GetFeeCollectorAddr,
	authority string,
) *Keeper {
	return &Keeper{
		Codec:                  cdc,
		storeKey:               storeKey,
		memKey:                 memKey,
		channelKeeper:          channelKeeper,
		icaControllerKeeper:    icaControllerKeeper,
		icaControllerMsgServer: icaControllerMsgServer,
		sudoKeeper:             sudoKeeper,
		feeKeeper:              feeKeeper,
		bankKeeper:             bankKeeper,
		getFeeCollectorAddr:    getFeeCollectorAddr,
		authority:              authority,
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

	feeCollector := k.getFeeCollectorAddr(ctx)
	feeCollectorAddress, err := sdk.AccAddressFromBech32(feeCollector)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to convert fee collector, bech32 to AccAddress: %s: %s", feeCollector, err.Error())
	}

	err = k.bankKeeper.SendCoins(ctx, payer, feeCollectorAddress, fee)
	if err != nil {
		return errors.Wrapf(err, "failed send fee(%s) from %s to %s", fee, payer, feeCollectorAddress)
	}
	return nil
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetICARegistrationFeeFirstCodeID returns code id, starting from which we charge fee for ICA registration
func (k Keeper) GetICARegistrationFeeFirstCodeID(ctx sdk.Context) (codeID uint64) {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.ICARegistrationFeeFirstCodeID)
	if bytes == nil {
		k.Logger(ctx).Debug("Fee register ICA code id key don't exists, GetLastCodeID returns 0")
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}
