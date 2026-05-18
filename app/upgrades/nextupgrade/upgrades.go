package nextupgrade

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/neutron-org/neutron/v11/app/upgrades"
	dexkeeper "github.com/neutron-org/neutron/v11/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v11/x/dex/types"
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

		ctx.Logger().Info("Reconstructing tranche keys...")
		if err := ReconstructTrancheKeys(ctx, cdc, *keepers.DexKeeper); err != nil {
			return vm, err
		}
		ctx.Logger().Info("Done")

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func ReconstructTrancheKeys(ctx sdk.Context, cdc codec.Codec, k dexkeeper.Keeper) error {
	if err := reconstructLoExpirations(ctx, k); err != nil {
		return fmt.Errorf("failed to reconstruct LO expirations: %w", err)
	}

	if err := reconstructLoTranches(ctx, cdc, k); err != nil {
		return fmt.Errorf("failed to reconstruct LO tranches: %w", err)
	}

	if err := reconstructInactiveLoTranches(ctx, cdc, k); err != nil {
		return fmt.Errorf("failed to reconstruct inactive LO tranches: %w", err)
	}

	if err := reconstructLoTrancheUserLists(ctx, k); err != nil {
		return fmt.Errorf("failed to reconstruct LO tranche user lists: %w", err)
	}

	return nil
}

func reconstructLoExpirations(ctx sdk.Context, k dexkeeper.Keeper) error {
	allExpirations := k.GetAllLimitOrderExpiration(ctx) // total count varies but is expected to be small or even 0

	expirationsToRemove := make([]dextypes.LimitOrderExpiration, 0)
	expirationsToUpdate := make([]dextypes.LimitOrderExpiration, 0)
	for _, expiration := range allExpirations {
		tranche, found := k.GetLimitOrderTrancheByKey(ctx, expiration.TrancheRef)
		if !found {
			return fmt.Errorf("limit order tranche not found for expiration.TrancheRef %s", expiration.TrancheRef)
		}

		if !strings.HasPrefix(tranche.Key.TrancheKey, "tk-") {
			continue
		}

		expirationsToRemove = append(expirationsToRemove, *expiration)

		trancheIdxStr := strings.TrimPrefix(tranche.Key.TrancheKey, "tk-")
		trancheIdx, err := strconv.ParseUint(trancheIdxStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse tranche idx %s: %w", trancheIdxStr, err)
		}
		tranche.Key.TrancheKey = dextypes.NewTrancheKey(trancheIdx)
		expirationsToUpdate = append(expirationsToUpdate, *dexkeeper.NewLimitOrderExpiration(tranche))
	}

	if len(expirationsToRemove) != len(expirationsToUpdate) {
		return fmt.Errorf("mismatch in LO expirations to remove and update counts: %d != %d", len(expirationsToRemove), len(expirationsToUpdate))
	}

	for _, expiration := range expirationsToRemove {
		k.RemoveLimitOrderExpiration(ctx, expiration.ExpirationTime, expiration.TrancheRef)
	}
	for _, expiration := range expirationsToUpdate {
		k.SetLimitOrderExpiration(ctx, &expiration)
	}
	ctx.Logger().Info("LO expiration keys reconstructed", "count", len(expirationsToUpdate))

	return nil
}

func reconstructLoTranches(ctx sdk.Context, cdc codec.Codec, k dexkeeper.Keeper) error {
	tickLiquidities := k.GetAllTickLiquidity(ctx) // there are only 600-ish entries, so getting all is fine

	loTrancheKeysToRemove := make([]dextypes.LimitOrderTrancheKey, 0)
	loTranchesToUpdate := make([]dextypes.LimitOrderTranche, 0)
	for _, tickLiquidity := range tickLiquidities {
		if loTranche := tickLiquidity.GetLimitOrderTranche(); loTranche != nil {
			if !strings.HasPrefix(loTranche.Key.TrancheKey, "tk-") {
				continue
			}

			loTrancheKeysToRemove = append(loTrancheKeysToRemove, *loTranche.Key)

			trancheIdxStr := strings.TrimPrefix(loTranche.Key.TrancheKey, "tk-")
			trancheIdx, err := strconv.ParseUint(trancheIdxStr, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse tranche idx %s: %w", trancheIdxStr, err)
			}
			loTranche.Key.TrancheKey = dextypes.NewTrancheKey(trancheIdx)
			loTranchesToUpdate = append(loTranchesToUpdate, *loTranche)
		}
	}

	if len(loTrancheKeysToRemove) != len(loTranchesToUpdate) {
		return fmt.Errorf("mismatch in LO tranches to remove and update counts: %d != %d", len(loTrancheKeysToRemove), len(loTranchesToUpdate))
	}

	for _, loTrancheKey := range loTrancheKeysToRemove {
		k.RemoveLimitOrderTranche(ctx, &loTrancheKey)
	}
	for _, loTranche := range loTranchesToUpdate {
		k.SetLimitOrderTranche(ctx, &loTranche)
	}
	ctx.Logger().Info("LO tranche keys reconstructed", "count", len(loTranchesToUpdate))

	return nil
}

