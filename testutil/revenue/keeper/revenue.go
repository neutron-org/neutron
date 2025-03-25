package keeper

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

func RevenueKeeper(
	t testing.TB,
	bankKeeper revenuetypes.BankKeeper,
	oracleKeeper revenuetypes.OracleKeeper,
	authority string,
) (*keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(revenuetypes.StoreKey)
	ss := runtime.NewKVStoreService(storeKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))

	registry := codectypes.NewInterfaceRegistry()
	revenuetypes.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		ss,
		bankKeeper,
		oracleKeeper,
		authority,
	)

	// Initialize params
	err := k.SetParams(testCtx.Ctx, revenuetypes.DefaultParams())
	require.NoError(t, err)

	return k, testCtx.Ctx
}
