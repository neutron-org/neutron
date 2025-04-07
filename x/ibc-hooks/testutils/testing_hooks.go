package testutils

import (
	// external libraries
	sdk "github.com/cosmos/cosmos-sdk/types"

	// ibc-go
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	ibchooks "github.com/neutron-org/neutron/v6/x/ibc-hooks"
)

var (
	_ ibchooks.Hooks = TestRecvOverrideHooks{}
	_ ibchooks.Hooks = TestRecvBeforeAfterHooks{}
)

type Status struct {
	OverrideRan bool
	BeforeRan   bool
	AfterRan    bool
}

// Recv
type TestRecvOverrideHooks struct{ Status *Status }

func (t TestRecvOverrideHooks) OnRecvPacketOverride(im ibchooks.IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	t.Status.OverrideRan = true
	ack := im.App.OnRecvPacket(ctx, packet, relayer)
	return ack
}

type TestRecvBeforeAfterHooks struct{ Status *Status }

func (t TestRecvBeforeAfterHooks) OnRecvPacketBeforeHook(_ sdk.Context, _ channeltypes.Packet, _ sdk.AccAddress) {
	t.Status.BeforeRan = true
}

func (t TestRecvBeforeAfterHooks) OnRecvPacketAfterHook(_ sdk.Context, _ channeltypes.Packet, _ sdk.AccAddress, _ ibcexported.Acknowledgement) {
	t.Status.AfterRan = true
}
