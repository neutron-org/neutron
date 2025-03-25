package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v6/testutil/revenue/keeper"
	"github.com/neutron-org/neutron/v6/x/revenue/types"
)

func TestTWAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
	prices, err := keeper.GetAllRewardAssetPrices(ctx)
	require.Nil(t, err)
	require.Equal(t, len(prices), 0)

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyMustNewDecFromStr("10.0"), 1)
	require.Nil(t, err)

	prices, err = keeper.GetAllRewardAssetPrices(ctx)
	require.Nil(t, err)
	require.Equal(t, len(prices), 1)

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyMustNewDecFromStr("20.0"), 11)
	require.Nil(t, err)

	prices, err = keeper.GetAllRewardAssetPrices(ctx)
	require.Nil(t, err)
	require.Equal(t, len(prices), 2)

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyMustNewDecFromStr("20.0"), 21)
	require.Nil(t, err)

	// get twap 11-21
	price, err := keeper.GetTWAPStartingFromTime(ctx, 10)
	require.Nil(t, err)
	require.Equal(t, price, math.LegacyMustNewDecFromStr("20.0"))

	// get twap 0-21
	price, err = keeper.GetTWAPStartingFromTime(ctx, 0)
	require.Nil(t, err)
	require.Equal(t, price, math.LegacyMustNewDecFromStr("15.0"))

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyMustNewDecFromStr("20.0"), 111)
	require.Nil(t, err)

	// get twap 0-111
	// 10 for 10 block = 100 cumulative
	// 20 for 100 block = 2000 cumulative
	// (2000 + 100)/(100 + 10) = ~19.09

	/*
			price
			^
			|
		  20+         +------------------------------------------+
		  19+---------********************************----------+
			|                                                    |
		  10+---------+                  2000                    |
			|   100                                              |
			+----------------------------------------------------> block
			          10                           100           110
	*/
	price, err = keeper.GetTWAPStartingFromTime(ctx, 0)
	require.Nil(t, err)
	require.Equal(t, price, math.LegacyMustNewDecFromStr("19.090909090909090909"))

	prices, err = keeper.GetAllRewardAssetPrices(ctx)
	require.Nil(t, err)
	require.Equal(t, len(prices), 4)

	ctx = ctx.WithBlockTime(time.Unix(types.DefaultTWAPWindow+2, 0))
	// now price at time 1 is outdated
	err = keeper.CleanOutdatedRewardAssetPrices(ctx, ctx.BlockTime().Unix()-types.DefaultTWAPWindow)
	require.Nil(t, err)

	prices, err = keeper.GetAllRewardAssetPrices(ctx)
	require.Nil(t, err)
	require.Equal(t, len(prices), 3)

	price, err = keeper.GetTWAPStartingFromTime(ctx, 0)
	require.Nil(t, err)
	require.Equal(t, price, math.LegacyMustNewDecFromStr("20.0"))
}
