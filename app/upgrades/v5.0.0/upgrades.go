package v500

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	marketmapkeeper "github.com/skip-mev/connect/v2/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/connect/v2/x/marketmap/types"

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

		err = setMarketMapParams(ctx, keepers.MarketmapKeeper)
		if err != nil {
			return nil, err
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

func setMarketMapParams(ctx sdk.Context, marketmapKeeper *marketmapkeeper.Keeper) error {
	marketmapParams := marketmaptypes.Params{
		MarketAuthorities: []string{authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(), MarketMapAuthorityMultisig},
		Admin:             authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	}
	return marketmapKeeper.SetParams(ctx, marketmapParams)
}
