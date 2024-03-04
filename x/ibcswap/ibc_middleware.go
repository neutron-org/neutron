package ibcswap

import (
	"encoding/json"
	"errors"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/neutron-org/neutron/v2/x/ibcswap/keeper"
	"github.com/neutron-org/neutron/v2/x/ibcswap/types"
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

// For clarity, here is a breakdown of the steps
// 1. Check if this is a swap packet; if not pass it to next middleware
// 2. validate swapMetadata; ErrAck if invalid
// 3. Pass through the middleware stack to ibc-go/transfer#OnRecvPacket; transfer coins are sent to receiver
// 4. Do swap; handle failures

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

	if err := validateSwapPacket(packet, data, *metadata); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Use overrideReceiver so that users cannot ibcswap through arbitrary addresses.
	// Instead generate a unique address for each user based on their channel and origin-address
	originalCreator := m.Swap.Creator
	overrideReceiver, err := packetforward.GetReceiver(packet.DestinationChannel, data.Sender)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}
	metadata.Creator = overrideReceiver
	// Update packet data to match the new receiver so that transfer middleware adds tokens to the expected address
	packet = newPacketWithOverrideReceiver(packet, data, overrideReceiver)

	ack := im.app.OnRecvPacket(ctx, packet, relayer)
	if ack == nil || !ack.Success() {
		return ack
	}

	// Attempt to perform a swap using a cacheCtx
	cacheCtx, writeCache := ctx.CacheContext()
	res, err := im.keeper.Swap(cacheCtx, originalCreator, metadata.MsgPlaceLimitOrder)
	if err != nil {
		return im.handleFailedSwap(ctx, packet, data, metadata, err)
	}

	// If there is no next field set in the metadata return ack
	if metadata.Next == nil {
		writeCache()
		return ack
	}

	// We need to reset the packets memo field so that the root key in the metadata is the
	// next field from the current metadata.
	memoBz, err := json.Marshal(metadata.Next)
	if err != nil {
		return ack
	}

	postSwapData := data
	postSwapData.Memo = string(memoBz)

	// Override the packet data to include the token denom and amount that was received from the swap.
	postSwapData.Denom = res.TakerCoinOut.Denom
	postSwapData.Amount = res.TakerCoinOut.Amount.String()

	// After a successful swap funds are now in the receiver account from the MsgPlaceLimitOrder so,
	// we need to override the packets receiver field before invoking the forward middlewares OnRecvPacket.
	postSwapData.Receiver = m.Swap.Receiver

	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&postSwapData)
	if err != nil {
		return ack
	}

	packet.Data = dataBz

	// The forward middleware should return a nil ack if the forward is initiated properly.
	// If not an error occurred, and we return the original ack.
	newAck := im.app.OnRecvPacket(cacheCtx, packet, relayer)
	if newAck != nil {
		return im.handleFailedSwap(ctx, packet, data, metadata, errors.New(string(newAck.Acknowledgement())))
	}

	writeCache()
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
		"refund address", metadata.NeutronRefundAddress,
	)

	// The current denom is from the sender chains perspective, we need to compose the appropriate denom for this side
	denomOnThisChain := getDenomForThisChain(packet, data.Denom)

	if len(metadata.NeutronRefundAddress) != 0 {
		return im.handleOnChainRefund(ctx, data, metadata, denomOnThisChain, err)
	}

	return im.handleIBCRefund(ctx, packet, data, metadata, denomOnThisChain, err)
}

// handleOnChainRefund will compose a successful ack to send back to the counterparty chain containing any error messages.
// Returning a successful ack ensures that a refund is not issued on the counterparty chain.
// See: https://github.com/cosmos/ibc-go/blob/3ecc7dd3aef5790ec5d906936a297b34adf1ee41/modules/apps/transfer/keeper/relay.go#L320
func (im IBCMiddleware) handleOnChainRefund(
	ctx sdk.Context,
	data transfertypes.FungibleTokenPacketData,
	metadata *types.SwapMetadata,
	newDenom string,
	swapErr error,
) ibcexported.Acknowledgement {
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
	err := im.keeper.SendCoins(ctx, metadata.Creator, metadata.NeutronRefundAddress, sdk.NewCoins(token))
	if err != nil {
		wrappedErr := sdkerrors.Wrap(err, "failed to move funds to refund address")
		wrappedErr = sdkerrors.Wrap(swapErr, wrappedErr.Error())
		return channeltypes.NewResultAcknowledgement([]byte(wrappedErr.Error()))
	}

	return channeltypes.NewResultAcknowledgement([]byte(swapErr.Error()))
}

// handleIBCRefund will either burn or transfer the funds back to the appropriate escrow account.
// When a packet comes in the transfer module's OnRecvPacket callback is invoked which either
// mints or unescrows funds on this side so if the swap fails an explicit refund is required.
func (im IBCMiddleware) handleIBCRefund(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	metadata *types.SwapMetadata,
	newDenom string,
	swapErr error,
) ibcexported.Acknowledgement {
	data.Denom = newDenom

	err := im.keeper.RefundPacketToken(ctx, packet, data, metadata)
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
func getDenomForThisChain(packet channeltypes.Packet, denom string) string {
	counterpartyPrefix := transfertypes.GetDenomPrefix(packet.SourcePort, packet.SourceChannel)
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
	prefixedDenom := transfertypes.GetDenomPrefix(packet.DestinationPort, packet.DestinationChannel) + denom
	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}

// Update the packet data to reflect the new receiver address that is used by the PFM
func newPacketWithOverrideReceiver(oldPacket channeltypes.Packet, data transfertypes.FungibleTokenPacketData, overrideReceiver string) channeltypes.Packet {
	overrideData := transfertypes.FungibleTokenPacketData{
		Denom:    data.Denom,
		Amount:   data.Amount,
		Sender:   data.Sender,
		Receiver: overrideReceiver, // override receiver
	}
	overrideDataBz := transfertypes.ModuleCdc.MustMarshalJSON(&overrideData)

	return channeltypes.Packet{
		Sequence:           oldPacket.Sequence,
		SourcePort:         oldPacket.SourcePort,
		SourceChannel:      oldPacket.SourceChannel,
		DestinationPort:    oldPacket.DestinationPort,
		DestinationChannel: oldPacket.DestinationChannel,
		Data:               overrideDataBz, // override data
		TimeoutHeight:      oldPacket.TimeoutHeight,
		TimeoutTimestamp:   oldPacket.TimeoutTimestamp,
	}
}

func validateSwapPacket(packet channeltypes.Packet, transferData transfertypes.FungibleTokenPacketData, sm types.SwapMetadata) error {
	denomOnNeutron := getDenomForThisChain(packet, transferData.Denom)
	if denomOnNeutron != sm.TokenIn {
		return sdkerrors.Wrap(types.ErrInvalidSwapMetadata, "Transfer Denom must match TokenIn")
	}

	transferAmount, ok := math.NewIntFromString(transferData.Amount)
	if !ok {
		return sdkerrors.Wrapf(
			transfertypes.ErrInvalidAmount,
			"unable to parse transfer amount (%s) into math.Int",
			transferData.Amount,
		)
	}

	if transferAmount.LT(sm.AmountIn) {
		return sdkerrors.Wrap(types.ErrInvalidSwapMetadata, "Transfer amount must be >= AmountIn")
	}

	if sm.ContainsPFM() {
		return sdkerrors.Wrap(
			types.ErrInvalidSwapMetadata,
			"ibcswap middleware cannot be used in conjunction with packet-forward-middleware",
		)
	}
	return nil
}
