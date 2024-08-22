package v5

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dex/types"
)

type DexKeeperI interface {
	GetAllPoolShareholders(sdk.Context) map[uint64][]types.PoolShareholder
	GetPoolByID(sdk.Context, uint64) (*types.Pool, bool)
	WithdrawCore(context.Context, *types.PairID, sdk.AccAddress, sdk.AccAddress, []math.Int, []int64, []uint64) error
}

// MigrateStore performs in-place store migrations.
func MigrateStore(ctx sdk.Context, k DexKeeperI) error {
	if err := migratePools(ctx, k); err != nil {
		return err
	}

	return nil
}

func migratePools(ctx sdk.Context, k DexKeeperI) error {
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

				err := k.WithdrawCore(ctx, pairID, addr, addr, []math.Int{nShares}, []int64{tick}, []uint64{fee})
				if err != nil {
					return fmt.Errorf("user %s failed to withdraw from pool %d", addr, poolID)
				}

				ctx.Logger().Info(
					"Withdrew user from pool",
					"User", addr.String(),
					"Pool", poolID,
					"Shares", nShares.String(),
				)

			}
		}
	}

	ctx.Logger().Info("Finished migrating Pools...")

	return nil
}
