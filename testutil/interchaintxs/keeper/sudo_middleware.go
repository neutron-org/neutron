package keeper

import (
	"testing"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/x/contractmanager"
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

func NewSudoLimitWrapper(t testing.TB, cmKeeper types.ContractManagerKeeper, wasmKeeper types.WasmKeeper) (types.WasmKeeper, sdk.Context, *storetypes.KVStoreKey) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	limitWrapper := contractmanager.NewSudoLimitWrapper(cmKeeper, wasmKeeper)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return limitWrapper, ctx, storeKey
}
