package wasmbinding

import (
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
)

type QueryPlugin struct {
	icaControllerKeeper *icacontrollerkeeper.Keeper
	icqKeeper           *icqkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(icaControllerKeeper *icacontrollerkeeper.Keeper, icqKeeper *icqkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{icaControllerKeeper: icaControllerKeeper, icqKeeper: icqKeeper}
}
