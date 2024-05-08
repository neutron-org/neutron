package v400

import (
	"context"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"fmt"
	adminmoduletypes "github.com/cosmos/admin-module/x/adminmodule/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"

	"github.com/neutron-org/neutron/v4/app/upgrades"
	slinkyutils "github.com/neutron-org/neutron/v4/utils/slinky"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
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

		ctx.Logger().Info("Setting marketmap and oracle state...")
		err = setMarketState(ctx, keepers.MarketmapKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func setMarketMapParams(ctx sdk.Context, marketmapKeeper *marketmapkeeper.Keeper) error {
	marketmapParams := marketmaptypes.Params{
		MarketAuthorities: []string{authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String()},
		Admin:             authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	}
	return marketmapKeeper.SetParams(ctx, marketmapParams)
}

func setMarketState(ctx sdk.Context, mmKeeper *marketmapkeeper.Keeper) error {
	markets, err := slinkyutils.ReadMarketsFromFile("markets.json")
	if err != nil {
		return err
	}

	for _, market := range markets {
		err = mmKeeper.CreateMarket(ctx, market)
		if err != nil {
			return err
		}

		err = mmKeeper.Hooks().AfterMarketCreated(ctx, market)
	}
	return nil
}
