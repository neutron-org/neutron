package keeper

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

func (k Keeper) SaveQuery(ctx sdk.Context, query types.RegisteredQuery) error {
	store := ctx.KVStore(k.storeKey)

	bz, err := k.cdc.Marshal(&query)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrProtoMarshal, "failed to marshal registered query: %v", err)
	}

	store.Set(types.GetRegisteredQueryByIDKey(query.Id), bz)

	return nil
}

func (k Keeper) GetQueryByID(ctx sdk.Context, id uint64) (*types.RegisteredQuery, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetRegisteredQueryByIDKey(id))
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidQueryID, "there is no query with id: %v", id)
	}

	var query types.RegisteredQuery
	if err := k.cdc.Unmarshal(bz, &query); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unmarshal registered query: %v", err)
	}

	return &query, nil
}

func (k Keeper) SaveQueryResult(ctx sdk.Context, id uint64, query *types.QueryResult) error {
	store := ctx.KVStore(k.storeKey)

	if query.Blocks != nil {
		if err := k.SaveTransactions(ctx, id, query.Blocks); err != nil {
			return sdkerrors.Wrapf(types.ErrInternal, "failed to save transactions: %v", err)
		}
	}

	if query.KvResults != nil {
		cleanResult := clearQueryResult(query)
		bz, err := k.cdc.Marshal(&cleanResult)
		if err != nil {
			return sdkerrors.Wrapf(types.ErrProtoMarshal, "failed to marshal query result: %v", err)
		}

		store.Set(types.GetRegisteredQueryResultByIDKey(id), bz)

		if err = k.UpdateLastRemoteHeight(ctx, id, query.Height); err != nil {
			return sdkerrors.Wrapf(types.ErrInternal, "failed to update last remote height for a query with id %d: %v", id, err)
		}

		if err = k.UpdateLastLocalHeight(ctx, id, uint64(ctx.BlockHeight())); err != nil {
			return sdkerrors.Wrapf(types.ErrInternal, "failed to update last local height for a query with id %d: %v", id, err)
		}
	}

	return nil
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
func (k Keeper) SaveTransactions(ctx sdk.Context, queryID uint64, blocks []*types.Block) error {
	lastSubmittedTxID := k.GetLastSubmittedTransactionIDForQuery(ctx, queryID)

	maxHeight := int64(0)
	for _, block := range blocks {
		header, err := ibcclienttypes.UnpackHeader(block.Header)
		if err != nil {
			return sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack block header: %v", err)
		}

		tmHeader, ok := header.(*tendermintLightClientTypes.Header)
		if !ok {
			return sdkerrors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header: %v", err)
		}

		for _, tx := range block.Txs {
			lastSubmittedTxID += 1
			if err = k.SaveSubmittedTransaction(ctx, queryID, lastSubmittedTxID, uint64(tmHeader.Header.Height), tx.Data); err != nil {
				return sdkerrors.Wrapf(types.ErrInternal, "failed save submitted transaction: %v", err)
			}
		}

		if tmHeader.Header.Height > maxHeight {
			maxHeight = tmHeader.Header.Height
		}
	}

	if err := k.UpdateLastRemoteHeight(ctx, queryID, uint64(maxHeight)); err != nil {
		return sdkerrors.Wrapf(types.ErrInternal, "failed to update last remote height for a query with id %d: %v", queryID, err)
	}

	if err := k.UpdateLastLocalHeight(ctx, queryID, uint64(ctx.BlockHeight())); err != nil {
		return sdkerrors.Wrapf(types.ErrInternal, "failed to update last local height for a query with id %d: %v", queryID, err)
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
		return sdkerrors.Wrapf(types.ErrProtoMarshal, "failed to marshal transaction: %v", err)
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
		return nil, sdkerrors.Wrapf(types.ErrInvalidQueryID, "there is no query result with id: %v", id)
	}

	var query types.QueryResult
	if err := k.cdc.Unmarshal(bz, &query); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unmarshal registered query: %v", err)
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
			return nil, sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unmarshal transaction: %v", err)
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

func (k Keeper) UpdateLastLocalHeight(ctx sdk.Context, queryID uint64, newLocalHeight uint64) error {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetRegisteredQueryByIDKey(queryID))
	if bz == nil {
		return sdkerrors.Wrapf(types.ErrInvalidQueryID, "query with ID %d not found", queryID)
	}

	var query types.RegisteredQuery
	if err := k.cdc.Unmarshal(bz, &query); err != nil {
		return sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unmarshal registered query: %v", err)
	}

	query.LastSubmittedResultLocalHeight = newLocalHeight

	return k.SaveQuery(ctx, query)
}

func (k Keeper) UpdateLastRemoteHeight(ctx sdk.Context, queryID uint64, newRemoteHeight uint64) error {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetRegisteredQueryByIDKey(queryID))
	if bz == nil {
		return sdkerrors.Wrapf(types.ErrInvalidQueryID, "query with ID %d not found", queryID)
	}

	var query types.RegisteredQuery
	if err := k.cdc.Unmarshal(bz, &query); err != nil {
		return sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unmarshal registered query: %v", err)
	}

	if query.LastSubmittedResultRemoteHeight >= newRemoteHeight {
		return nil
	}

	query.LastSubmittedResultRemoteHeight = newRemoteHeight
	return k.SaveQuery(ctx, query)
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
