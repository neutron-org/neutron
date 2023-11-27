package ibcswap

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/x/ibcswap/keeper"
	"github.com/neutron-org/neutron/x/ibcswap/types"
)

var (
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModule           = AppModule{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic is the swap middleware AppModuleBasic.
type AppModuleBasic struct{}

// Name implements AppModuleBasic interface.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec implements AppModuleBasic interface.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterInterfaces registers module concrete types into protobuf Any.
func (AppModuleBasic) RegisterInterfaces(_ codectypes.InterfaceRegistry) {}

// DefaultGenesis returns default genesis state as raw bytes for the swap module.
func (AppModuleBasic) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	return nil
}

// ValidateGenesis performs genesis state validation for the swap module.
func (AppModuleBasic) ValidateGenesis(
	_ codec.JSONCodec,
	_ client.TxEncodingConfig,
	_ json.RawMessage,
) error {
	return nil
}

// RegisterRESTRoutes implements AppModuleBasic interface.
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the swap module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

// GetTxCmd implements AppModuleBasic interface.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd implements AppModuleBasic interface.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// AppModule implements the module.AppModule
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(keeper keeper.Keeper) *AppModule {
	return &AppModule{
		keeper: keeper,
	}
}

// RegisterInvariants implements the AppModule interface.
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// QuerierRoute implements the AppModule interface.
func (AppModule) QuerierRoute() string {
	return ""
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(_ module.Configurator) {}

// InitGenesis performs genesis initialization for the ibc-router module. It returns
// no validator updates.
func (am AppModule) InitGenesis(
	_ sdk.Context,
	_ codec.JSONCodec,
	_ json.RawMessage,
) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the swap module.
func (am AppModule) ExportGenesis(_ sdk.Context, _ codec.JSONCodec) json.RawMessage {
	return nil
}

// ConsensusVersion returns the consensus state breaking version for the swap module.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock implements the AppModule interface.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock implements the AppModule interface.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// GenerateGenesisState implements the AppModuleSimulation interface.
func (AppModule) GenerateGenesisState(_ *module.SimulationState) {}

// ProposalContents implements the AppModuleSimulation interface.
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalMsg {
	return nil
}

// RegisterStoreDecoder implements the AppModuleSimulation interface.
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations implements the AppModuleSimulation interface.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
