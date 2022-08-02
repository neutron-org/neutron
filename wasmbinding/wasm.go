package wasmbinding

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	ictxkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
)

func RegisterCustomPlugins(ictx ictxkeeper.Keeper, icq icqkeeper.Keeper) []wasmkeeper.Option {
	messageDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(CustomMessageDecorator(ictx, icq))

	return []wasm.Option{
		messageDecoratorOpt,
	}
}
