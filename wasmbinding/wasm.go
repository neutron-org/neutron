package wasmbinding

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	interchainqueriesmodulekeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	icacontrollerkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
	interchaintransactionsmodulekeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
)

// RegisterCustomPlugins returns wasmkeeper.Option that we can use to connect handlers for implemented custom queries and messages to the App
func RegisterCustomPlugins(icaControllerKeeper *icacontrollerkeeper.Keeper, icqKeeper *interchainqueriesmodulekeeper.Keeper, ictxKeeper *interchaintransactionsmodulekeeper.Keeper) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(icaControllerKeeper, icqKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messagePluginOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(ictxKeeper, icqKeeper),
	)

	return []wasm.Option{
		queryPluginOpt,
		messagePluginOpt,
	}
}
