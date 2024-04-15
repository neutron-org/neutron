package wasmbinding

import (
	contractmanagerkeeper "github.com/neutron-org/neutron/v3/x/contractmanager/keeper"
	dexkeeper "github.com/neutron-org/neutron/v3/x/dex/keeper"
	feeburnerkeeper "github.com/neutron-org/neutron/v3/x/feeburner/keeper"
	feerefunderkeeper "github.com/neutron-org/neutron/v3/x/feerefunder/keeper"
	icqkeeper "github.com/neutron-org/neutron/v3/x/interchainqueries/keeper"
	icacontrollerkeeper "github.com/neutron-org/neutron/v3/x/interchaintxs/keeper"

	tokenfactorykeeper "github.com/neutron-org/neutron/v3/x/tokenfactory/keeper"

	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	oraclekeeper "github.com/skip-mev/slinky/x/oracle/keeper"
)

type QueryPlugin struct {
	icaControllerKeeper   *icacontrollerkeeper.Keeper
	icqKeeper             *icqkeeper.Keeper
	feeBurnerKeeper       *feeburnerkeeper.Keeper
	feeRefunderKeeper     *feerefunderkeeper.Keeper
	tokenFactoryKeeper    *tokenfactorykeeper.Keeper
	contractmanagerKeeper *contractmanagerkeeper.Keeper
	dexKeeper             *dexkeeper.Keeper
	oracleKeeper          *oraclekeeper.Keeper
	marketmapKeeper       *marketmapkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(icaControllerKeeper *icacontrollerkeeper.Keeper, icqKeeper *icqkeeper.Keeper, feeBurnerKeeper *feeburnerkeeper.Keeper, feeRefunderKeeper *feerefunderkeeper.Keeper, tfk *tokenfactorykeeper.Keeper, contractmanagerKeeper *contractmanagerkeeper.Keeper, dexKeeper *dexkeeper.Keeper, oracleKeeper *oraclekeeper.Keeper, marketmapKeeper *marketmapkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		icaControllerKeeper:   icaControllerKeeper,
		icqKeeper:             icqKeeper,
		feeBurnerKeeper:       feeBurnerKeeper,
		feeRefunderKeeper:     feeRefunderKeeper,
		tokenFactoryKeeper:    tfk,
		contractmanagerKeeper: contractmanagerKeeper,
		dexKeeper:             dexKeeper,
		oracleKeeper:          oracleKeeper,
		marketmapKeeper:       marketmapKeeper,
	}
}
