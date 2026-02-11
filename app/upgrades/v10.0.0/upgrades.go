package v10_0_0

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/neutron-org/neutron/v10/app/upgrades"
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

		err = wasmBurnerPerms(c, keepers)

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func wasmBurnerPerms(c context.Context, k *upgrades.UpgradeKeepers) error {
	// read module acc from storage or create new one
	accI := k.AccountKeeper.GetModuleAccount(c, wasmtypes.ModuleName)
	moduleAcc, ok := accI.(*types.ModuleAccount)
	if !ok {
		return fmt.Errorf("account is not a ModuleAccount")
	}
	// update permissions
	moduleAcc.Permissions = []string{types.Burner}

	k.AccountKeeper.SetModuleAccount(c, moduleAcc)
	return nil
}
