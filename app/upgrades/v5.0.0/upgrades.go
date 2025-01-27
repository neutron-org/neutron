package v500

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"

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

		err = upgradePools(ctx, *keepers.DexKeeper)
		if err != nil {
			return nil, err
		}

		err = setMarketMapParams(ctx, keepers.MarketmapKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Running ibc-rate-limit upgrades...")

		err = upgradeIbcRateLimitSetContract(ctx, *keepers.IbcRateLimitKeeper)
		if err != nil {
			return nil, err
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

func upgradePools(ctx sdk.Context, k dexkeeper.Keeper) error {
	// Due to an issue with autoswap logic any pools with multiple shareholders must be withdrawn to ensure correct accounting
	ctx.Logger().Info("Migrating Pools...")

	allSharesholders := k.GetAllPoolShareholders(ctx)

	for poolID, shareholders := range allSharesholders {
		if len(shareholders) > 1 {
			pool, found := k.GetPoolByID(ctx, poolID)
			if !found {
				return fmt.Errorf("cannot find pool with ID %d", poolID)
			}
			for _, shareholder := range shareholders {
				addr := sdk.MustAccAddressFromBech32(shareholder.Address)
				pairID := pool.LowerTick0.Key.TradePairId.MustPairID()
				tick := pool.CenterTickIndexToken1()
				fee := pool.Fee()
				nShares := shareholder.Shares

				reserve0Removed, reserve1Removed, sharesBurned, err := k.WithdrawCore(ctx, pairID, addr, addr, []math.Int{nShares}, []int64{tick}, []uint64{fee})
				if err != nil {
					return fmt.Errorf("user %s failed to withdraw from pool %d", addr, poolID)
				}

				ctx.Logger().Info(
					"Withdrew user from pool",
					"User", addr.String(),
					"Pool", poolID,
					"SharesBurned", sharesBurned.String(),
					"Reserve0Withdrawn", reserve0Removed.String(),
					"Reserve1Withdrawn", reserve1Removed.String(),
				)

			}
		}
	}

	ctx.Logger().Info("Finished migrating Pools...")

	return nil
}

func upgradeIbcRateLimitSetContract(ctx sdk.Context, k ibcratelimitkeeper.Keeper) error {
	// Set the dex to paused
	ctx.Logger().Info("Setting ibc rate limiting contract...")

	switch ctx.ChainID() {
	case "neutron-1":
		if err := k.SetParams(ctx, ibcratelimittypes.Params{ContractAddress: MainnetRateLimitContract}); err != nil {
			return err
		}
	case "pion-1":
		if err := k.SetParams(ctx, ibcratelimittypes.Params{ContractAddress: TestnetRateLimitContract}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown chain id %s", ctx.ChainID())
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
