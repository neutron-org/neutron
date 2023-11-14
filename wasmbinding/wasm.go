package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	contractmanagerkeeper "github.com/neutron-org/neutron/v2/x/contractmanager/keeper"
	cronkeeper "github.com/neutron-org/neutron/v2/x/cron/keeper"
	dexkeeper "github.com/neutron-org/neutron/x/dex/keeper"
	feeburnerkeeper "github.com/neutron-org/neutron/v2/x/feeburner/keeper"
	feerefunderkeeper "github.com/neutron-org/neutron/v2/x/feerefunder/keeper"

	adminmodulekeeper "github.com/cosmos/admin-module/x/adminmodule/keeper"

	interchainqueriesmodulekeeper "github.com/neutron-org/neutron/v2/x/interchainqueries/keeper"
	interchaintransactionsmodulekeeper "github.com/neutron-org/neutron/v2/x/interchaintxs/keeper"
	tokenfactorykeeper "github.com/neutron-org/neutron/v2/x/tokenfactory/keeper"
	transfer "github.com/neutron-org/neutron/v2/x/transfer/keeper"
)

// RegisterCustomPlugins returns wasmkeeper.Option that we can use to connect handlers for implemented custom queries and messages to the App
func RegisterCustomPlugins(
	ictxKeeper *interchaintransactionsmodulekeeper.Keeper,
	icqKeeper *interchainqueriesmodulekeeper.Keeper,
	transfer transfer.KeeperTransferWrapper,
	adminKeeper *adminmodulekeeper.Keeper,
	feeBurnerKeeper *feeburnerkeeper.Keeper,
	feeRefunderKeeper *feerefunderkeeper.Keeper,
	bank *bankkeeper.BaseKeeper,
	tfk *tokenfactorykeeper.Keeper,
	cronKeeper *cronkeeper.Keeper,
	contractmanagerKeeper *contractmanagerkeeper.Keeper,
	dexKeeper *dexkeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(ictxKeeper, icqKeeper, feeBurnerKeeper, feeRefunderKeeper, tfk, contractmanagerKeeper, dexKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messagePluginOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(ictxKeeper, icqKeeper, transfer, adminKeeper, bank, tfk, cronKeeper, contractmanagerKeeper, dexKeeper),
	)

	return []wasmkeeper.Option{
		queryPluginOpt,
		messagePluginOpt,
	}
}
