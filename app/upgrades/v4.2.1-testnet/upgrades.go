package v421testnet

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	math_utils "github.com/neutron-org/neutron/v4/utils/math"

	"cosmossdk.io/math"

	"github.com/neutron-org/neutron/v4/app/upgrades"
	dexkeeper "github.com/neutron-org/neutron/v4/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
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

		// Only run this migration for testnet.
		// NEUTRON-1 HAS ALREADY BEEN UPGRADED AND RE-RUNNING THE MIGRATION WOULD BREAK STATE!
		if ctx.ChainID() == "pion-1" {
			migrateLimitOrderTrancheAccounting(ctx, *keepers.DexKeeper)
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

type TrancheData struct {
	SharesWithdrawn math.Int
	Tranche         dextypes.LimitOrderTranche
}

func fetchTrancheByTrancheUser(
	ctx sdk.Context,
	k dexkeeper.Keeper,
	trancheUser *dextypes.LimitOrderTrancheUser,
) (*dextypes.LimitOrderTranche, bool) {
	key := dextypes.LimitOrderTrancheKey{
		TickIndexTakerToMaker: trancheUser.TickIndexTakerToMaker,
		TradePairId:           trancheUser.TradePairId,
		TrancheKey:            trancheUser.TrancheKey,
	}

	return k.GetLimitOrderTrancheByKey(ctx, key.KeyMarshal())
}

// Due to a bug in our limitorder accounting we must update the TotalTakerDenom to ensure correct accounting for withdrawals
func migrateLimitOrderTrancheAccounting(ctx sdk.Context, k dexkeeper.Keeper) {
	ctx.Logger().Info("Migrating LimitOrderTranches...")

	trancheUpdateData := make(map[string]TrancheData)
	trancheUsersToDelete := make([]*dextypes.LimitOrderTrancheUser, 0)

	allTrancheUsers := k.GetAllLimitOrderTrancheUser(ctx)

	// Iterate through all limitOrderTrancheUsers
	// Sum up the total SharesWithdrawn for each tranche
	for _, trancheUser := range allTrancheUsers {
		val, ok := trancheUpdateData[trancheUser.TrancheKey]
		if !ok {
			tranche, found := fetchTrancheByTrancheUser(ctx, k, trancheUser)

			// Due to an earlier error in our rounding behavior / tranche accounting
			// there are trancheUsers that have dust amounts of un-withdrawn shares but the tranche has already been deleted
			// FWIW this issue only exists on pion-1
			// We can safely delete these trancheUsers
			if !found {
				trancheUsersToDelete = append(trancheUsersToDelete, trancheUser)
				continue
			}

			trancheUpdateData[trancheUser.TrancheKey] = TrancheData{SharesWithdrawn: trancheUser.SharesWithdrawn, Tranche: *tranche}
		} else {
			newVal := TrancheData{
				SharesWithdrawn: val.SharesWithdrawn.Add(trancheUser.SharesWithdrawn),
				Tranche:         val.Tranche,
			}
			trancheUpdateData[trancheUser.TrancheKey] = newVal
		}
	}

	// Update tranches
	for _, trancheData := range trancheUpdateData {
		tranche := trancheData.Tranche
		// Calculate total sharesWithdrawn from tranche (denominated in TakerDenom)
		sharesWithdrawnTaker := math_utils.NewPrecDecFromInt(trancheData.SharesWithdrawn).Quo(trancheData.Tranche.PriceTakerToMaker)
		// TotalTakerDenom = ReservesTakerDenom + SharesWithdrawn
		newTotalTakerDenom := sharesWithdrawnTaker.TruncateInt().Add(tranche.ReservesTakerDenom)
		// Update Tranche
		tranche.TotalTakerDenom = newTotalTakerDenom

		// Save new limit order tranche
		k.SetLimitOrderTranche(ctx, &tranche)

	}

	// Delete the orphaned trancheUsers
	for _, trancheUser := range trancheUsersToDelete {
		k.RemoveLimitOrderTrancheUser(ctx, trancheUser)
	}
}
