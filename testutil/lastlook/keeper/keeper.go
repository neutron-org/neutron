package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	cosmosdb "github.com/cosmos/cosmos-db"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/x/lastlook/keeper"
	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

func LastLookKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := cosmosdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	err := k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return k, ctx
}
