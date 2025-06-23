package v800

import (
	"context"
	"fmt"

	feerefunderkeeper "github.com/neutron-org/neutron/v7/x/feerefunder/keeper"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/neutron-org/neutron/v7/app/upgrades"
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

		ctx.Logger().Info("Running feerefunder upgrades...")
		err = UpdateFeerefunderParams(ctx, keepers.FeerefunderKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func UpdateFeerefunderParams(ctx sdk.Context, fk *feerefunderkeeper.Keeper) error {
	params := fk.GetParams(ctx)
	params.FeeEnabled = true
	err := fk.SetParams(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update feerefunder params: %w", err)
	}

	return nil
}
