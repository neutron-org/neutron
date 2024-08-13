package v5

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/types"
)

// MigrateStore performs in-place store migrations.
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	// Only run this migration for testnet.
	// NEUTRON-1 HAS ALREADY BEEN UPGRADED AND RE-RUNNING THE MIGRATION WOULD BREAK STATE!
	if ctx.ChainID() == "pion-1" {
		if err := migrateLimitOrderTrancheAccounting(ctx, cdc, storeKey); err != nil {
			return err
		}
	}

	return nil
}

type TrancheData struct {
	SharesWithdrawn math.Int
	Tranche         types.LimitOrderTranche
}

func fetchTrancheByTrancheUser(
	ctx sdk.Context,
	trancheUser types.LimitOrderTrancheUser,
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey) (*types.LimitOrderTranche, bool, error) {
	key := types.LimitOrderTrancheKey{
		TickIndexTakerToMaker: trancheUser.TickIndexTakerToMaker,
		TradePairId:           trancheUser.TradePairId,
		TrancheKey:            trancheUser.TrancheKey,
	}
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	b := store.Get(key.KeyMarshal())

	if b == nil {
		return nil, false, nil
	}

	var tick types.TickLiquidity
	err := cdc.Unmarshal(b, &tick)
	if err != nil {
		return nil, false, errorsmod.Wrapf(err, "could not unmarshal limitOrderTranche: %s ", &key)
	}

	return tick.GetLimitOrderTranche(), true, nil
}

// Due to a bug in our limitorder accounting we must update the TotalTakerDenom to ensure correct accounting for withdrawals
func migrateLimitOrderTrancheAccounting(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating LimitOrderTranches...")

	// Iterate through all limitOrderTrancheUsers
	store := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.LimitOrderTrancheUserKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	trancheUpdateData := make(map[string]TrancheData)
	trancheUsersToDelete := make([][]byte, 0)

	for ; iterator.Valid(); iterator.Next() {
		var trancheUser types.LimitOrderTrancheUser
		cdc.MustUnmarshal(iterator.Value(), &trancheUser)

		//Sum up the total SharesWithdrawn for each tranche
		if trancheUser.SharesWithdrawn.IsPositive() {
			val, ok := trancheUpdateData[trancheUser.TrancheKey]
			if !ok {
				tranche, found, err := fetchTrancheByTrancheUser(ctx, trancheUser, cdc, storeKey)
				if err != nil {
					return err
				}

				// Due to an earlier error in our rounding behavior / tranche accounting
				// there are trancheUsers that have dust amounts of un-withdrawn shares but the tranche has already been deleted
				// FWIW this issue only exists on pion-1
				// We can safely delete these trancheUsers
				if !found {
					trancheUsersToDelete = append(trancheUsersToDelete, iterator.Key())
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

	}

	err := iterator.Close()
	if err != nil {
		return errorsmod.Wrap(err, "iterator failed to close during migration")
	}

	// Update tranches
	trancheStore := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.TickLiquidityKeyPrefix))
	for _, trancheData := range trancheUpdateData {
		tranche := trancheData.Tranche
		// Calculate total sharesWithdrawn from tranche (denominated in TakerDenom)
		sharesWithdrawnTaker := math_utils.NewPrecDecFromInt(trancheData.SharesWithdrawn).Quo(trancheData.Tranche.PriceTakerToMaker)
		// TotalTakerDenom = ReservesTakerDenom + SharesWithdrawn
		newTotalTakerDenom := sharesWithdrawnTaker.TruncateInt().Add(tranche.ReservesTakerDenom)
		// Update Tranche
		tranche.TotalTakerDenom = newTotalTakerDenom

		// Wrap tranche back into TickLiquidity
		tick := types.TickLiquidity{
			Liquidity: &types.TickLiquidity_LimitOrderTranche{
				LimitOrderTranche: &tranche,
			},
		}

		b, err := cdc.Marshal(&tick)
		if err != nil {
			errorsmod.Wrapf(err, "could not marshal limitOrderTranche: %v ", &tranche)
		}
		trancheStore.Set(tranche.Key.KeyMarshal(), b)

	}

	//Delete the orphaned trancheUsers
	trancheUserStore := prefix.NewStore(ctx.KVStore(storeKey), types.KeyPrefix(types.LimitOrderTrancheUserKeyPrefix))
	for _, key := range trancheUsersToDelete {
		trancheUserStore.Delete(key)
	}

	return nil
}
