package v601

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/neutron-org/neutron/v8/app/upgrades"
	dexkeeper "github.com/neutron-org/neutron/v8/x/dex/keeper"
	"github.com/neutron-org/neutron/v8/x/dex/types"
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

		ctx.Logger().Info("Running dex upgrades...")
		err = UpgradeRemoveOrphanedLimitOrders(ctx, cdc, *keepers.DexKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func UpgradeRemoveOrphanedLimitOrders(ctx sdk.Context, cdc codec.Codec, k dexkeeper.Keeper) error {
	allLimitOrderExpirations := k.GetAllLimitOrderExpiration(ctx)

	expirationTrancheKeys := make(map[string]bool)
	for _, limitOrderExpiration := range allLimitOrderExpirations {
		expirationTrancheKeys[string(limitOrderExpiration.TrancheRef)] = true
	}

	tickLiquidityIterator := k.GetTickLiquidityIterator(ctx)

	defer tickLiquidityIterator.Close()

	tranchesToRemove := make([]*types.LimitOrderTranche, 0)
	for ; tickLiquidityIterator.Valid(); tickLiquidityIterator.Next() {
		var tickLiquidity types.TickLiquidity
		cdc.MustUnmarshal(tickLiquidityIterator.Value(), &tickLiquidity)

		if tickLiquidity.GetLimitOrderTranche() != nil && tickLiquidity.GetLimitOrderTranche().HasExpiration() {
			tranche := tickLiquidity.GetLimitOrderTranche()
			// If tranche is expiring and does not have an expiration record
			// then it is an orphaned tranche and should be removed
			if _, ok := expirationTrancheKeys[string(tranche.Key.KeyMarshal())]; !ok {
				tranchesToRemove = append(tranchesToRemove, tranche)
			}
		}
	}

	for _, tranche := range tranchesToRemove {
		// Set the orphaned tranche to inactive
		k.SetInactiveLimitOrderTranche(ctx, tranche)
		k.RemoveLimitOrderTranche(ctx, tranche.Key)
		ctx.EventManager().EmitEvents(types.GetEventsDecExpiringOrders(tranche.Key.TradePairId))
	}

	return nil
}
