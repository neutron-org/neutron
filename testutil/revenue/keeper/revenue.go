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

	"github.com/neutron-org/neutron/v5/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
)

func RevenueKeeper(
	t testing.TB,
	voteAggregator revenuetypes.VoteAggregator,
	stakingKeeper revenuetypes.StakingKeeper,
	bankKeeper revenuetypes.BankKeeper,
	oracleKeeper revenuetypes.OracleKeeper,
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
		voteAggregator,
		stakingKeeper,
		bankKeeper,
		oracleKeeper,
	)

	// Initialize params
	err := k.SetParams(testCtx.Ctx, revenuetypes.DefaultParams())
	require.NoError(t, err)

	return k, testCtx.Ctx
}
