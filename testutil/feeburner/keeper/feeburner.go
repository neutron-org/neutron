package keeper

import (
	"testing"

	"cosmossdk.io/log"
	metrics2 "cosmossdk.io/store/metrics"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	db2 "github.com/cosmos/cosmos-db"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/x/feeburner/keeper"
	"github.com/neutron-org/neutron/v4/x/feeburner/types"
)

func FeeburnerKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	return FeeburnerKeeperWithDeps(t, nil, nil)
}

func FeeburnerKeeperWithDeps(t testing.TB, accountKeeper types.AccountKeeper, bankkeeper types.BankKeeper) (*keeper.Keeper, sdk.Context) {
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
		accountKeeper,
		bankkeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	err := k.SetParams(ctx, types.NewParams(types.DefaultNeutronDenom, "neutron1vguuxez2h5ekltfj9gjd62fs5k4rl2zy5hfrncasykzw08rezpfsd2rhm7"))
	require.NoError(t, err)

	return k, ctx
}
