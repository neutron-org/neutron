package v044

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	feeburnertypes "github.com/neutron-org/neutron/x/feeburner/types"

	"github.com/neutron-org/neutron/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Migrating SlashingKeeper Params...")
		oldSlashingParams := keepers.SlashingKeeper.GetParams(ctx)
		oldSlashingParams.SignedBlocksWindow = int64(36000)

		keepers.SlashingKeeper.SetParams(ctx, oldSlashingParams)

		ctx.Logger().Info("Migrating FeeBurner Params...")

		s, ok := keepers.ParamsKeeper.GetSubspace(feeburnertypes.ModuleName)
		if !ok {
			panic("global fee burner params subspace not found")
		}
		var reserveAddress string
		s.Get(ctx, feeburnertypes.KeyReserveAddress, &reserveAddress)

		var neutronDenom string
		s.Get(ctx, feeburnertypes.KeyNeutronDenom, &neutronDenom)

		feeburnerDefaultParams := feeburnertypes.DefaultParams()
		ctx.Logger().Info("Default params", "params", feeburnerDefaultParams)
		feeburnerDefaultParams.TreasuryAddress = reserveAddress
		feeburnerDefaultParams.NeutronDenom = neutronDenom
		ctx.Logger().Info("Updted params", "params", feeburnerDefaultParams)

		keepers.FeeBurnerKeeper.GetParams(ctx)

		ctx.Logger().Info("Set params...")
		keepers.FeeBurnerKeeper.SetParams(ctx, feeburnerDefaultParams)

		panic("halt it")

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
