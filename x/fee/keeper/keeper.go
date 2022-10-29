package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/neutron-org/neutron/x/fee/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		bankKeeper types.BankKeeper
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

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
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) LockFees(ctx sdk.Context, payer sdk.AccAddress, packetID channeltypes.PacketId, fee *ibcfeetypes.Fee) error {
	store := ctx.KVStore(k.storeKey)

	feeInfo := types.FeeInfo{
		Payer:    payer.String(),
		PayerFee: fee,
	}
	bzFeeInfo := k.cdc.MustMarshal(&feeInfo)
	store.Set(types.GetFeePacketKey(packetID.ChannelId, packetID.PortId, packetID.Sequence), bzFeeInfo)

	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, payer, types.ModuleName, fee.Total())
}

func (k Keeper) PayFees(ctx sdk.Context, receiver sdk.AccAddress, packetID channeltypes.PacketId) error {
	store := ctx.KVStore(k.storeKey)

	var feeInfo types.FeeInfo
	bzFeeInfo := store.Get(types.GetFeePacketKey(packetID.ChannelId, packetID.PortId, packetID.Sequence))
	if bzFeeInfo == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "no fee info for the given channelID = %s, portID = %s and sequence = %d", packetID.ChannelId, packetID.PortId, packetID.Sequence)
	}
	k.cdc.MustUnmarshal(bzFeeInfo, &feeInfo)

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, feeInfo.PayerFee.Total())
}

func (k Keeper) GetFeeInfo(ctx sdk.Context, packetID channeltypes.PacketId) (*types.FeeInfo, error) {
	store := ctx.KVStore(k.storeKey)

	var feeInfo types.FeeInfo
	bzFeeInfo := store.Get(types.GetFeePacketKey(packetID.ChannelId, packetID.PortId, packetID.Sequence))
	if bzFeeInfo == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "no fee info for the given channelID = %s, portID = %s and sequence = %d", packetID.ChannelId, packetID.PortId, packetID.Sequence)
	}
	k.cdc.MustUnmarshal(bzFeeInfo, &feeInfo)

	return &feeInfo, nil
}
