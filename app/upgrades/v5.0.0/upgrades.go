package v500

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/neutron-org/neutron/v5/app/upgrades"
	dexkeeper "github.com/neutron-org/neutron/v5/x/dex/keeper"
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

		err = setDexParams(ctx, keepers.DexKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func setDexParams(ctx sdk.Context, dexKeeper *dexkeeper.Keeper) error {
	params := dexKeeper.GetParams(ctx)
	params.Paused = true
	err := dexKeeper.SetParams(ctx, params)
	if err != nil {
		return errors.Wrap(err, "failed to set dex params")
	}

	return nil
}
