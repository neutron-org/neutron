package transfer

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	wrapkeeper "github.com/neutron-org/neutron/x/transfer/keeper"
	neutrontypes "github.com/neutron-org/neutron/x/transfer/types"
)

/*
	In addition to original ack processing of ibc transfer acknowledgement we want to pass the acknowledgement to originating wasm contract.
	The package contains a code to achieve the purpose.
*/

type IBCModule struct {
	wrappedKeeper         wrapkeeper.KeeperTransferWrapper
	keeper                keeper.Keeper
	ContractManagerKeeper neutrontypes.ContractManagerKeeper
	transfer.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k wrapkeeper.KeeperTransferWrapper) IBCModule {
	return IBCModule{
		wrappedKeeper:         k,
		keeper:                k.Keeper,
		ContractManagerKeeper: k.ContractManagerKeeper,
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
	return im.HandleAcknowledgement(ctx, packet, acknowledgement, relayer)
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
	return im.HandleTimeout(ctx, packet, relayer)
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
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

type AppModuleBasic struct {
	transfer.AppModuleBasic
}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{AppModuleBasic: transfer.AppModuleBasic{}}
}

func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {
	neutrontypes.RegisterLegacyAminoCodec(cdc)
}

func (am AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	neutrontypes.RegisterLegacyAminoCodec(cdc)
	am.AppModuleBasic.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (am AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	neutrontypes.RegisterInterfaces(reg)
	am.AppModuleBasic.RegisterInterfaces(reg)
}

// Name returns the capability module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// Deprecated: Route returns the capability module's message routing key.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper))
}

func NewHandler(k wrapkeeper.KeeperTransferWrapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgTransfer:
			neutronMsg := neutrontypes.MsgTransfer{
				SourcePort:       msg.SourcePort,
				SourceChannel:    msg.SourceChannel,
				Token:            msg.Token,
				Sender:           msg.Sender,
				Receiver:         msg.Receiver,
				TimeoutHeight:    msg.TimeoutHeight,
				TimeoutTimestamp: msg.TimeoutTimestamp,
				Fee:              feetypes.Fee{},
			}
			res, err := k.Transfer(sdk.WrapSDKContext(ctx), &neutronMsg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
