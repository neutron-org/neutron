package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/neutron-org/neutron/testutil/dex/keeper"
	"github.com/neutron-org/neutron/x/dex/keeper"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/require"
)

func createNPools(k *keeper.Keeper, ctx sdk.Context, n int) []*types.Pool {
	items := make([]*types.Pool, n)
	for i := range items {
		pool, err := k.InitPool(ctx, types.MustNewPairID("TokenA", "TokenB"), int64(i), uint64(i))
		if err != nil {
			panic("failed to create pool")
		}
		pool.Deposit(math.NewInt(10), math.NewInt(0), math.ZeroInt(), true)
		k.SetPool(ctx, pool)
		items[i] = pool
	}

	return items
}

func TestPoolInit(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)

	pool, err := keeper.InitPool(ctx, defaultPairID, 0, 1)
	require.NoError(t, err)
	pool.Deposit(math.NewInt(100), math.NewInt(100), math.NewInt(0), true)
	keeper.SetPool(ctx, pool)

	dbPool, found := keeper.GetPool(ctx, defaultPairID, 0, 1)

	require.True(t, found)

	require.Equal(t, pool.ID, dbPool.ID)
	require.Equal(t, pool.LowerTick0, dbPool.LowerTick0)
	require.Equal(t, pool.UpperTick1, dbPool.UpperTick1)
}

func TestPoolCount(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNPools(keeper, ctx, 10)
	count := uint64(len(items))
	require.Equal(t, count, keeper.GetPoolCount(ctx))
}

func TestGetPoolByID(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNPools(keeper, ctx, 2)

	pool0, found := keeper.GetPoolByID(ctx, items[0].ID)
	require.True(t, found)
	require.Equal(t, items[0], pool0)

	pool1, found := keeper.GetPoolByID(ctx, items[1].ID)
	require.True(t, found)
	require.Equal(t, items[1], pool1)

	_, found = keeper.GetPoolByID(ctx, 99)
	require.False(t, found)
}

func TestGetPoolIDByParams(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNPools(keeper, ctx, 2)

	id0, found := keeper.GetPoolIDByParams(
		ctx,
		items[0].LowerTick0.Key.TradePairID.MustPairID(),
		items[0].CenterTickIndex(),
		items[0].Fee(),
	)
	require.True(t, found)
	require.Equal(t, items[0].ID, id0)

	id1, found := keeper.GetPoolIDByParams(
		ctx,
		items[1].LowerTick0.Key.TradePairID.MustPairID(),
		items[1].CenterTickIndex(),
		items[1].Fee(),
	)
	require.True(t, found)
	require.Equal(t, items[1].ID, id1)

	_, found = keeper.GetPoolIDByParams(ctx, defaultPairID, 99, 2)
	require.False(t, found)
}
