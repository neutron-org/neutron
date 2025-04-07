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

func createNLimitOrderTrancheUser(keeper *keeper.Keeper, ctx sdk.Context, n int) []*types.LimitOrderTrancheUser {
	items := make([]*types.LimitOrderTrancheUser, n)
	for i := range items {
		val := &types.LimitOrderTrancheUser{
			TrancheKey:            strconv.Itoa(i),
			Address:               strconv.Itoa(i),
			TradePairId:           &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
			TickIndexTakerToMaker: int64(i),
			SharesOwned:           math.NewInt(100),
			SharesWithdrawn:       math.ZeroInt(),
		}
		items[i] = val
		keeper.SetLimitOrderTrancheUser(ctx, items[i])
	}

	return items
}

func createNLimitOrderTrancheUserWithAddress(keeper *keeper.Keeper, ctx sdk.Context, address string, n int) []*types.LimitOrderTrancheUser {
	items := make([]*types.LimitOrderTrancheUser, n)
	for i := range items {
		val := &types.LimitOrderTrancheUser{
			TrancheKey:            strconv.Itoa(i),
			Address:               address,
			TradePairId:           &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
			TickIndexTakerToMaker: 0,
			SharesOwned:           math.ZeroInt(),
			SharesWithdrawn:       math.ZeroInt(),
		}
		items[i] = val
		keeper.SetLimitOrderTrancheUser(ctx, items[i])
	}

	return items
}

func TestLimitOrderTrancheUserGet(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNLimitOrderTrancheUser(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetLimitOrderTrancheUser(ctx, item.Address, item.TrancheKey)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(item),
			nullify.Fill(rst),
		)
	}
}

func TestLimitOrderTrancheUserRemove(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	items := createNLimitOrderTrancheUser(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLimitOrderTrancheUserByKey(ctx, item.TrancheKey, item.Address)
		_, found := keeper.GetLimitOrderTrancheUser(ctx, item.Address, item.TrancheKey)
		require.False(t, found)
	}
}

func (s *DexTestSuite) TestGetAllLimitOrders() {
	// WHEN Alice places 2 limit orders
	s.fundAliceBalances(20, 20)
	s.fundBobBalances(20, 20)
	trancheKeyA := s.aliceLimitSells("TokenA", -1, 10)
	trancheKeyB := s.aliceLimitSells("TokenB", 0, 10)
	s.bobLimitSells("TokenA", -1, 10)

	// THEN GetAllLimitOrders returns alice's same two orders
	LOList := s.App.DexKeeper.GetAllLimitOrderTrancheUserForAddress(s.Ctx, s.alice)
	s.Assert().Equal(2, len(LOList))
	s.Assert().Equal(&types.LimitOrderTrancheUser{
		TradePairId:           defaultTradePairID1To0,
		TickIndexTakerToMaker: 1,
		TrancheKey:            trancheKeyA,
		Address:               s.alice.String(),
		SharesOwned:           math.NewInt(10_000_000),
		SharesWithdrawn:       math.NewInt(0),
		SharesCancelled:       math.ZeroInt(),
	},
		LOList[0],
	)
	s.Assert().Equal(&types.LimitOrderTrancheUser{
		TradePairId:           defaultTradePairID0To1,
		TickIndexTakerToMaker: 0,
		TrancheKey:            trancheKeyB,
		Address:               s.alice.String(),
		SharesOwned:           math.NewInt(10_000_000),
		SharesWithdrawn:       math.NewInt(0),
		SharesCancelled:       math.ZeroInt(),
	},
		LOList[1],
	)
}
