package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	cronkeeper "github.com/neutron-org/neutron/x/cron/keeper"
	crontypes "github.com/neutron-org/neutron/x/cron/types"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	tokenfactorykeeper "github.com/neutron-org/neutron/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/neutron-org/neutron/x/tokenfactory/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	icqKeeper icqkeeper.Keeper,
	cronKeeper cronkeeper.Keeper,
	tokenfactoryKeeper *tokenfactorykeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		// todo: FIXME
		icqKeeper.SetParams(ctx, icqtypes.DefaultParams())
		cronKeeper.SetParams(ctx, crontypes.DefaultParams())
		tokenfactoryKeeper.SetParams(ctx, tokenfactorytypes.DefaultParams())

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
