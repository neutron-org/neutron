package dex_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v4/testutil/dex/keeper"
	"github.com/neutron-org/neutron/v4/x/dex"
	"github.com/neutron-org/neutron/v4/x/dex/types"
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
				SharesCancelled:       math.NewInt(0),
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
				SharesCancelled:       math.NewInt(0),
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
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		PoolCount: 2,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.DexKeeper(t)
	dex.InitGenesis(ctx, *k, genesisState)
	got := dex.ExportGenesis(ctx, *k)

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

	require.NotNil(t, got)

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
