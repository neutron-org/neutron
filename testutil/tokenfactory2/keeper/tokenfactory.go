package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	metrics2 "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	db2 "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v8/testutil"
	"github.com/neutron-org/neutron/v8/x/tokenfactory2/keeper"
	"github.com/neutron-org/neutron/v8/x/tokenfactory2/types"
)

func TokenFactoryKeeper(
	t testing.TB,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	contractKeeper types.ContractKeeper,
) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := db2.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics2.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		nil,
		accountKeeper,
		bankKeeper,
		contractKeeper,
		testutil.TestOwnerAddress,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	err := k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return k, ctx
}
