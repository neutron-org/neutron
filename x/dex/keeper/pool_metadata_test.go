package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/v2/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v2/testutil/dex/nullify"
	"github.com/neutron-org/neutron/v2/x/dex/keeper"
	"github.com/neutron-org/neutron/v2/x/dex/types"
)

func createNPoolMetadata(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.PoolMetadata {
	items := make([]types.PoolMetadata, n)
	for i := range items {
		items[i].Id = uint64(i)
		keeper.SetPoolMetadata(ctx, items[i])
	}

	return items
}

func TestPoolMetadataGet(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNPoolMetadata(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetPoolMetadata(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(item),
			nullify.Fill(got),
		)
	}
}

func TestPoolMetadataRemove(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNPoolMetadata(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemovePoolMetadata(ctx, item.Id)
		_, found := keeper.GetPoolMetadata(ctx, item.Id)
		require.False(t, found)
	}
}

func TestPoolMetadataGetAll(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNPoolMetadata(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllPoolMetadata(ctx)),
	)
}
