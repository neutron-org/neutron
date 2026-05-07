package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v10/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v10/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v10/x/dex/keeper"
	"github.com/neutron-org/neutron/v10/x/dex/types"
)

func createNLimitOrderTranches(
	keeper *keeper.Keeper,
	ctx sdk.Context,
	n int,
) []*types.LimitOrderTranche {
	items := make([]*types.LimitOrderTranche, n)
	for i := range items {
		items[i] = types.MustNewLimitOrderTranche(
			"TokenA",
			"TokenB",
			keeper.NextTrancheKey(ctx),
			int64(i),
			math.ZeroInt(),
			math.ZeroInt(),
			math.ZeroInt(),
			math.ZeroInt(),
		)
		keeper.SetLimitOrderTranche(ctx, items[i])
	}

	return items
}

func TestGetLimitOrderTranche(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNLimitOrderTranches(keeper, ctx, 10)
	for n, item := range items {
		rst := keeper.GetLimitOrderTranche(ctx, item.Key)
		require.NotNil(t, rst)
		require.Equal(t,
			nullify.Fill(item),
			nullify.Fill(rst),
		)
		require.Equal(t, types.NewTrancheKey(uint64(n)), item.Key.TrancheKey)
	}
}

func TestRemoveLimitOrderTranche(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNLimitOrderTranches(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLimitOrderTranche(ctx, item.Key)
		rst := keeper.GetLimitOrderTranche(ctx, item.Key)
		require.Nil(t, rst)
	}
}
