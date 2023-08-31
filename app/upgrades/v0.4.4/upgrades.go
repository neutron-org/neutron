package v044

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	feeburnertypes "github.com/neutron-org/neutron/x/feeburner/types"
	tokenfactorytypes "github.com/neutron-org/neutron/x/tokenfactory/types"

	"github.com/neutron-org/neutron/app/upgrades"
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

		ctx.Logger().Info("Migrating SlashingKeeper Params...")
		oldSlashingParams := keepers.SlashingKeeper.GetParams(ctx)
		oldSlashingParams.SignedBlocksWindow = int64(36000)

		err = keepers.SlashingKeeper.SetParams(ctx, oldSlashingParams)
		if err != nil {
			return vm, err
		}
		ctx.Logger().Info("Migrating FeeBurner Params...")
		s, ok := keepers.ParamsKeeper.GetSubspace(feeburnertypes.ModuleName)
		if !ok {
			return nil, errors.New("global fee burner params subspace not found")
		}
		var reserveAddress string
		s.Get(ctx, feeburnertypes.KeyReserveAddress, &reserveAddress)

		var neutronDenom string
		s.Get(ctx, feeburnertypes.KeyNeutronDenom, &neutronDenom)

		feeburnerDefaultParams := feeburnertypes.DefaultParams()
		feeburnerDefaultParams.TreasuryAddress = reserveAddress
		feeburnerDefaultParams.NeutronDenom = neutronDenom
		err = keepers.FeeBurnerKeeper.SetParams(ctx, feeburnerDefaultParams)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Migrating TokenFactory Params...")
		tokenfactoryDefaultParams := tokenfactorytypes.DefaultParams()
		keepers.TokenFactoryKeeper.SetParams(ctx, tokenfactoryDefaultParams)

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
