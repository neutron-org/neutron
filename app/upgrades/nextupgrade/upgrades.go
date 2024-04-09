package nextupgrade

import (
	"context"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"

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

		migrateAllowedClients(ctx, keepers.ClientsKeeper)
		ctx.Logger().Info(fmt.Sprintf("Added {%s} client to allowed clients list", ibcwasmtypes.Wasm))

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func migrateAllowedClients(ctx sdk.Context, clientKeeper clientkeeper.Keeper) {
	// explicitly update the IBC 02-client params, adding the wasm client type
	params := clientKeeper.GetParams(ctx)
	params.AllowedClients = append(params.AllowedClients, ibcwasmtypes.Wasm)
	clientKeeper.SetParams(ctx, params)
}
