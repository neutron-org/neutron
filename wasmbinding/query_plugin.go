package wasmbinding

import icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"

type QueryPlugin struct {
	ICAControllerKeeper *icacontrollerkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(ICAControllerKeeper *icacontrollerkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{ICAControllerKeeper: ICAControllerKeeper}
}
