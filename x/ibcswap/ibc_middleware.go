package ibcswap

import (
	"context"
	"encoding/json"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	forwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/neutron-org/neutron/x/ibcswap/keeper"
	"github.com/neutron-org/neutron/x/ibcswap/types"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks for the swap middleware given the
// swap keeper and the underlying application.
type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper keeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application.
func NewIBCMiddleware(app porttypes.IBCModule, k keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		chanCap,
		counterparty,
		version,
	)
}

// OnChanOpenTry implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID, channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	return im.app.OnChanOpenTry(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		chanCap,
		counterparty,
		counterpartyVersion,
	)
}

// OnChanOpenAck implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID, channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket checks the memo field on this packet and if the metadata inside's root key indicates this packet
// should be handled by the swap middleware it attempts to perform a swap. If the swap is successful
// the underlying application's OnRecvPacket callback is invoked.
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	m := &types.PacketMetadata{}
	err := json.Unmarshal([]byte(data.Memo), m)
	if err != nil || m.Swap == nil {
		// Not a packet that should be handled by the swap middleware
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	metadata := m.Swap
	if err := metadata.Validate(); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Compose our context with values that will be used to pass through to the forward middleware
	ctxWithForwardFlags := context.WithValue(ctx.Context(), forwardtypes.ProcessedKey{}, true)
	ctxWithForwardFlags = context.WithValue(
		ctxWithForwardFlags,
		forwardtypes.NonrefundableKey{},
		true,
	)
	ctxWithForwardFlags = context.WithValue(
		ctxWithForwardFlags,
		forwardtypes.DisableDenomCompositionKey{},
		true,
	)
	wrappedSdkCtx := ctx.WithContext(ctxWithForwardFlags)

	ack := im.app.OnRecvPacket(wrappedSdkCtx, packet, relayer)
	if ack == nil || !ack.Success() {
		return ack
	}

	// Attempt to perform a swap since this packets memo included swap metadata.
	res, err := im.keeper.Swap(ctx, metadata.MsgPlaceLimitOrder)
	if err != nil {
		return im.handleFailedSwap(ctx, packet, data, metadata, err)
	}

	// If there is no next field set in the metadata return ack
	if metadata.Next == nil {
		return ack
	}

	// We need to reset the packets memo field so that the root key in the metadata is the
	// next field from the current metadata.
	memoBz, err := json.Marshal(metadata.Next)
	if err != nil {
		return ack
	}

	data.Memo = string(memoBz)

	// Override the packet data to include the token denom and amount that was received from the swap.
	data.Denom = res.TakerCoinOut.Denom
	data.Amount = res.TakerCoinOut.Amount.String()

	// After a successful swap funds are now in the receiver account from the MsgPlaceLimitOrder so,
	// we need to override the packets receiver field before invoking the forward middlewares OnRecvPacket.
	data.Receiver = m.Swap.Receiver

	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	if err != nil {
		return ack
	}

	packet.Data = dataBz

	// The forward middleware should return a nil ack if the forward is initiated properly.
	// If not an error occurred, and we return the original ack.
	newAck := im.app.OnRecvPacket(wrappedSdkCtx, packet, relayer)
	if newAck != nil {
		return ack
	}

	return nil
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

func (im IBCMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return im.keeper.SendPacket(
		ctx,
		chanCap,
		sourcePort,
		sourceChannel,
		timeoutHeight,
		timeoutTimestamp,
		data,
	)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface.
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	return im.keeper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

func (im IBCMiddleware) GetAppVersion(
	ctx sdk.Context,
	portID string,
	channelID string,
) (string, bool) {
	return im.keeper.GetAppVersion(ctx, portID, channelID)
}

// handleFailedSwap will invoke the appropriate failover logic depending on if this swap was marked refundable
// or non-refundable in the SwapMetadata.
func (im IBCMiddleware) handleFailedSwap(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	metadata *types.SwapMetadata,
	err error,
) ibcexported.Acknowledgement {
	swapErr := sdkerrors.Wrap(types.ErrSwapFailed, err.Error())
	im.keeper.Logger(ctx).Error(
		"ibc swap failed",
		"err", swapErr,
		"creator", metadata.Creator,
		"receiver", metadata.Receiver,
		"tokenIn", metadata.TokenIn,
		"tokenOut", metadata.TokenOut,
		"AmountIn", metadata.AmountIn,
		"TickIndexInToOut", metadata.TickIndexInToOut,
		"OrderType", metadata.OrderType,
		"refundable", metadata.NonRefundable,
		"refund address", metadata.RefundAddress,
	)

	// The current denom is from the sender chains perspective, we need to compose the appropriate denom for this side
	denomOnThisChain := getDenomForThisChain(
		packet.DestinationPort, packet.DestinationChannel,
		packet.SourcePort, packet.SourceChannel,
		data.Denom,
	)

	if metadata.NonRefundable {
		return im.handleNoRefund(ctx, data, metadata, denomOnThisChain, err)
	}

	return im.handleRefund(ctx, packet, data, denomOnThisChain, err)
}

// handleNoRefund will compose a successful ack to send back to the counterparty chain containing any error messages.
// Returning a successful ack ensures that a refund is not issued on the counterparty chain.
// See: https://github.com/cosmos/ibc-go/blob/3ecc7dd3aef5790ec5d906936a297b34adf1ee41/modules/apps/transfer/keeper/relay.go#L320
func (im IBCMiddleware) handleNoRefund(
	ctx sdk.Context,
	data transfertypes.FungibleTokenPacketData,
	metadata *types.SwapMetadata,
	newDenom string,
	swapErr error,
) ibcexported.Acknowledgement {
	if metadata.RefundAddress == "" {
		return channeltypes.NewResultAcknowledgement([]byte(swapErr.Error()))
	}

	amount, ok := math.NewIntFromString(data.Amount)
	if !ok {
		wrappedErr := sdkerrors.Wrapf(
			transfertypes.ErrInvalidAmount,
			"unable to parse transfer amount (%s) into math.Int",
			data.Amount,
		)
		wrappedErr = sdkerrors.Wrap(swapErr, wrappedErr.Error())
		return channeltypes.NewResultAcknowledgement([]byte(wrappedErr.Error()))
	}

	token := sdk.NewCoin(newDenom, amount)
	err := im.keeper.SendCoins(ctx, data.Receiver, metadata.RefundAddress, sdk.NewCoins(token))
	if err != nil {
		wrappedErr := sdkerrors.Wrap(err, "failed to move funds to refund address")
		wrappedErr = sdkerrors.Wrap(swapErr, wrappedErr.Error())
		return channeltypes.NewResultAcknowledgement([]byte(wrappedErr.Error()))
	}

	return channeltypes.NewResultAcknowledgement([]byte(swapErr.Error()))
}

// handleRefund will either burn or transfer the funds back to the appropriate escrow account.
// When a packet comes in the transfer module's OnRecvPacket callback is invoked which either
// mints or unescrows funds on this side so if the swap fails an explicit refund is required.
func (im IBCMiddleware) handleRefund(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	newDenom string,
	swapErr error,
) ibcexported.Acknowledgement {
	data.Denom = newDenom

	err := im.keeper.RefundPacketToken(ctx, packet, data)
	if err != nil {
		wrappedErr := sdkerrors.Wrap(swapErr, err.Error())

		// If the refund fails on this side we want to make sure that the refund does not happen on the counterparty,
		// so we return a successful ack containing the error
		return channeltypes.NewResultAcknowledgement([]byte(wrappedErr.Error()))
	}

	return channeltypes.NewErrorAcknowledgement(swapErr)
}

// getDenomForThisChain composes a new token denom by either unwinding or prefixing the specified token denom appropriately.
// This is necessary because the token denom in the packet data is from the perspective of the counterparty chain.
func getDenomForThisChain(
	port, channel, counterpartyPort, counterpartyChannel, denom string,
) string {
	counterpartyPrefix := transfertypes.GetDenomPrefix(counterpartyPort, counterpartyChannel)
	if strings.HasPrefix(denom, counterpartyPrefix) {
		// unwind denom
		unwoundDenom := denom[len(counterpartyPrefix):]
		denomTrace := transfertypes.ParseDenomTrace(unwoundDenom)
		if denomTrace.Path == "" {
			// denom is now unwound back to native denom
			return unwoundDenom
		}
		// denom is still IBC denom
		return denomTrace.IBCDenom()
	}

	// append port and channel from this chain to denom
	prefixedDenom := transfertypes.GetDenomPrefix(port, channel) + denom
	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}
