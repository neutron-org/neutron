package v500

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/neutron-org/neutron/v5/app/upgrades"
	dexkeeper "github.com/neutron-org/neutron/v5/x/dex/keeper"
	ibcratelimitkeeper "github.com/neutron-org/neutron/v5/x/ibc-rate-limit/keeper"
	ibcratelimittypes "github.com/neutron-org/neutron/v5/x/ibc-rate-limit/types"
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

		ctx.Logger().Info("Running dex upgrades...")
		// Only pause dex for mainnet
		if ctx.ChainID() == "neutron-1" {
			err = upgradeDexPause(ctx, *keepers.DexKeeper)
			if err != nil {
				return nil, err
			}
		}

		ctx.Logger().Info("Running ibc-rate-limit upgrades...")
		// Only set rate limit contract for mainnet
		if ctx.ChainID() == "neutron-1" {
			err = upgradeIbcRateLimitSetContract(ctx, *keepers.IbcRateLimitKeeper)
			if err != nil {
				return nil, err
			}
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func upgradeDexPause(ctx sdk.Context, k dexkeeper.Keeper) error {
	// Set the dex to paused
	ctx.Logger().Info("Pausing dex...")

	params := k.GetParams(ctx)
	params.Paused = true

	if err := k.SetParams(ctx, params); err != nil {
		return err
	}

	ctx.Logger().Info("Dex is paused")

	return nil
}

func upgradeIbcRateLimitSetContract(ctx sdk.Context, k ibcratelimitkeeper.Keeper) error {
	// Set the dex to paused
	ctx.Logger().Info("Setting ibc rate limiting contract...")

	if err := k.SetParams(ctx, ibcratelimittypes.Params{ContractAddress: RateLimitContract}); err != nil {
		return err
	}

	ctx.Logger().Info("Rate limit contract is set")

	return nil
}
