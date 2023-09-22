package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/neutron-org/neutron/x/incentives/types"
)

func NewAccountHistory(account string, coins sdk.Coins) *types.AccountHistory {
	return &types.AccountHistory{
		Account: account,
		Coins:   coins,
	}
}

// SetAccountHistory set a specific goodTilRecord in the store from its index
func (k Keeper) SetAccountHistory(
	ctx sdk.Context,
	accountHistory *types.AccountHistory,
) error {
	store := ctx.KVStore(k.storeKey)
	b, err := proto.Marshal(accountHistory)
	if err != nil {
		return err
	}
	store.Set(types.GetKeyAccountHistory(
		accountHistory.Account,
	), b)
	return nil
}

// GetAccountHistory returns a goodTilRecord from its index
func (k Keeper) GetAccountHistory(
	ctx sdk.Context,
	account string,
) (val *types.AccountHistory, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.GetKeyAccountHistory(account))
	if b == nil {
		return val, false
	}

	val = &types.AccountHistory{}
	err := proto.Unmarshal(b, val)
	if err != nil {
		panic(err)
	}

	return val, true
}

// RemoveAccountHistory removes a goodTilRecord from the store
func (k Keeper) RemoveAccountHistory(
	ctx sdk.Context,
	account string,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetKeyAccountHistory(account))
}

// GetAllAccountHistory returns all goodTilRecord
func (k Keeper) GetAllAccountHistory(ctx sdk.Context) (list []*types.AccountHistory) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefixAccountHistory,
	)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := &types.AccountHistory{}
		err := proto.Unmarshal(iterator.Value(), val)
		if err != nil {
			panic(err)
		}
		list = append(list, val)
	}

	return
}
