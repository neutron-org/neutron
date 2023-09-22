package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/types"
)

// addStakeRefs adds appropriate reference keys preceded by a prefix.
// A prefix indicates whether the stake is unstaking or not.
func (k Keeper) addStakeRefs(ctx sdk.Context, stake *types.Stake) error {
	refKeys, err := k.getStakeRefKeys(ctx, stake)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		if err := k.addRefByKey(ctx, refKey, stake.ID); err != nil {
			return err
		}
	}
	return nil
}

// deleteStakeRefs deletes all the stake references of the stake with the given stake prefix.
func (k Keeper) deleteStakeRefs(ctx sdk.Context, stake *types.Stake) error {
	refKeys, err := k.getStakeRefKeys(ctx, stake)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		err = k.deleteRefByKey(ctx, refKey, stake.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) getStakeRefKeys(ctx sdk.Context, stake *types.Stake) ([][]byte, error) {
	owner, err := sdk.AccAddressFromBech32(stake.Owner)
	if err != nil {
		return nil, err
	}

	refKeys := make(map[string]bool)
	refKeys[string(types.KeyPrefixStakeIndex)] = true
	refKeys[string(types.CombineKeys(types.KeyPrefixStakeIndexAccount, owner))] = true

	for _, coin := range stake.Coins {
		poolMetadata, err := k.dk.GetPoolMetadataByDenom(ctx, coin.Denom)
		if err != nil {
			panic("Only valid LP tokens should be staked")
		}
		denomBz := []byte(coin.Denom)
		pairIDBz := []byte(poolMetadata.PairID.CanonicalString())
		tickBz := dextypes.TickIndexToBytes(poolMetadata.Tick)
		refKeys[string(types.CombineKeys(types.KeyPrefixStakeIndexDenom, denomBz))] = true
		refKeys[string(types.CombineKeys(types.KeyPrefixStakeIndexPairTick, pairIDBz, tickBz))] = true
		refKeys[string(types.CombineKeys(types.KeyPrefixStakeIndexAccountDenom, owner, denomBz))] = true
		refKeys[string(types.CombineKeys(
			types.KeyPrefixStakeIndexPairDistEpoch,
			pairIDBz,
			types.GetKeyInt64(stake.StartDistEpoch),
		))] = true
	}

	refKeyBytes := make([][]byte, 0, len(refKeys))
	for k := range refKeys {
		refKeyBytes = append(refKeyBytes, []byte(k))
	}
	return refKeyBytes, nil
}
