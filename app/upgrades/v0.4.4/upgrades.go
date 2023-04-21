package v044

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/neutron-org/neutron/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		ctx.Logger().Info("Migrating FeeBurner Params...")
		oldFeeBurnerParams := keepers.FeeBurnerKeeper.GetParams(ctx)
		oldFeeBurnerParams.TreasuryAddress = oldFeeBurnerParams.ReserveAddress

		keepers.FeeBurnerKeeper.SetParams(ctx, oldFeeBurnerParams)

		ctx.Logger().Info("Migrating SlashingKeeper Params...")
		oldSlashingParams := keepers.SlashingKeeper.GetParams(ctx)
		oldSlashingParams.SignedBlocksWindow = int64(36000)

		keepers.SlashingKeeper.SetParams(ctx, oldSlashingParams)

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
