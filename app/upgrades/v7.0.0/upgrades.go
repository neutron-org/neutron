package v700

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/neutron-org/neutron/v7/app/upgrades"
	v601 "github.com/neutron-org/neutron/v7/app/upgrades/v6.0.1"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	cdc codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		if ctx.ChainID() == "pion-1" {
			// this migration was already applied on mainnet, we don't need to do that twice
			if err := v601.UpgradeRemoveOrphanedLimitOrders(ctx, cdc, *keepers.DexKeeper); err != nil {
				return vm, errors.Wrapf(err, "failed to remove orphaned limit orders")
			}
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}
