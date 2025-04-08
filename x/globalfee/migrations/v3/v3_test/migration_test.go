package v3_test

import (
	"testing"

	"cosmossdk.io/math"
	metrics2 "cosmossdk.io/store/metrics"

	v3 "github.com/neutron-org/neutron/v6/x/globalfee/migrations/v3"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmdb "github.com/cosmos/cosmos-db"

	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	globalfeetypes "github.com/neutron-org/neutron/v6/x/globalfee/types"
)

func TestMigrateStore(t *testing.T) {
	db := cmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics2.NewNoOpMetrics())

	storeKey := storetypes.NewKVStoreKey(paramtypes.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey("mem_key")

	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	require.NoError(t, stateStore.LoadLatestVersion())

	// Create new empty subspace
	newSubspace := paramtypes.NewSubspace(cdc,
		codec.NewLegacyAmino(),
		storeKey,
		memStoreKey,
		paramtypes.ModuleName,
	)

	// register the subspace with the v11 paramKeyTable
	newSubspace = newSubspace.WithKeyTable(globalfeetypes.ParamKeyTable())
	params := globalfeetypes.DefaultParams()
	minGasPrices := sdk.NewDecCoinsFromCoins(sdk.NewCoin("untrn", math.NewInt(1000)))
	params.MinimumGasPrices = minGasPrices
	newSubspace.SetParamSet(ctx, &params)

	store := ctx.KVStore(storeKey)
	bz := store.Get(globalfeetypes.ParamsKey)
	require.Equal(t, 0, len(bz))

	err := v3.MigrateStore(ctx, cdc, newSubspace, storeKey)
	require.NoError(t, err)
	moduleParams := globalfeetypes.Params{}
	bz = store.Get(globalfeetypes.ParamsKey)
	cdc.MustUnmarshal(bz, &moduleParams)
	require.Equal(t, globalfeetypes.DefaultBypassMinFeeMsgTypes, moduleParams.BypassMinFeeMsgTypes)
	require.Equal(t, minGasPrices, moduleParams.MinimumGasPrices)
	require.Equal(t, globalfeetypes.DefaultmaxTotalBypassMinFeeMsgGasUsage, moduleParams.MaxTotalBypassMinFeeMsgGasUsage)
}
