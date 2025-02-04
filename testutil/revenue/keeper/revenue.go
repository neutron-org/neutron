package keeper

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v5/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/require"
)

func RevenueKeeper(
	t testing.TB,
	bankKeeper revenuetypes.BankKeeper,
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
		authority,
	)

	// Initialize params
	err := k.SetParams(testCtx.Ctx, revenuetypes.DefaultParams())
	require.NoError(t, err)

	return k, testCtx.Ctx
}
