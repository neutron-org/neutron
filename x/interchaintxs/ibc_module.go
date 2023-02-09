package interchaintxs

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"

	"github.com/neutron-org/neutron/x/interchaintxs/keeper"
)

var _ porttypes.IBCModule = IBCModule{}

// IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCModule interface. We don't need to implement this handler.
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	_order channeltypes.Order,
	_connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	_counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	// FIXME: always returning plain version is probably a bad idea!
	return version, im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID))
}

// OnChanOpenTry implements the IBCModule interface. We don't need to implement this handler.
func (im IBCModule) OnChanOpenTry(
	_ctx sdk.Context,
	_order channeltypes.Order,
	_connectionHops []string,
	_portID,
	_channelID string,
	_chanCap *capabilitytypes.Capability,
	_counterparty channeltypes.Counterparty,
	_counterpartyVersion string,
) (string, error) {
	return "", nil
}

// OnChanOpenAck implements the IBCModule interface. This handler is called after we create an
// account on a remote zone (because icaControllerKeeper.RegisterInterchainAccount opens a channel).
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterPartyChannelID string,
	counterpartyVersion string,
) error {
	return im.keeper.HandleChanOpenAck(ctx, portID, channelID, counterPartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface. We don't need to implement this handler.
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface. We don't need to implement this handler.
// Handler will be implemented in https://p2pvalidator.atlassian.net/browse/LSC-137
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseConfirm implements the IBCModule interface. We don't need to implement this handler.
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface. A successful acknowledgement
// is returned if the packet data is successfully decoded and the receiving application
// logic returns without error.
func (im IBCModule) OnRecvPacket(
	_ctx sdk.Context,
	_packet channeltypes.Packet,
	_relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return channeltypes.NewErrorAcknowledgement(errors.New("cannot receive packet via interchain accounts authentication module"))
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.keeper.HandleAcknowledgement(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.keeper.HandleTimeout(ctx, packet, relayer)
}
