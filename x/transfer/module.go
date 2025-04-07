package transfer

import (
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/errors"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	wrapkeeper "github.com/neutron-org/neutron/v6/x/transfer/keeper"
	neutrontypes "github.com/neutron-org/neutron/v6/x/transfer/types"
)

/*
	In addition to original ack processing of ibc transfer acknowledgement we want to pass the acknowledgement to originating wasm contract.
	The package contains a code to achieve the purpose.
*/

type IBCModule struct {
	wrappedKeeper      wrapkeeper.KeeperTransferWrapper
	keeper             keeper.Keeper
	sudoKeeper         neutrontypes.WasmKeeper
	tokenfactoryKeeper neutrontypes.TokenfactoryKeeper
	transfer.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k wrapkeeper.KeeperTransferWrapper, sudoKeeper neutrontypes.WasmKeeper, tokenfactoryKeeper neutrontypes.TokenfactoryKeeper) IBCModule {
	return IBCModule{
		wrappedKeeper:      k,
		keeper:             k.Keeper,
		sudoKeeper:         sudoKeeper,
		IBCModule:          transfer.NewIBCModule(k.Keeper),
		tokenfactoryKeeper: tokenfactoryKeeper,
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
		return errors.Wrap(err, "failed to process original OnAcknowledgementPacket")
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
		return errors.Wrap(err, "failed to process original OnTimeoutPacket")
	}
	return im.HandleTimeout(ctx, packet, relayer)
}

func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	_ string,
	counterpartyVersion string,
) error {
	err := im.IBCModule.OnChanOpenAck(ctx, portID, channelID, "", counterpartyVersion)
	if err != nil {
		return errors.Wrap(err, "failed to process original OnChanOpenAck")
	}

	escrowAddress := transfertypes.GetEscrowAddress(portID, channelID)
	im.tokenfactoryKeeper.StoreEscrowAddress(ctx, escrowAddress.Bytes())

	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	err := im.IBCModule.OnChanOpenConfirm(ctx, portID, channelID)
	if err != nil {
		return errors.Wrap(err, "failed to process original OnChanOpenConfirm")
	}

	escrowAddress := transfertypes.GetEscrowAddress(portID, channelID)
	im.tokenfactoryKeeper.StoreEscrowAddress(ctx, escrowAddress.Bytes())

	return nil
}

var _ appmodule.AppModule = AppModule{}

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

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() { // marker
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() { // marker
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	transfertypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)
	neutrontypes.RegisterMsgServer(cfg.MsgServer(), am.keeper)

	cfg.MsgServer().RegisterService(&neutrontypes.MsgServiceDescOrig, am.keeper)

	m := keeper.NewMigrator(am.keeper.Keeper)
	if err := cfg.RegisterMigration(transfertypes.ModuleName, 1, m.MigrateTraces); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 1 to 2: %v", err))
	}

	if err := cfg.RegisterMigration(transfertypes.ModuleName, 2, m.MigrateTotalEscrowForDenom); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 2 to 3: %v", err))
	}

	if err := cfg.RegisterMigration(transfertypes.ModuleName, 3, m.MigrateParams); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app version 3 to 4: %v", err))
	}

	if err := cfg.RegisterMigration(transfertypes.ModuleName, 4, m.MigrateDenomMetadata); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 4 to 5: %v", err))
	}
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