func reconstructInactiveLoTranches(ctx sdk.Context, cdc codec.Codec, k dexkeeper.Keeper) error {
	iter := k.GetInactiveLimitOrderTrancheIterator(ctx) // there are more than 400k entries -> iterating

	inactiveKeysToRemove := make([]dextypes.LimitOrderTrancheKey, 0)
	inactiveTranchesToUpdate := make([]dextypes.LimitOrderTranche, 0)
	for ; iter.Valid(); iter.Next() {
		var tranche dextypes.LimitOrderTranche
		cdc.MustUnmarshal(iter.Value(), &tranche)

		if !strings.HasPrefix(tranche.Key.TrancheKey, "tk-") {
			continue
		}

		inactiveKeysToRemove = append(inactiveKeysToRemove, *tranche.Key)

		trancheIdxStr := strings.TrimPrefix(tranche.Key.TrancheKey, "tk-")
		trancheIdx, err := strconv.ParseUint(trancheIdxStr, 10, 64)
		if err != nil {
			iter.Close() //nolint:errcheck
			return fmt.Errorf("failed to parse tranche idx %s: %w", trancheIdxStr, err)
		}
		tranche.Key.TrancheKey = dextypes.NewTrancheKey(trancheIdx)
		inactiveTranchesToUpdate = append(inactiveTranchesToUpdate, tranche)
	}
	iter.Close() //nolint:errcheck

	if len(inactiveKeysToRemove) != len(inactiveTranchesToUpdate) {
		return fmt.Errorf("mismatch in inactive LO tranches to remove and update counts: %d != %d", len(inactiveKeysToRemove), len(inactiveTranchesToUpdate))
	}

	for _, key := range inactiveKeysToRemove {
		k.RemoveInactiveLimitOrderTranche(ctx, &key)
	}
	for _, tranche := range inactiveTranchesToUpdate {
		k.SetInactiveLimitOrderTranche(ctx, &tranche)
	}
	ctx.Logger().Info("inactive LO tranche keys reconstructed", "count", len(inactiveTranchesToUpdate))

	return nil
}

func reconstructLoTrancheUserLists(ctx sdk.Context, k dexkeeper.Keeper) error {
	allUsers := k.GetAllLimitOrderTrancheUser(ctx) // there are only 300-ish entries, so getting all is fine

	usersToRemove := make([]dextypes.LimitOrderTrancheUser, 0)
	usersToUpdate := make([]dextypes.LimitOrderTrancheUser, 0)
	for _, user := range allUsers {
		if !strings.HasPrefix(user.TrancheKey, "tk-") {
			continue
		}

		usersToRemove = append(usersToRemove, *user)

		trancheIdxStr := strings.TrimPrefix(user.TrancheKey, "tk-")
		trancheIdx, err := strconv.ParseUint(trancheIdxStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse tranche idx %s: %w", trancheIdxStr, err)
		}
		user.TrancheKey = dextypes.NewTrancheKey(trancheIdx)
		usersToUpdate = append(usersToUpdate, *user)
	}

	if len(usersToRemove) != len(usersToUpdate) {
		return fmt.Errorf("mismatch in LO tranche user keys to remove and update counts: %d != %d", len(usersToRemove), len(usersToUpdate))
	}

	for _, user := range usersToRemove {
		k.RemoveLimitOrderTrancheUser(ctx, &user)
	}
	for _, user := range usersToUpdate {
		k.SetLimitOrderTrancheUser(ctx, &user)
	}
	ctx.Logger().Info("LO tranche user keys reconstructed", "count", len(usersToUpdate))

	return nil
}
