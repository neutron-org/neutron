package keeper

import (
	"testing"

	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	keeper "github.com/neutron-org/neutron/x/transfer/keeper"
	"github.com/neutron-org/neutron/x/transfer/types"
)

func TransferKeeper(t testing.TB, managerKeeper types.ContractManagerKeeper, refunderKeeper types.FeeRefunderKeeper, channelKeeper types.ChannelKeeper, authKeeper types.AccountKeeper) (*keeper.KeeperTransferWrapper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(transfertypes.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey("mem_" + transfertypes.StoreKey)

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
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, transfertypes.DefaultParams())

	return &k, ctx
}
