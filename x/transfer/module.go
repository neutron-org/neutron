package transfer

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	wrapkeeper "github.com/neutron-org/neutron/x/transfer/keeper"
	neutrontypes "github.com/neutron-org/neutron/x/transfer/types"
)

/*
	In addition to original ack processing of ibc transfer acknowledgement we want to pass the acknowledgement to originating wasm contract.
	The package contains a code to achieve the purpose.
*/

type IBCModule struct {
	keeper                keeper.Keeper
	ContractmanagerKeeper neutrontypes.ContractManagerKeeper

	transfer.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k wrapkeeper.KeeperTransferWrapper, wasmKeeper *wasm.Keeper) IBCModule {
	return IBCModule{
		keeper:                k.Keeper,
		ContractmanagerKeeper: k.ContractmanagerKeeper,
		IBCModule:             transfer.NewIBCModule(k.Keeper),
	}
}

// OnAcknowledgementPacket implements the IBCModule interface.
// Wrapper struct shadows(overrides) the OnAcknowledgementPacket method to achieve the package's purpose.
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	err := im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to process original OnAcknowledgementPacket")
	}
	return im.HandleAcknowledgement(ctx, packet, acknowledgement)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	err := im.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to process original OnTimeoutPacket")
	}
	return im.HandleTimeout(ctx, packet)
}

type AppModule struct {
	transfer.AppModule
	keeper wrapkeeper.KeeperTransferWrapper
}

// NewAppModule creates a new 20-transfer module
func NewAppModule(k wrapkeeper.KeeperTransferWrapper) AppModule {
	return AppModule{
		AppModule: transfer.NewAppModule(k.Keeper),
		keeper:    k,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	neutrontypes.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	neutrontypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}
