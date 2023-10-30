package wasmbinding

import (
	contractmanagerkeeper "github.com/neutron-org/neutron/x/contractmanager/keeper"
	feeburnerkeeper "github.com/neutron-org/neutron/x/feeburner/keeper"
	feerefunderkeeper "github.com/neutron-org/neutron/x/feerefunder/keeper"
	ibchooks "github.com/neutron-org/neutron/x/ibc-hooks"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	icacontrollerkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"

	tokenfactorykeeper "github.com/neutron-org/neutron/x/tokenfactory/keeper"
)

type QueryPlugin struct {
	icaControllerKeeper   *icacontrollerkeeper.Keeper
	icqKeeper             *icqkeeper.Keeper
	feeBurnerKeeper       *feeburnerkeeper.Keeper
	feeRefunderKeeper     *feerefunderkeeper.Keeper
	tokenFactoryKeeper    *tokenfactorykeeper.Keeper
	wasmHooks             *ibchooks.WasmHooks
	contractmanagerKeeper *contractmanagerkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(
	icaControllerKeeper *icacontrollerkeeper.Keeper,
	icqKeeper *icqkeeper.Keeper,
	feeBurnerKeeper *feeburnerkeeper.Keeper,
	feeRefunderKeeper *feerefunderkeeper.Keeper,
	tfk *tokenfactorykeeper.Keeper,
	wasmHooks *ibchooks.WasmHooks,
	contractmanagerKeeper *contractmanagerkeeper.Keeper,
) *QueryPlugin {
	return &QueryPlugin{
		icaControllerKeeper:   icaControllerKeeper,
		icqKeeper:             icqKeeper,
		feeBurnerKeeper:       feeBurnerKeeper,
		feeRefunderKeeper:     feeRefunderKeeper,
		tokenFactoryKeeper:    tfk,
		wasmHooks:             wasmHooks,
		contractmanagerKeeper: contractmanagerKeeper,
	}
}
