package keeper

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"

	"github.com/neutron-org/neutron/x/incentives/types"
)

// GetLastStakeID returns ID used last time.
func (k Keeper) GetLastStakeID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastStakeID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastStakeID save ID used by last stake.
func (k Keeper) SetLastStakeID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastStakeID, sdk.Uint64ToBigEndian(id))
}

func (k Keeper) Unstake(ctx sdk.Context, stake *types.Stake, coins sdk.Coins) (uint64, error) {
	if coins.Empty() {
		coins = stake.Coins
	}

	if !coins.IsAllLTE(stake.Coins) {
		return 0, fmt.Errorf("requested amount to unstake exceeds locked tokens")
	}

	// remove existing stake refs from not unlocking queue
	err := k.deleteStakeRefs(ctx, stake)
	if err != nil {
		return 0, err
	}

	if len(coins) != 0 && !coins.IsEqual(stake.Coins) {
		stake.Coins = stake.Coins.Sub(coins...)
		err := k.setStake(ctx, stake)
		if err != nil {
			return 0, err
		}

		// re-add remaining stake refs
		err = k.addStakeRefs(ctx, stake)
		if err != nil {
			return 0, err
		}
	} else {
		k.deleteStake(ctx, stake.ID)
	}

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, stake.OwnerAddress(), coins)
	if err != nil {
		return 0, err
	}

	if k.hooks != nil {
		k.hooks.OnTokenUnstaked(ctx, stake.OwnerAddress(), stake.ID, stake.Coins, ctx.BlockTime())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtUnstake,
			sdk.NewAttribute(types.AttributeStakeID, strconv.FormatUint(stake.ID, 10)),
			sdk.NewAttribute(types.AttributeStakeOwner, stake.Owner),
			sdk.NewAttribute(types.AttributeStakeStakeTime, stake.StartTime.String()),
			sdk.NewAttribute(types.AttributeUnstakedCoins, coins.String()),
		),
	})

	return stake.ID, err
}

// setStake is a utility to store stake object into the store.
func (k Keeper) setStake(ctx sdk.Context, stake *types.Stake) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(stake)
	if err != nil {
		return err
	}
	store.Set(types.GetStakeStoreKey(stake.ID), bz)
	return nil
}

// deleteStake removes the stake object from the state.
func (k Keeper) deleteStake(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetStakeStoreKey(id))
}

// GetStakeByID Returns stake from stakeID.
func (k Keeper) GetStakeByID(ctx sdk.Context, stakeID uint64) (*types.Stake, error) {
	stake := types.Stake{}
	store := ctx.KVStore(k.storeKey)
	lockKey := types.GetStakeStoreKey(stakeID)
	if !store.Has(lockKey) {
		return nil, sdkerrors.Wrap(
			types.ErrStakeNotFound,
			fmt.Sprintf("stake with ID %d does not exist", stakeID),
		)
	}
	bz := store.Get(lockKey)
	err := proto.Unmarshal(bz, &stake)
	return &stake, err
}

// GetAccountStakes Returns the period locks associated to an account.
func (k Keeper) GetStakesByQueryCondition(
	ctx sdk.Context,
	distrTo *types.QueryCondition,
) types.Stakes {
	pairIDString := distrTo.PairID.CanonicalString()
	tickStakeIds := k.getIDsFromIterator(
		k.iteratorStartEnd(
			ctx,
			types.GetKeyStakeIndexByPairTick(pairIDString, distrTo.StartTick),
			types.GetKeyStakeIndexByPairTick(pairIDString, distrTo.EndTick+1),
		),
	)

	idMemo := make(map[uint64]bool)
	for _, id := range tickStakeIds {
		idMemo[id] = true
	}

	params := k.GetParams(ctx)
	curEpoch := k.ek.GetEpochInfo(ctx, params.GetDistrEpochIdentifier())
	timeStakeIds := k.getIDsFromIterator(
		k.iteratorStartEnd(
			ctx,
			types.CombineKeys(
				types.KeyPrefixStakeIndexPairDistEpoch,
				[]byte(pairIDString),
			),
			sdk.PrefixEndBytes(types.GetKeyStakeIndexByDistEpoch(
				pairIDString,
				curEpoch.CurrentEpoch-2,
			)),
		),
	)

	resultIds := []uint64{}
	for _, id := range timeStakeIds {
		if _, ok := idMemo[id]; ok {
			resultIds = append(resultIds, id)
		}
	}

	results := make([]*types.Stake, len(resultIds))
	for i, stakeID := range resultIds {
		stake, err := k.GetStakeByID(ctx, stakeID)
		if err != nil {
			// This represents a db inconsistency
			panic(err)
		}
		results[i] = stake
	}
	return results
}

func (k Keeper) GetStakes(ctx sdk.Context) types.Stakes {
	return k.getStakesFromIterator(ctx, k.iterator(ctx, types.KeyPrefixStakeIndex))
}

func (k Keeper) getStakesByAccount(ctx sdk.Context, acct sdk.AccAddress) types.Stakes {
	return k.getStakesFromIterator(ctx, k.iterator(ctx, types.GetKeyStakeIndexByAccount(acct)))
}

// GetAccountStakes Returns the period locks associated to an account.
func (k Keeper) GetStakesByAccount(ctx sdk.Context, addr sdk.AccAddress) types.Stakes {
	return k.getStakesFromIterator(ctx, k.iterator(ctx, types.GetKeyStakeIndexByAccount(addr)))
}

func (k Keeper) CreateStake(
	ctx sdk.Context,
	owner sdk.AccAddress,
	coins sdk.Coins,
	startTime time.Time,
	startDistEpoch int64,
) (*types.Stake, error) {
	ID := k.GetLastStakeID(ctx) + 1

	// unlock time is initially set without a value, gets set as unlock start time + duration
	// when unlocking starts.
	stake := types.NewStake(ID, owner, coins, startTime, startDistEpoch)

	owner, err := sdk.AccAddressFromBech32(stake.Owner)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	err = stake.ValidateBasic()
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, stake.Coins); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// store stake object into the store
	err = k.setStake(ctx, stake)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	k.hooks.OnTokenStaked(ctx, owner, stake.ID, stake.Coins, ctx.BlockTime())
	k.SetLastStakeID(ctx, stake.ID)

	// add stake refs into not unlocking queue
	err = k.addStakeRefs(ctx, stake)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return stake, nil
}

func (k Keeper) StakeCoinsPassingQueryCondition(
	ctx sdk.Context,
	stake *types.Stake,
	distrTo types.QueryCondition,
) sdk.Coins {
	coins := stake.Coins
	result := sdk.NewCoins()
	for _, c := range coins {
		poolMetadata, err := k.dk.GetPoolMetadataByDenom(ctx, c.Denom)
		if err != nil {
			continue
		}

		if distrTo.Test(poolMetadata) {
			result = result.Add(c)
		}
	}
	return result
}
