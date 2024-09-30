package keeper

import (
	"testing"

	"cosmossdk.io/log"
	metrics2 "cosmossdk.io/store/metrics"
	db2 "github.com/cosmos/cosmos-db"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"

	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	keeper "github.com/neutron-org/neutron/v5/x/transfer/keeper"
	"github.com/neutron-org/neutron/v5/x/transfer/types"
)

func TransferKeeper(
	t testing.TB,
	managerKeeper types.WasmKeeper,
	refunderKeeper types.FeeRefunderKeeper,
	channelKeeper types.ChannelKeeper,
	authKeeper types.AccountKeeper,
) (*keeper.KeeperTransferWrapper, sdk.Context, *storetypes.KVStoreKey) {
	storeKey := storetypes.NewKVStoreKey(transfertypes.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey("mem_" + transfertypes.StoreKey)

	db := db2.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics2.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"TransferParams",
	)
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		paramsSubspace,
		nil, // iscwrapper
		channelKeeper,
		nil,
		authKeeper,
		nil,
		capabilitykeeper.ScopedKeeper{},
		refunderKeeper,
		managerKeeper,
		"authority",
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, transfertypes.DefaultParams())

	return &k, ctx, storeKey
}
