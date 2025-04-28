package keeper

import (
	"testing"

	"cosmossdk.io/log"
	metrics2 "cosmossdk.io/store/metrics"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	db2 "github.com/cosmos/cosmos-db"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/x/interchainqueries/keeper"
	"github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

func InterchainQueriesKeeper(
	t testing.TB,
	ibcKeeper *ibckeeper.Keeper,
	contractManager types.ContractManagerKeeper,
	headerVerifier types.HeaderVerifier,
	txVerifier types.TransactionVerifier,
) (*keeper.Keeper, sdk.Context) {
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
		memStoreKey,
		ibcKeeper, // TODO: do a real ibc keeper
		nil,       // TODO: do a real wasm keeper
		contractManager,
		headerVerifier,
		txVerifier,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	err := k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return k, ctx
}
