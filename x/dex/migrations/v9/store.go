package v9

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dextypes "github.com/neutron-org/neutron/v11/x/dex/types"
)

// dexKeeper defines an interface with dex keeper methods required for the migration. It is defined
// to avoid import loop (x/dex/migrations <-> x/dex/keeper).
type dexKeeper interface {
	GetAllLimitOrderExpiration(ctx sdk.Context) (list []*dextypes.LimitOrderExpiration)
	GetLimitOrderTrancheByKey(ctx sdk.Context, key []byte) (tranche *dextypes.LimitOrderTranche, found bool)
	RemoveLimitOrderExpiration(ctx sdk.Context, goodTilDate time.Time, trancheRef []byte)
	SetLimitOrderExpiration(ctx sdk.Context, goodTilRecord *dextypes.LimitOrderExpiration)
	GetAllTickLiquidity(ctx sdk.Context) (list []*dextypes.TickLiquidity)
	RemoveLimitOrderTranche(ctx sdk.Context, trancheKey *dextypes.LimitOrderTrancheKey)
	SetLimitOrderTranche(ctx sdk.Context, tranche *dextypes.LimitOrderTranche)
	GetInactiveLimitOrderTrancheIterator(ctx sdk.Context) storetypes.Iterator
	RemoveInactiveLimitOrderTranche(ctx sdk.Context, limitOrderTrancheKey *dextypes.LimitOrderTrancheKey)
	SetInactiveLimitOrderTranche(ctx sdk.Context, limitOrderTranche *dextypes.LimitOrderTranche)
	GetAllLimitOrderTrancheUser(ctx sdk.Context) (list []*dextypes.LimitOrderTrancheUser)
	RemoveLimitOrderTrancheUser(ctx sdk.Context, trancheUser *dextypes.LimitOrderTrancheUser)
	SetLimitOrderTrancheUser(ctx sdk.Context, limitOrderTrancheUser *dextypes.LimitOrderTrancheUser)
}

// MigrateStore performs in-place store migrations. It reconstructs the tranche keys for limit order
// expirations, tranches, inactive tranches, and tranche user lists.
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, dexKeeper dexKeeper) error {
	ctx.Logger().Info("Starting dex store migration...")

	ctx.Logger().Info("Reconstructing tranche keys...")
	if err := ReconstructTrancheKeys(ctx, cdc, dexKeeper); err != nil {
		return err
	}

	ctx.Logger().Info("Dex store migration completed")
	return nil
}

func ReconstructTrancheKeys(ctx sdk.Context, cdc codec.BinaryCodec, k dexKeeper) error {
	ctx.Logger().Info("Reconstructing LO expirations...")
	if err := reconstructLoExpirations(ctx, k); err != nil {
		return fmt.Errorf("failed to reconstruct LO expirations: %w", err)
	}
	ctx.Logger().Info("Done")

	ctx.Logger().Info("Reconstructing LO tranches...")
	if err := reconstructLoTranches(ctx, k); err != nil {
		return fmt.Errorf("failed to reconstruct LO tranches: %w", err)
	}
	ctx.Logger().Info("Done")

	ctx.Logger().Info("Reconstructing inactive LO tranches...")
	if err := reconstructInactiveLoTranches(ctx, cdc, k); err != nil {
		return fmt.Errorf("failed to reconstruct inactive LO tranches: %w", err)
	}
	ctx.Logger().Info("Done")

	ctx.Logger().Info("Reconstructing LO tranche user lists...")
	if err := reconstructLoTrancheUserLists(ctx, k); err != nil {
		return fmt.Errorf("failed to reconstruct LO tranche user lists: %w", err)
	}
	ctx.Logger().Info("Done")

	return nil
}

func reconstructLoExpirations(ctx sdk.Context, k dexKeeper) error {
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
		expirationsToUpdate = append(expirationsToUpdate, *dextypes.NewLimitOrderExpiration(tranche))
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

func reconstructLoTranches(ctx sdk.Context, k dexKeeper) error {
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

func reconstructInactiveLoTranches(ctx sdk.Context, cdc codec.BinaryCodec, k dexKeeper) error {
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
			iter.Close() //nolint:errcheck,gosec
			return fmt.Errorf("failed to parse tranche idx %s: %w", trancheIdxStr, err)
		}
		tranche.Key.TrancheKey = dextypes.NewTrancheKey(trancheIdx)
		inactiveTranchesToUpdate = append(inactiveTranchesToUpdate, tranche)
	}
	iter.Close() //nolint:errcheck,gosec

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

func reconstructLoTrancheUserLists(ctx sdk.Context, k dexKeeper) error {
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
