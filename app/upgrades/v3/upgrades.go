package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/neutron-org/neutron/app/upgrades"

	crontypes "github.com/neutron-org/neutron/x/cron/types"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	tokenfactorytypes "github.com/neutron-org/neutron/x/tokenfactory/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		// todo: FIXME
		err := keepers.IcqKeeper.SetParams(ctx, icqtypes.DefaultParams())
		if err != nil {
			return vm, err
		}

		err = keepers.CronKeeper.SetParams(ctx, crontypes.DefaultParams())
		if err != nil {
			return vm, err
		}

		err = keepers.TokenFactoryKeeper.SetParams(ctx, tokenfactorytypes.DefaultParams())
		if err != nil {
			return vm, err
		}

		vm, err = mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
