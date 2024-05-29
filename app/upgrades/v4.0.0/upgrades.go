package v400

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	"github.com/neutron-org/neutron/v4/app/params"
	dynamicfeeskeeper "github.com/neutron-org/neutron/v4/x/dynamicfees/keeper"
	dynamicfeestypes "github.com/neutron-org/neutron/v4/x/dynamicfees/types"

	adminmoduletypes "github.com/cosmos/admin-module/x/adminmodule/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/neutron-org/neutron/v4/app/upgrades"
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

		ctx.Logger().Info("Setting dynamicfees/feemarket params...")
		err = setDynamicFeesParams(ctx, keepers.FeeMarketKeeper, keepers.DynamicfeesKeeper)
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

// TODO: add a test for the migrations: check that feemarket state is consistent with feemarket params
func setDynamicFeesParams(ctx sdk.Context, feemarketKeeper *feemarketkeeper.Keeper, dynamicfeesKeeper *dynamicfeeskeeper.Keeper) error {
	// TODO: set params values
	feemarketParams := feemarkettypes.Params{
		Alpha:                  math.LegacyDec{},
		Beta:                   math.LegacyDec{},
		Theta:                  math.LegacyDec{},
		Delta:                  math.LegacyDec{},
		MinBaseGasPrice:        math.LegacyDec{},
		MinLearningRate:        math.LegacyDec{},
		MaxLearningRate:        math.LegacyDec{},
		TargetBlockUtilization: 0,
		MaxBlockUtilization:    0,
		Window:                 0,
		FeeDenom:               "",
		Enabled:                false,
		DistributeFees:         false,
	}
	feemarketState := feemarkettypes.NewState(feemarketParams.Window, feemarketParams.MinBaseGasPrice, feemarketParams.MinLearningRate)
	err := feemarketKeeper.SetParams(ctx, feemarketParams)
	if err != nil {
		return errors.Wrap(err, "failed to to set feemarket params")
	}
	err = feemarketKeeper.SetState(ctx, feemarketState)
	if err != nil {
		return errors.Wrap(err, "failed to to set feemarket state")
	}

	dynamicfeesParams := dynamicfeestypes.Params{
		NtrnPrices: sdk.NewDecCoins(
			sdk.NewDecCoin(params.DefaultDenom, math.OneInt()),
		),
	}

	return dynamicfeesKeeper.SetParams(ctx, dynamicfeesParams)
}
