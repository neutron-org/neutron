package keeper_test

import (
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v6/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func createNInactiveLimitOrderTranche(
	keeper *keeper.Keeper,
	ctx sdk.Context,
	n int,
) []*types.LimitOrderTranche {
	items := make([]*types.LimitOrderTranche, n)
	for i := range items {
		items[i] = types.MustNewLimitOrderTranche(
			"TokenA",
			"TokenB",
			strconv.Itoa(i),
			int64(i),
			math.ZeroInt(),
			math.ZeroInt(),
			math.ZeroInt(),
			math.ZeroInt(),
		)
		keeper.SetInactiveLimitOrderTranche(ctx, items[i])
	}

	return items
}

func TestInactiveLimitOrderTrancheGet(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNInactiveLimitOrderTranche(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetInactiveLimitOrderTranche(ctx, item.Key)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(item),
			nullify.Fill(rst),
		)
	}
}

func TestInactiveLimitOrderTrancheRemove(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNInactiveLimitOrderTranche(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveInactiveLimitOrderTranche(ctx, item.Key)
		_, found := keeper.GetInactiveLimitOrderTranche(ctx, item.Key)
		require.False(t, found)
	}
}

func TestInactiveLimitOrderTrancheGetAll(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNInactiveLimitOrderTranche(keeper, ctx, 10)

	require.ElementsMatch(t,
		items,
		keeper.GetAllInactiveLimitOrderTranche(ctx),
	)
}
