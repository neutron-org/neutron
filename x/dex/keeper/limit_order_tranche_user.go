package keeper

import (
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dex/types"
)

func (k Keeper) GetOrInitLimitOrderTrancheUser(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
	tickIndex int64,
	trancheKey string,
	orderType types.LimitOrderType,
	receiver string,
) *types.LimitOrderTrancheUser {
	userShareData, found := k.GetLimitOrderTrancheUser(ctx, receiver, trancheKey)

	if !found {
		return &types.LimitOrderTrancheUser{
			TrancheKey:            trancheKey,
			Address:               receiver,
			SharesOwned:           math.ZeroInt(),
			SharesWithdrawn:       math.ZeroInt(),
			TickIndexTakerToMaker: tickIndex,
			TradePairId:           tradePairID,
			OrderType:             orderType,
		}
	}

	return userShareData
}

// SetLimitOrderTrancheUser set a specific LimitOrderTrancheUser in the store from its index
func (k Keeper) SetLimitOrderTrancheUser(ctx sdk.Context, limitOrderTrancheUser *types.LimitOrderTrancheUser) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LimitOrderTrancheUserKeyPrefix))
	b := k.cdc.MustMarshal(limitOrderTrancheUser)
	store.Set(types.LimitOrderTrancheUserKey(
		limitOrderTrancheUser.Address,
		limitOrderTrancheUser.TrancheKey,
	), b)
}

// GetLimitOrderTrancheUser returns a LimitOrderTrancheUser from its index
func (k Keeper) GetLimitOrderTrancheUser(
	ctx sdk.Context,
	address string,
	trancheKey string,
) (val *types.LimitOrderTrancheUser, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LimitOrderTrancheUserKeyPrefix))

	b := store.Get(types.LimitOrderTrancheUserKey(
		address,
		trancheKey,
	))
	if b == nil {
		return nil, false
	}

	val = &types.LimitOrderTrancheUser{}
	k.cdc.MustUnmarshal(b, val)

	return val, true
}

// RemoveLimitOrderTrancheUserByKey removes a LimitOrderTrancheUser from the store
func (k Keeper) RemoveLimitOrderTrancheUserByKey(
	ctx sdk.Context,
	trancheKey string,
	address string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LimitOrderTrancheUserKeyPrefix))
	store.Delete(types.LimitOrderTrancheUserKey(
		address,
		trancheKey,
	))
}

func (k Keeper) RemoveLimitOrderTrancheUser(ctx sdk.Context, trancheUser *types.LimitOrderTrancheUser) {
	k.RemoveLimitOrderTrancheUserByKey(
		ctx,
		trancheUser.TrancheKey,
		trancheUser.Address,
	)
}

func (k Keeper) SaveTrancheUser(ctx sdk.Context, trancheUser *types.LimitOrderTrancheUser) {
	if trancheUser.IsEmpty() {
		k.RemoveLimitOrderTrancheUser(ctx, trancheUser)
	} else {
		k.SetLimitOrderTrancheUser(ctx, trancheUser)
	}
}

// GetAllLimitOrderTrancheUser returns all LimitOrderTrancheUser
func (k Keeper) GetAllLimitOrderTrancheUser(ctx sdk.Context) (list []*types.LimitOrderTrancheUser) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LimitOrderTrancheUserKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := &types.LimitOrderTrancheUser{}
		k.cdc.MustUnmarshal(iterator.Value(), val)
		list = append(list, val)
	}

	return
}

func (k Keeper) GetAllLimitOrderTrancheUserForAddress(
	ctx sdk.Context,
	address sdk.AccAddress,
) (list []*types.LimitOrderTrancheUser) {
	addressPrefix := types.LimitOrderTrancheUserAddressPrefix(address.String())
	store := prefix.NewStore(ctx.KVStore(k.storeKey), addressPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := &types.LimitOrderTrancheUser{}
		k.cdc.MustUnmarshal(iterator.Value(), val)
		list = append(list, val)
	}

	return
}
