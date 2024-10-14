package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	metrics2 "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	db2 "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v5/x/contractmanager"
	"github.com/neutron-org/neutron/v5/x/contractmanager/types"
)

func NewSudoLimitWrapper(t testing.TB, cmKeeper types.ContractManagerKeeper, wasmKeeper types.WasmKeeper) (types.WasmKeeper, sdk.Context, *storetypes.KVStoreKey) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := db2.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics2.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	limitWrapper := contractmanager.NewSudoLimitWrapper(cmKeeper, wasmKeeper)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return limitWrapper, ctx, storeKey
}
