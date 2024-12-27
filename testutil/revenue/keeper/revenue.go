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
	"github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/require"
)

func RevenueKeeper(
	t testing.TB,
	voteAggregator types.VoteAggregator,
	stakingKeeper types.StakingKeeper,
	bankKeeper types.BankKeeper,
) (*keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	ss := runtime.NewKVStoreService(storeKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		ss,
		voteAggregator,
		stakingKeeper,
		bankKeeper,
	)

	// Initialize params
	err := k.SetParams(testCtx.Ctx, types.DefaultParams())
	require.NoError(t, err)

	return k, testCtx.Ctx
}
