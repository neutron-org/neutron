package wasmbinding

import (
	feeburnerkeeper "github.com/neutron-org/neutron/x/feeburner/keeper"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	icacontrollerkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
)

type QueryPlugin struct {
	icaControllerKeeper *icacontrollerkeeper.Keeper
	icqKeeper           *icqkeeper.Keeper
	feeBurnerKeeper     *feeburnerkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(icaControllerKeeper *icacontrollerkeeper.Keeper, icqKeeper *icqkeeper.Keeper, feeBurnerKeeper *feeburnerkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{icaControllerKeeper: icaControllerKeeper, icqKeeper: icqKeeper, feeBurnerKeeper: feeBurnerKeeper}
}
