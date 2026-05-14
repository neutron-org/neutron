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
	"github.com/neutron-org/neutron/v10/app/upgrades"
	dexkeeper "github.com/neutron-org/neutron/v10/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v10/x/dex/types"
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

	for _, loTrancheKey := range loTrancheKeysToRemove {
		k.RemoveLimitOrderTranche(ctx, &loTrancheKey)
	}
	for _, loTranche := range loTranchesToUpdate {
		k.SetLimitOrderTranche(ctx, &loTranche)
	}

	return nil
}

func reconstructInactiveLoTranches(ctx sdk.Context, cdc codec.Codec, k dexkeeper.Keeper) error {
	iter := k.GetInactiveLimitOrderTrancheIterator(ctx)

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

	for _, key := range inactiveKeysToRemove {
		k.RemoveInactiveLimitOrderTranche(ctx, &key)
	}
	for i := range inactiveTranchesToUpdate {
		k.SetInactiveLimitOrderTranche(ctx, &inactiveTranchesToUpdate[i])
	}

	return nil
}

func reconstructLoTrancheUserLists(ctx sdk.Context, k dexkeeper.Keeper) error {
	allUsers := k.GetAllLimitOrderTrancheUser(ctx) // there are only 300-ish entries, so getting all is fine

	// Each LimitOrderTrancheUser has its TrancheKey embedded in both the KV store key
	// (address + trancheKey) and the serialised value. It is required to remove the old entry and
	// write a new one under the updated key.
	type userRemoveKey struct {
		address    string
		trancheKey string
	}

	keysToRemove := make([]userRemoveKey, 0)
	usersToUpdate := make([]*dextypes.LimitOrderTrancheUser, 0)

	for _, user := range allUsers {
		if !strings.HasPrefix(user.TrancheKey, "tk-") {
			continue
		}

		// Snapshot address + old key before mutation.
		keysToRemove = append(keysToRemove, userRemoveKey{
			address:    user.Address,
			trancheKey: user.TrancheKey,
		})

		trancheIdxStr := strings.TrimPrefix(user.TrancheKey, "tk-")
		trancheIdx, err := strconv.ParseUint(trancheIdxStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse tranche idx %s: %w", trancheIdxStr, err)
		}
		user.TrancheKey = dextypes.NewTrancheKey(trancheIdx)
		usersToUpdate = append(usersToUpdate, user)
	}

	for _, key := range keysToRemove {
		k.RemoveLimitOrderTrancheUserByKey(ctx, key.trancheKey, key.address)
	}
	for _, user := range usersToUpdate {
		k.SetLimitOrderTrancheUser(ctx, user)
	}

	return nil
}
