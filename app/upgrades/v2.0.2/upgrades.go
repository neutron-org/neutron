package v202

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/neutron-org/neutron/v2/app/upgrades"
)

func CreateUpgradeHandler(
	_ *module.Manager,
	_ module.Configurator,
	_ *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info(fmt.Sprintf("Empty migration {%s} applied", UpgradeName))
		return vm, nil
	}
}
