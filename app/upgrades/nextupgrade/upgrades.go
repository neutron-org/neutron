package nextupgrade

import (
	"context"
	"fmt"
	adminmoduletypes "github.com/cosmos/admin-module/x/adminmodule/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/neutron-org/neutron/v3/app/config"

	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/neutron-org/neutron/v3/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Setting marketmap params...")
		err = setMarketMapParams(ctx, keepers.MarketmapKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func setMarketMapParams(ctx sdk.Context, marketmapKeeper *marketmapkeeper.Keeper) error {
	config.GetDefaultConfig() // make sure we use neutron prefix for admin address
	params, err := marketmapKeeper.GetParams(ctx)
	if err != nil {
		return err
	}

	marketmapParams := marketmaptypes.Params{
		MarketAuthority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		Version:         params.Version,
	}
	return marketmapKeeper.SetParams(ctx, marketmapParams)
}
