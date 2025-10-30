package wasmbinding

import (
	contractmanagerkeeper "github.com/neutron-org/neutron/v8/x/contractmanager/keeper"
	contractmanagertypes "github.com/neutron-org/neutron/v8/x/contractmanager/types"
	dexkeeper "github.com/neutron-org/neutron/v8/x/dex/keeper"
	feeburnerkeeper "github.com/neutron-org/neutron/v8/x/feeburner/keeper"
	feerefunderkeeper "github.com/neutron-org/neutron/v8/x/feerefunder/keeper"
	icqkeeper "github.com/neutron-org/neutron/v8/x/interchainqueries/keeper"
	icacontrollerkeeper "github.com/neutron-org/neutron/v8/x/interchaintxs/keeper"
	tokenfactory2keeper "github.com/neutron-org/neutron/v8/x/tokenfactory2/keeper"

	tokenfactorykeeper "github.com/neutron-org/neutron/v8/x/tokenfactory/keeper"

	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	oraclekeeper "github.com/skip-mev/slinky/x/oracle/keeper"
)

type QueryPlugin struct {
	icaControllerKeeper        *icacontrollerkeeper.Keeper
	icqKeeper                  *icqkeeper.Keeper
	feeBurnerKeeper            *feeburnerkeeper.Keeper
	feeRefunderKeeper          *feerefunderkeeper.Keeper
	tokenFactoryKeeper         *tokenfactorykeeper.Keeper
	tokenFactory2Keeper        *tokenfactory2keeper.Keeper
	contractmanagerQueryServer contractmanagertypes.QueryServer
	dexKeeper                  *dexkeeper.Keeper
	oracleKeeper               *oraclekeeper.Keeper
	marketmapKeeper            *marketmapkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(icaControllerKeeper *icacontrollerkeeper.Keeper, icqKeeper *icqkeeper.Keeper, feeBurnerKeeper *feeburnerkeeper.Keeper, feeRefunderKeeper *feerefunderkeeper.Keeper, tfk *tokenfactorykeeper.Keeper, tfk2 *tokenfactory2keeper.Keeper, contractmanagerKeeper *contractmanagerkeeper.Keeper, dexKeeper *dexkeeper.Keeper, oracleKeeper *oraclekeeper.Keeper, marketmapKeeper *marketmapkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		icaControllerKeeper:        icaControllerKeeper,
		icqKeeper:                  icqKeeper,
		feeBurnerKeeper:            feeBurnerKeeper,
		feeRefunderKeeper:          feeRefunderKeeper,
		tokenFactoryKeeper:         tfk,
		tokenFactory2Keeper:        tfk2,
		contractmanagerQueryServer: contractmanagerkeeper.NewQueryServerImpl(*contractmanagerKeeper),
		dexKeeper:                  dexKeeper,
		oracleKeeper:               oracleKeeper,
		marketmapKeeper:            marketmapKeeper,
	}
}
