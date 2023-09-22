package keeper_test

import (
	"strconv"
	"testing"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/neutron-org/neutron/testutil/dex/keeper"
	"github.com/neutron-org/neutron/x/dex/keeper"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/require"
)

func CreateNTickLiquidity(keeper *keeper.Keeper, ctx sdk.Context, n int) []*types.TickLiquidity {
	items := make([]*types.TickLiquidity, n)
	for i := range items {
		tick := types.TickLiquidity{
			Liquidity: &types.TickLiquidity_LimitOrderTranche{
				LimitOrderTranche: types.MustNewLimitOrderTranche(
					"TokenA",
					"TokenB",
					strconv.Itoa(i),
					int64(i),
					math.NewInt(10),
					math.NewInt(10),
					math.NewInt(10),
					math.NewInt(10),
				),
			},
		}
		keeper.SetLimitOrderTranche(ctx, tick.GetLimitOrderTranche())
		items[i] = &tick
	}

	return items
}

func TestTickLiquidityGetAll(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := CreateNTickLiquidity(keeper, ctx, 10)
	require.ElementsMatch(t,
		items,
		keeper.GetAllTickLiquidity(ctx),
	)
}
