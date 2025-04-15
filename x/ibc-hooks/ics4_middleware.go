package ibchooks

import (
	"cosmossdk.io/errors"
	// external libraries
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck

	"github.com/neutron-org/neutron/v6/x/ibc-hooks/types"

	// ibc-go
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

var _ porttypes.ICS4Wrapper = &ICS4Middleware{}

type ICS4Middleware struct {
	channelKeeper types.ChannelKeeper
	channel       porttypes.ICS4Wrapper

	// Hooks
	Hooks Hooks
}

func NewICS4Middleware(channelKeeper types.ChannelKeeper, channel porttypes.ICS4Wrapper, hooks Hooks) ICS4Middleware {
	return ICS4Middleware{
		channelKeeper: channelKeeper,
		channel:       channel,
		Hooks:         hooks,
	}
}

func (i ICS4Middleware) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, sourcePort, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (sequence uint64, err error) {
	channel, found := i.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return 0, errors.Wrap(channeltypes.ErrChannelNotFound, sourceChannel)
	}

	packet := channeltypes.NewPacket(data, sequence, sourcePort, sourceChannel,
		channel.Counterparty.PortId, channel.Counterparty.ChannelId, timeoutHeight, timeoutTimestamp)
	if hook, ok := i.Hooks.(SendPacketOverrideHooks); ok {
		return 0, hook.SendPacketOverride(i, ctx, channelCap, packet)
	}

	if hook, ok := i.Hooks.(SendPacketBeforeHooks); ok {
		hook.SendPacketBeforeHook(ctx, channelCap, packet)
	}

	sequence, err = i.channel.SendPacket(ctx, channelCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)

	if hook, ok := i.Hooks.(SendPacketAfterHooks); ok {
		hook.SendPacketAfterHook(ctx, channelCap, packet, err)
	}

	return sequence, err
}

func (i ICS4Middleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	if hook, ok := i.Hooks.(WriteAcknowledgementOverrideHooks); ok {
		return hook.WriteAcknowledgementOverride(i, ctx, chanCap, packet, ack)
	}

	if hook, ok := i.Hooks.(WriteAcknowledgementBeforeHooks); ok {
		hook.WriteAcknowledgementBeforeHook(ctx, chanCap, packet, ack)
	}
	err := i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
	if hook, ok := i.Hooks.(WriteAcknowledgementAfterHooks); ok {
		hook.WriteAcknowledgementAfterHook(ctx, chanCap, packet, ack, err)
	}

	return err
}

func (i ICS4Middleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	if hook, ok := i.Hooks.(GetAppVersionOverrideHooks); ok {
		return hook.GetAppVersionOverride(i, ctx, portID, channelID)
	}

	if hook, ok := i.Hooks.(GetAppVersionBeforeHooks); ok {
		hook.GetAppVersionBeforeHook(ctx, portID, channelID)
	}
	version, err := i.channel.GetAppVersion(ctx, portID, channelID)
	if hook, ok := i.Hooks.(GetAppVersionAfterHooks); ok {
		hook.GetAppVersionAfterHook(ctx, portID, channelID, version, err)
	}

	return version, err
}
