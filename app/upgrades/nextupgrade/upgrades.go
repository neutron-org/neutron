package nextupgrade

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/neutron-org/neutron/v2/app/upgrades"
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
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		migrateICSOutstandingDowntime(ctx, keepers)

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func migrateICSOutstandingDowntime(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) {
	ctx.Logger().Info("Migrating ICS outstanding downtime...")

	// TODO: discuss the following code with ICS team
	downtimes := keepers.ConsumerKeeper.GetAllOutstandingDowntimes(ctx)
	for _, od := range downtimes {
		keepers.ConsumerKeeper.DeleteOutstandingDowntime(ctx, od.ValidatorConsensusAddress)
	}

	ctx.Logger().Info("Finished ICS outstanding downtime")
}
