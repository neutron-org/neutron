package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	dextypes "github.com/neutron-org/neutron/v3/x/dex/types"
	"github.com/neutron-org/neutron/v3/x/ibcswap/types"
)

// Keeper defines the swap middleware keeper.
type Keeper struct {
	cdc              codec.BinaryCodec
	msgServiceRouter *baseapp.MsgServiceRouter

	ics4Wrapper porttypes.ICS4Wrapper
	bankKeeper  types.BankKeeper
}

// NewKeeper creates a new swap Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	msgServiceRouter *baseapp.MsgServiceRouter,
	ics4Wrapper porttypes.ICS4Wrapper,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:              cdc,
		msgServiceRouter: msgServiceRouter,

		ics4Wrapper: ics4Wrapper,
		bankKeeper:  bankKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ibcexported.ModuleName+"-"+types.ModuleName)
}

// Swap calls into the base app's msg service router so that the appropriate handler is called when sending the swap msg.
func (k Keeper) Swap(
	ctx sdk.Context,
	originalCreator string,
	msg *dextypes.MsgPlaceLimitOrder,
) (*dextypes.MsgPlaceLimitOrderResponse, error) {
	swapHandler := k.msgServiceRouter.Handler(msg)
	if swapHandler == nil {
		return nil, sdkerrors.Wrap(
			types.ErrMsgHandlerInvalid,
			fmt.Sprintf("could not find the handler for %T", msg),
		)
	}

	res, err := swapHandler(ctx, msg)
	if err != nil {
		return nil, err
	}

	msgSwapRes := &dextypes.MsgPlaceLimitOrderResponse{}
	if err := proto.Unmarshal(res.Data, msgSwapRes); err != nil {
		return nil, err
	}

	amountUnused := msg.AmountIn.Sub(msgSwapRes.CoinIn.Amount)
	// If not all tokenIn is swapped and a temporary creator address has been used
	// return the unused portion to the original creator address
	if amountUnused.IsPositive() && originalCreator != msg.Creator {
		overrrideCreatorAddr := sdk.MustAccAddressFromBech32(msg.Creator)
		originalCreatorAddr := sdk.MustAccAddressFromBech32(originalCreator)
		unusedCoin := sdk.NewCoin(msg.TokenIn, amountUnused)

		err := k.bankKeeper.SendCoins(ctx, overrrideCreatorAddr, originalCreatorAddr, sdk.Coins{unusedCoin})
		if err != nil {
			return nil, err
		}
	}

	return msgSwapRes, nil
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function.
func (k Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return k.ics4Wrapper.SendPacket(
		ctx,
		chanCap,
		sourcePort,
		sourceChannel,
		timeoutHeight,
		timeoutTimestamp,
		data,
	)
}

// WriteAcknowledgement wraps IBC ChannelKeeper's WriteAcknowledgement function.
func (k Keeper) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	acknowledgement ibcexported.Acknowledgement,
) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}

// RefundPacketToken handles the burning or escrow lock up of vouchers when an asset should be refunded.
// This is only used in the case where we call into the transfer modules OnRecvPacket callback but then the swap fails.
func (k Keeper) RefundPacketToken(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	metadata *types.SwapMetadata,
) error {
	// parse the denomination from the full denom path
	trace := transfertypes.ParseDenomTrace(data.Denom)

	// parse the transfer amount
	transferAmount, ok := math.NewIntFromString(data.Amount)
	if !ok {
		return sdkerrors.Wrapf(
			transfertypes.ErrInvalidAmount,
			"unable to parse transfer amount (%s) into math.Int",
			data.Amount,
		)
	}
	token := sdk.NewCoin(trace.IBCDenom(), transferAmount)

	// decode the creator address
	receiver, err := sdk.AccAddressFromBech32(metadata.Creator)
	if err != nil {
		return err
	}

	// if the sender chain is source that means a voucher was minted on Neutron when the ics20 transfer took place
	if transfertypes.SenderChainIsSource(packet.SourcePort, packet.SourceChannel, data.Denom) {
		// transfer coins from user account to transfer module
		err = k.bankKeeper.SendCoinsFromAccountToModule(
			ctx,
			receiver,
			types.ModuleName,
			sdk.NewCoins(token),
		)
		if err != nil {
			return err
		}

		// burn the coins
		err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(token))
		if err != nil {
			return err
		}

		return nil
	}

	// transfer coins from user account to escrow address
	escrowAddress := transfertypes.GetEscrowAddress(
		packet.GetSourcePort(),
		packet.GetSourceChannel(),
	)
	err = k.bankKeeper.SendCoins(ctx, receiver, escrowAddress, sdk.NewCoins(token))
	if err != nil {
		return err
	}

	return nil
}

// SendCoins wraps the BankKeepers SendCoins function so it can be invoked from the middleware.
func (k Keeper) SendCoins(ctx sdk.Context, fromAddr, toAddr string, amt sdk.Coins) error {
	from, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return err
	}

	to, err := sdk.AccAddressFromBech32(toAddr)
	if err != nil {
		return err
	}

	return k.bankKeeper.SendCoins(ctx, from, to, amt)
}

func (k Keeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
