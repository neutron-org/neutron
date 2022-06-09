package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	itypes "github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"testing"
)

// test that GetSubmittedTransactions with start/end works properly
func TestQueryTransactions(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(itypes.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(itypes.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"InterchainqueriesParams",
	)
	keeper := NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		nil, // TODO: do a real ibc keeper
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	queryID := uint64(1)

	lastID := keeper.GetLastSubmittedTransactionIDForQuery(ctx, queryID)

	submittedTransactions := make([]*itypes.Transaction, 0)
	for i := 0; i < 10; i++ {

		tx := itypes.Transaction{
			Height: 0,
			Data:   append([]byte("data"), byte(i)),
		}

		if err := keeper.SaveSubmittedTransaction(ctx, queryID, lastID, tx.Height, tx.Data); err != nil {
			t.Fatalf(err.Error())
		}
		lastID += 1

		submittedTransactions = append(submittedTransactions, &tx)
	}
	keeper.SetLastSubmittedTransactionIDForQuery(ctx, queryID, lastID)

	start, end := 4, 9

	txs, err := keeper.GetSubmittedTransactions(ctx, queryID, uint64(start), uint64(end))
	require.NoError(t, err)

	require.Equal(t, txs, submittedTransactions[start:end])

	// check the same but with multiple query IDS, they should not conflict with each other
	queryID = uint64(2)

	lastID = keeper.GetLastSubmittedTransactionIDForQuery(ctx, queryID)

	submittedTransactions = make([]*itypes.Transaction, 0)
	for i := 0; i < 20; i++ {

		tx := itypes.Transaction{
			Height: 1,
			Data:   append([]byte("another data"), byte(i)),
		}

		if err := keeper.SaveSubmittedTransaction(ctx, queryID, lastID, tx.Height, tx.Data); err != nil {
			t.Fatalf(err.Error())
		}
		lastID += 1

		submittedTransactions = append(submittedTransactions, &tx)
	}
	keeper.SetLastSubmittedTransactionIDForQuery(ctx, queryID, lastID)

	start, end = 3, 8

	txs, err = keeper.GetSubmittedTransactions(ctx, queryID, uint64(start), uint64(end))
	require.NoError(t, err)

	require.Equal(t, txs, submittedTransactions[start:end])
}
