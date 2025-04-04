package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/v6/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func createNPools(k *keeper.Keeper, ctx sdk.Context, n int) []*types.Pool {
	items := make([]*types.Pool, n)
	for i := range items {
		pool, err := k.InitPool(ctx, types.MustNewPairID("TokenA", "TokenB"), int64(i), uint64(i)) //nolint:gosec
		if err != nil {
			panic("failed to create pool")
		}
		pool.Deposit(math.NewInt(10), math.NewInt(0), math.ZeroInt(), true)
		k.UpdatePool(ctx, pool)
		items[i] = pool
	}

	return items
}

func TestPoolInit(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)

	pool, err := keeper.InitPool(ctx, defaultPairID, 0, 1)
	require.NoError(t, err)
	pool.Deposit(math.NewInt(1000), math.NewInt(1000), math.NewInt(0), true)
	keeper.UpdatePool(ctx, pool)

	dbPool, found := keeper.GetPool(ctx, defaultPairID, 0, 1)

	require.True(t, found)

	require.Equal(t, pool.Id, dbPool.Id)
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

	pool0, found := keeper.GetPoolByID(ctx, items[0].Id)
	require.True(t, found)
	require.Equal(t, items[0], pool0)

	pool1, found := keeper.GetPoolByID(ctx, items[1].Id)
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
		items[0].LowerTick0.Key.TradePairId.MustPairID(),
		items[0].CenterTickIndexToken1(),
		items[0].Fee(),
	)
	require.True(t, found)
	require.Equal(t, items[0].Id, id0)

	id1, found := keeper.GetPoolIDByParams(
		ctx,
		items[1].LowerTick0.Key.TradePairId.MustPairID(),
		items[1].CenterTickIndexToken1(),
		items[1].Fee(),
	)
	require.True(t, found)
	require.Equal(t, items[1].Id, id1)

	_, found = keeper.GetPoolIDByParams(ctx, defaultPairID, 99, 2)
	require.False(t, found)
}
