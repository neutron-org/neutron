package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	contractmanagerkeeper "github.com/neutron-org/neutron/v8/x/contractmanager/keeper"
	cronkeeper "github.com/neutron-org/neutron/v8/x/cron/keeper"
	dexkeeper "github.com/neutron-org/neutron/v8/x/dex/keeper"
	feeburnerkeeper "github.com/neutron-org/neutron/v8/x/feeburner/keeper"
	feerefunderkeeper "github.com/neutron-org/neutron/v8/x/feerefunder/keeper"
	tokenfactory2keeper "github.com/neutron-org/neutron/v8/x/tokenfactory2/keeper"

	adminmodulekeeper "github.com/cosmos/admin-module/v2/x/adminmodule/keeper"

	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	oraclekeeper "github.com/skip-mev/slinky/x/oracle/keeper"

	interchainqueriesmodulekeeper "github.com/neutron-org/neutron/v8/x/interchainqueries/keeper"
	interchaintransactionsmodulekeeper "github.com/neutron-org/neutron/v8/x/interchaintxs/keeper"
	tokenfactorykeeper "github.com/neutron-org/neutron/v8/x/tokenfactory/keeper"
	transfer "github.com/neutron-org/neutron/v8/x/transfer/keeper"
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
	tfk2 *tokenfactory2keeper.Keeper,
	cronKeeper *cronkeeper.Keeper,
	contractmanagerKeeper *contractmanagerkeeper.Keeper,
	dexKeeper *dexkeeper.Keeper,
	oracleKeeper *oraclekeeper.Keeper,
	markemapKeeper *marketmapkeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(ictxKeeper, icqKeeper, feeBurnerKeeper, feeRefunderKeeper, tfk, tfk2, contractmanagerKeeper, dexKeeper, oracleKeeper, markemapKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messagePluginOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(ictxKeeper, icqKeeper, transfer, adminKeeper, bank, tfk, tfk2, cronKeeper, contractmanagerKeeper, dexKeeper),
	)

	return []wasmkeeper.Option{
		queryPluginOpt,
		messagePluginOpt,
	}
}
