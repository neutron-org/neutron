package dex_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v6/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		LimitOrderTrancheUserList: []*types.LimitOrderTrancheUser{
			{
				TradePairId: &types.TradePairID{
					TakerDenom: "TokenA",
					MakerDenom: "TokenB",
				},
				TickIndexTakerToMaker: 1,
				TrancheKey:            "0",
				Address:               "fakeAddr",
				SharesOwned:           math.NewInt(10),
				SharesWithdrawn:       math.NewInt(0),
			},
			{
				TradePairId: &types.TradePairID{
					TakerDenom: "TokenB",
					MakerDenom: "TokenA",
				},
				TickIndexTakerToMaker: 20,
				TrancheKey:            "0",
				Address:               "fakeAddr",
				SharesOwned:           math.NewInt(10),
				SharesWithdrawn:       math.NewInt(0),
			},
		},
		TickLiquidityList: []*types.TickLiquidity{
			{
				Liquidity: &types.TickLiquidity_LimitOrderTranche{
					LimitOrderTranche: types.MustNewLimitOrderTranche(
						"TokenB",
						"TokenA",
						"0",
						0,
						math.ZeroInt(),
						math.ZeroInt(),
						math.ZeroInt(),
						math.ZeroInt(),
					),
				},
			},
			{
				Liquidity: &types.TickLiquidity_LimitOrderTranche{
					LimitOrderTranche: types.MustNewLimitOrderTranche(
						"TokenB",
						"TokenA",
						"0",
						0,
						math.ZeroInt(),
						math.ZeroInt(),
						math.ZeroInt(),
						math.ZeroInt(),
					),
				},
			},
			{
				Liquidity: &types.TickLiquidity_LimitOrderTranche{
					LimitOrderTranche: types.MustNewLimitOrderTranche(
						"TokenB",
						"TokenA",
						"0",
						0,
						math.ZeroInt(),
						math.ZeroInt(),
						math.ZeroInt(),
						math.ZeroInt(),
						time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
					),
				},
			},
			{
				Liquidity: &types.TickLiquidity_LimitOrderTranche{
					LimitOrderTranche: types.MustNewLimitOrderTranche(
						"TokenB",
						"TokenA",
						"0",
						0,
						math.ZeroInt(),
						math.ZeroInt(),
						math.ZeroInt(),
						math.ZeroInt(),
						time.Date(2024, 2, 1, 1, 0, 0, 0, time.UTC),
					),
				},
			},
			{
				Liquidity: &types.TickLiquidity_PoolReserves{
					PoolReserves: types.MustNewPoolReserves(
						&types.PoolReservesKey{
							TradePairId:           types.MustNewTradePairID("TokenA", "TokenB"),
							TickIndexTakerToMaker: 0,
							Fee:                   1,
						},
					),
				},
			},
		},
		InactiveLimitOrderTrancheList: []*types.LimitOrderTranche{
			{
				Key: &types.LimitOrderTrancheKey{
					TradePairId: &types.TradePairID{
						TakerDenom: "TokenA",
						MakerDenom: "TokenB",
					},
					TickIndexTakerToMaker: 0,
					TrancheKey:            "0",
				},
			},
			{
				Key: &types.LimitOrderTrancheKey{
					TradePairId: &types.TradePairID{
						TakerDenom: "TokenA",
						MakerDenom: "TokenB",
					},
					TickIndexTakerToMaker: 1,
					TrancheKey:            "1",
				},
			},
		},
		PoolMetadataList: []types.PoolMetadata{
			{
				PairId: types.MustNewPairID("TokenA", "TokenB"),
				Tick:   0,
				Fee:    1,
				Id:     0,
			},
			{
				PairId: types.MustNewPairID("TokenA", "TokenB"),
				Tick:   1,
				Fee:    1,
				Id:     1,
			},
		},
		PoolCount: 2,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.DexKeeper(t)
	dex.InitGenesis(ctx, *k, genesisState)
	got := dex.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	// check that LimitorderExpirations are recreated correctly
	expectedLimitOrderExpirations := []*types.LimitOrderExpiration{
		{
			ExpirationTime: *genesisState.TickLiquidityList[2].GetLimitOrderTranche().ExpirationTime,
			TrancheRef:     genesisState.TickLiquidityList[2].GetLimitOrderTranche().Key.KeyMarshal(),
		},
		{
			ExpirationTime: *genesisState.TickLiquidityList[3].GetLimitOrderTranche().ExpirationTime,
			TrancheRef:     genesisState.TickLiquidityList[3].GetLimitOrderTranche().Key.KeyMarshal(),
		},
	}
	loExpirations := k.GetAllLimitOrderExpiration(ctx)
	require.Equal(t, *expectedLimitOrderExpirations[0], *loExpirations[0])
	require.Equal(t, *expectedLimitOrderExpirations[1], *loExpirations[1])
	require.Equal(t, len(expectedLimitOrderExpirations), len(loExpirations))

	// Check that poolID refs works

	_, found := k.GetPool(ctx, types.MustNewPairID("TokenA", "TokenB"), 0, 1)
	require.True(t, found)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.LimitOrderTrancheUserList, got.LimitOrderTrancheUserList)
	require.ElementsMatch(t, genesisState.TickLiquidityList, got.TickLiquidityList)
	require.ElementsMatch(
		t,
		genesisState.InactiveLimitOrderTrancheList,
		got.InactiveLimitOrderTrancheList,
	)
	require.ElementsMatch(t, genesisState.PoolMetadataList, got.PoolMetadataList)
	require.Equal(t, genesisState.PoolCount, got.PoolCount)
	// this line is used by starport scaffolding # genesis/test/assert
}
