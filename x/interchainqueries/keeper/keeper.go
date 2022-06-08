package keeper

import (
	"fmt"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	tendermintLightClientTypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"

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

	cleanResult := clearQueryResult(query)
	bz := k.cdc.MustMarshal(&cleanResult)

	store.Set(types.GetRegisteredQueryResultByIDKey(id), bz)
}

// We don't need to store proofs or transactions, so we just remove unnecessary fields
func clearQueryResult(result *types.QueryResult) types.QueryResult {
	storageValues := make([]*types.StorageValue, 0, len(result.KvResults))
	for _, v := range result.KvResults {
		storageValues = append(storageValues, &types.StorageValue{
			StoragePrefix: v.StoragePrefix,
			Key:           v.Key,
			Value:         v.Value,
			Proof:         nil,
		})
	}

	cleanResult := types.QueryResult{
		KvResults: storageValues,
		Blocks:    nil,
		Height:    result.Height,
	}

	return cleanResult
}

// SaveTransactions save transactions to the storage and updates LastSubmittedTransactionID
func (k Keeper) SaveTransactions(ctx sdk.Context, queryID uint64, blocks []types.Block) error {
	lastSubmittedTxID := k.GetLastSubmittedTransactionIDForQuery(ctx, queryID)

	for _, block := range blocks {
		header, err := ibcclienttypes.UnpackHeader(block.Header)
		if err != nil {
			return fmt.Errorf("failed to unpack block header: %w", err)
		}

		tmHeader, ok := header.(*tendermintLightClientTypes.Header)
		if !ok {
			return fmt.Errorf("failed to cast header to tendermint Header: %w", err)
		}

		for _, tx := range block.Txs {
			lastSubmittedTxID += 1
			if err = k.SaveSubmittedTransaction(ctx, queryID, lastSubmittedTxID, uint64(tmHeader.Header.Height), tx.Data); err != nil {
				return fmt.Errorf("failed save submitted transaction: %w", err)
			}
		}
	}

	k.SetLastSubmittedTransactionIDForQuery(ctx, queryID, lastSubmittedTxID)

	return nil
}

// SaveSubmittedTransaction saves a transaction data into the storage with a key (SubmittedTxKey + bigEndianBytes(queryID) + bigEndianBytes(txID))
func (k Keeper) SaveSubmittedTransaction(ctx sdk.Context, queryID uint64, txID uint64, height uint64, txData []byte) error {
	txBz, err := (&types.Transaction{
		Height: height,
		Data:   txData,
	}).Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetSubmittedTransactionIDForQueryKey(queryID, txID)

	store.Set(key, txBz)

	return nil
}

// GetLastSubmittedTransactionIDForQuery returns last transaction id which was submitted for a query with queryID
func (k Keeper) GetLastSubmittedTransactionIDForQuery(ctx sdk.Context, queryID uint64) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.GetLastSubmittedTransactionIDForQueryKey(queryID))
	if bytes == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}

// SetLastSubmittedTransactionIDForQuery sets a last transaction id which was submitted for a query with queryID
func (k Keeper) SetLastSubmittedTransactionIDForQuery(ctx sdk.Context, queryID uint64, transactionID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetLastSubmittedTransactionIDForQueryKey(queryID), sdk.Uint64ToBigEndian(transactionID))
}

// GetQueryResultByID returns a QueryResult for query with id
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

// GetSubmittedTransactions returns a list of transactions from start ID to end ID
func (k Keeper) GetSubmittedTransactions(ctx sdk.Context, queryID uint64, start uint64, end uint64) ([]*types.Transaction, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.GetSubmittedTransactionIDForQueryKey(queryID, start), types.GetSubmittedTransactionIDForQueryKey(queryID, end))
	defer iterator.Close()

	transactions := make([]*types.Transaction, 0)
	for ; iterator.Valid(); iterator.Next() {
		var tx types.Transaction
		if err := tx.Unmarshal(iterator.Value()); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
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
