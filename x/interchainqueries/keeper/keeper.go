package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
		ibcKeeper  *ibckeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	ibcKeeper *ibckeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{

		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		ibcKeeper:  ibcKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetLastRegisteredQueryKey(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastRegisteredQueryIdKey)
	if bytes == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}

func (k Keeper) SetLastRegisteredQueryKey(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastRegisteredQueryIdKey, sdk.Uint64ToBigEndian(id))
}

func (k Keeper) SaveQuery(ctx sdk.Context, query types.RegisteredQuery) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&query)
	store.Set(types.GetRegisteredQueryByIDKey(query.Id), bz)
}

func (k Keeper) GetQueryByID(ctx sdk.Context, id uint64) (*types.RegisteredQuery, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetRegisteredQueryByIDKey(id))
	if bz == nil {
		return nil, fmt.Errorf("there is no query with id: %v", id)
	}

	var query types.RegisteredQuery
	if err := k.cdc.Unmarshal(bz, &query); err != nil {
		return nil, fmt.Errorf("failed to unmarshal registered query: %w", err)
	}

	return &query, nil
}

func (k Keeper) SaveQueryResult(ctx sdk.Context, id uint64, query *types.QueryResult) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(query)
	store.Set(types.GetRegisteredQueryResultByIDKey(id), bz)
}

func (k Keeper) GetQueryResultByID(ctx sdk.Context, id uint64) (*types.QueryResult, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetRegisteredQueryResultByIDKey(id))
	if bz == nil {
		return nil, fmt.Errorf("there is no query result with id: %v", id)
	}

	var query types.QueryResult
	if err := k.cdc.Unmarshal(bz, &query); err != nil {
		return nil, fmt.Errorf("failed to unmarshal registered query: %w", err)
	}

	return &query, nil
}

func (k Keeper) IterateRegisteredQueries(ctx sdk.Context, fn func(index int64, queryInfo types.RegisteredQuery) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RegisteredQueryKey)
	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		query := types.RegisteredQuery{}
		k.cdc.MustUnmarshal(iterator.Value(), &query)
		stop := fn(i, query)

		if stop {
			break
		}
		i++
	}
}
