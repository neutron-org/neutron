package wasmbinding

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	interchainqueriesmodulekeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
)

// RegisterCustomPlugins returns wasmkeeper.Option that we can use to connect handlers for implemented custom queries and messages to the App
func RegisterCustomPlugins(icaControllerKeeper *icacontrollerkeeper.Keeper, icqKeeper *interchainqueriesmodulekeeper.Keeper) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(icaControllerKeeper, icqKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})

	return []wasm.Option{
		queryPluginOpt,
	}
}
