package v800

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/neutron-org/neutron/v8/app/upgrades"
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

		ctx.Logger().Info("Running tokenfactory upgrades...")
		err = UpgradeDenomsMetadata(ctx, keepers.BankKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func UpgradeDenomsMetadata(ctx sdk.Context, bk bankkeeper.Keeper) error {
	allDenomMetadata := bk.GetAllDenomMetaData(ctx)

	for _, metadata := range allDenomMetadata {
		denom := metadata.Base

		if metadata.Name == "" {
			metadata.Name = denom
		}
		if metadata.Symbol == "" {
			metadata.Symbol = denom
		}
		if metadata.Display == "" {
			metadata.Display = denom
		}

		bk.SetDenomMetaData(ctx, metadata)
	}

	return nil
}
