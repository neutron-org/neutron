package wasmbinding

import (
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	icacontrollerkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
)

type QueryPlugin struct {
	icaControllerKeeper *icacontrollerkeeper.Keeper
	icqKeeper           *icqkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(icaControllerKeeper *icacontrollerkeeper.Keeper, icqKeeper *icqkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{icaControllerKeeper: icaControllerKeeper, icqKeeper: icqKeeper}
}
