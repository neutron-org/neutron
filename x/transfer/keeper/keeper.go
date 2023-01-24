package transfer

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	wrappedtypes "github.com/neutron-org/neutron/x/transfer/types"
)

// KeeperTransferWrapper is a wrapper for original ibc keeper to override response for "Transfer" method
type KeeperTransferWrapper struct {
	keeper.Keeper
	channelKeeper         wrappedtypes.ChannelKeeper
	FeeKeeper             wrappedtypes.FeeRefunderKeeper
	ContractManagerKeeper wrappedtypes.ContractManagerKeeper
}

func (k KeeperTransferWrapper) Transfer(goCtx context.Context, msg *wrappedtypes.MsgTransfer) (*wrappedtypes.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	payerInfo, err := k.FeeKeeper.GetPayerInfo(ctx, msg.Sender, msg.Fee.Payer)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to get payer info for sender: %s, payer: %s", msg.Sender, msg.Fee.Payer)
	}

	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, msg.SourcePort, msg.SourceChannel)
	if !found {
		return nil, sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", msg.SourcePort, msg.SourceChannel,
		)
	}

	// if the sender is a contract, lock fees.
	// Because contracts are required to pay fees for the acknowledgements
	if k.ContractManagerKeeper.HasContractInfo(ctx, payerInfo.Sender) {
		if err := k.FeeKeeper.LockFees(ctx, payerInfo, feetypes.NewPacketID(msg.SourcePort, msg.SourceChannel, sequence), msg.Fee); err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to lock fees to pay for transfer msg: %v", msg)
		}
	}

	if _, err := k.Keeper.Transfer(goCtx, types.NewMsgTransfer(msg.SourcePort, msg.SourceChannel, msg.Token, msg.Sender, msg.Receiver, msg.TimeoutHeight, msg.TimeoutTimestamp)); err != nil {
		return nil, err
	}

	return &wrappedtypes.MsgTransferResponse{
		SequenceId: sequence,
		Channel:    msg.SourceChannel,
	}, nil
}

// NewKeeper creates a new IBC transfer Keeper(KeeperTransferWrapper) instance
func NewKeeper(
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	ics4Wrapper types.ICS4Wrapper, channelKeeper wrappedtypes.ChannelKeeper, portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, scopedKeeper capabilitykeeper.ScopedKeeper,
	feeKeeper wrappedtypes.FeeRefunderKeeper,
	contractManagerKeeper wrappedtypes.ContractManagerKeeper,
) KeeperTransferWrapper {
	return KeeperTransferWrapper{
		channelKeeper: channelKeeper,
		Keeper: keeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper,
			authKeeper, bankKeeper, scopedKeeper),
		FeeKeeper:             feeKeeper,
		ContractManagerKeeper: contractManagerKeeper,
	}
}
