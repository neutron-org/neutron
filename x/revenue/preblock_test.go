package revenue_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/golang/mock/gomock"
	compression "github.com/skip-mev/slinky/abci/strategies/codec"
	"github.com/stretchr/testify/require"

	appconfig "github.com/neutron-org/neutron/v6/app/config"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v6/testutil/revenue/keeper"
	"github.com/neutron-org/neutron/v6/x/revenue"
	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

const (
	val1OperAddr = "neutronvaloper18zawa74y4xv6xg3zv0cstmfl9y38ecurgt4e70"
)

var (
	veCodec = compression.NewCompressionVoteExtensionCodec(
		compression.NewDefaultVoteExtensionCodec(),
		compression.NewZLibCompressor(),
	)
	ecCodec = compression.NewCompressionExtendedCommitCodec(
		compression.NewDefaultExtendedCommitCodec(),
		compression.NewZStdCompressor(),
	)
)

func TestPaymentScheduleCheckEmptyPaymentSchedule(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
	preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

	g := revenuetypes.DefaultGenesis()
	require.Nil(t, keeper.SetParams(ctx, g.Params))
	psi := &revenuetypes.EmptyPaymentSchedule{}
	require.Nil(t, keeper.SetPaymentSchedule(ctx, psi.IntoPaymentSchedule()))

	// init a fresh validator
	val1Info := val1Info()
	va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

	// prepare keeper state by setting validator info
	require.Nil(t, keeper.SetValidatorInfo(ctx, va1, val1Info))

	// no payment schedule updates are ever expected for the empty payment schedule
	for _, height := range []int64{1, 10, 1000, 100000, 1000000000} {
		err := preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(height))
		require.Nil(t, err)

		gotPsi, err := keeper.GetPaymentScheduleI(ctx)
		require.Nil(t, err)
		require.Equal(t, psi, gotPsi)
	}
}

func TestPaymentScheduleCheckMonthlyPaymentSchedule(t *testing.T) {
	t.Run("WithinTheSameMonth", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set monthly payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
			PaymentScheduleType: &revenuetypes.PaymentScheduleType_MonthlyPaymentScheduleType{
				MonthlyPaymentScheduleType: &revenuetypes.MonthlyPaymentScheduleType{},
			},
		}
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		psi := (&revenuetypes.MonthlyPaymentSchedule{
			CurrentMonthStartBlock:   1,
			CurrentMonthStartBlockTs: uint64(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()), //nolint:gosec
		})
		require.Nil(t, keeper.SetPaymentSchedule(ctx, psi.IntoPaymentSchedule()))

		// init a fresh validator
		val1Info := val1Info()
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		// expect no revenue distribution within the same period
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// no payment schedule updates are expected since the current period (month) hasn't ended yet
		for _, day := range []int{1, 10, 20, 31} {
			err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(2).WithBlockTime(time.Date(2000, 1, day, 0, 0, 0, 0, time.UTC)))
			require.Nil(t, err)

			gotPsi, err := keeper.GetPaymentScheduleI(ctx)
			require.Nil(t, err)
			require.Equal(t, psi, gotPsi)
		}
	})

	t.Run("MonthChange", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set monthly payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
			PaymentScheduleType: &revenuetypes.PaymentScheduleType_MonthlyPaymentScheduleType{
				MonthlyPaymentScheduleType: &revenuetypes.MonthlyPaymentScheduleType{},
			},
		}
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		psi := (&revenuetypes.MonthlyPaymentSchedule{
			CurrentMonthStartBlock:   1,
			CurrentMonthStartBlockTs: uint64(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()), //nolint:gosec
		})
		require.Nil(t, keeper.SetPaymentSchedule(ctx, psi.IntoPaymentSchedule()))

		// init a validator with 100% performance (the next block will be the 6th one)
		val1Info := val1Info()
		val1Info.CommitedBlocksInPeriod = 5
		val1Info.CommitedOracleVotesInPeriod = 5
		val1Info.InActiveValsetForBlocksInPeriod = 5
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
		require.Nil(t, err)

		bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
			DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
			Base:       "untrn", Symbol: "NTRN",
		}, true).AnyTimes()

		baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx)
		require.Nil(t, err)

		// expect one successful SendCoinsFromModuleToAccount call for val1 with full rewards
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
			gomock.Any(),
			revenuetypes.RevenueTreasuryPoolName,
			sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
			sdktypes.NewCoins(sdktypes.NewCoin(
				g.Params.RewardAsset,
				baseRevenueAmount)),
		).Times(1).Return(nil)

		// next block in the next month with expected revenue distribution
		err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(6).WithBlockTime(time.Date(2000, 2, 1, 0, 0, 0, 0, time.UTC)))
		require.Nil(t, err)

		// make sure payment schedule is updated to the new period (month)
		newPsi, err := keeper.GetPaymentScheduleI(ctx)
		require.Nil(t, err)
		newPs := newPsi.(*revenuetypes.MonthlyPaymentSchedule)
		require.Equal(t, time.February, time.Unix(int64(newPs.CurrentMonthStartBlockTs), 0).Month()) //nolint:gosec
		require.Equal(t, uint64(6), newPs.CurrentMonthStartBlock)

		// make sure validators' info is reset
		info, err := keeper.GetAllValidatorInfo(ctx)
		require.Nil(t, err)
		require.Equal(t, 0, len(info))
	})

	t.Run("EarlyPeriodEndOnPaymentScheduleTypeChange", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set monthly payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
			PaymentScheduleType: &revenuetypes.PaymentScheduleType_MonthlyPaymentScheduleType{
				MonthlyPaymentScheduleType: &revenuetypes.MonthlyPaymentScheduleType{},
			},
		}
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		psi := (&revenuetypes.MonthlyPaymentSchedule{
			CurrentMonthStartBlock:   1,
			CurrentMonthStartBlockTs: uint64(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()), //nolint:gosec
		})
		require.Nil(t, keeper.SetPaymentSchedule(ctx, psi.IntoPaymentSchedule()))

		// init a validator with 100% performance (the next block will be the 6th one)
		val1Info := val1Info()
		val1Info.CommitedBlocksInPeriod = 5
		val1Info.CommitedOracleVotesInPeriod = 5
		val1Info.InActiveValsetForBlocksInPeriod = 5
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
		require.Nil(t, err)

		// update payment schedule type to the empty one in module params
		g.Params.PaymentScheduleType.PaymentScheduleType = &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{
			EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
		}
		require.Nil(t, keeper.SetParams(ctx, g.Params))

		bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
			DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
			Base:       "untrn", Symbol: "NTRN",
		}, true).AnyTimes()

		baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx)
		require.Nil(t, err)
		// 50% of revenue for 1/2 of the payment period (see ctx.WithBlockTime(...) below)
		expectedRevenueAmount := baseRevenueAmount.Quo(math.NewInt(2))

		// expect one successful SendCoinsFromModuleToAccount call for val1
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
			gomock.Any(),
			revenuetypes.RevenueTreasuryPoolName,
			sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
			sdktypes.NewCoins(sdktypes.NewCoin(
				g.Params.RewardAsset,
				expectedRevenueAmount)),
		).Times(1).Return(nil)

		// next block is not in the next period but still revenue distribution is expected
		// due to the change of the payment schedule type
		err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(6).WithBlockTime(time.Date(2000, 1, 16, 12, 0, 0, 0, time.UTC)))
		require.Nil(t, err)

		// make sure payment schedule is updated in accordance with module params
		newPsi, err := keeper.GetPaymentScheduleI(ctx)
		require.Nil(t, err)
		require.IsType(t, &revenuetypes.EmptyPaymentSchedule{}, newPsi)

		// make sure validators' info is reset
		info, err := keeper.GetAllValidatorInfo(ctx)
		require.Nil(t, err)
		require.Equal(t, 0, len(info))
	})
}

func TestPaymentScheduleCheckBlockBasedPaymentSchedule(t *testing.T) {
	t.Run("WithinTheSamePeriod", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set block-based payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
			PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
				BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
			},
		}
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		psi := (&revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1})
		require.Nil(t, keeper.SetPaymentSchedule(ctx, psi.IntoPaymentSchedule()))

		// init a fresh validator
		val1Info := val1Info()
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		// no SendCoinsFromModuleToAccount calls expected
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// no payment schedule updates are expected since the current period hasn't ended yet
		for _, height := range []int64{2, 3, 4, 5} {
			err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(height))
			require.Nil(t, err)

			gotPsi, err := keeper.GetPaymentScheduleI(ctx)
			require.Nil(t, err)
			require.Equal(t, psi, gotPsi)
		}
	})

	t.Run("PeriodChange", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set block-based payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
			PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
				BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
			},
		}
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		psi := (&revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1})
		require.Nil(t, keeper.SetPaymentSchedule(ctx, psi.IntoPaymentSchedule()))

		// init a validator with 100% performance (the next block will be the 6th one)
		val1Info := val1Info()
		val1Info.CommitedBlocksInPeriod = 5
		val1Info.CommitedOracleVotesInPeriod = 5
		val1Info.InActiveValsetForBlocksInPeriod = 5
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
		require.Nil(t, err)

		bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
			DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
			Base:       "untrn", Symbol: "NTRN",
		}, true).AnyTimes()

		baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx)
		require.Nil(t, err)

		// expect one successful SendCoinsFromModuleToAccount call for val1 with full rewards
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
			gomock.Any(),
			revenuetypes.RevenueTreasuryPoolName,
			sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
			sdktypes.NewCoins(sdktypes.NewCoin(
				g.Params.RewardAsset,
				baseRevenueAmount)),
		).Times(1).Return(nil)

		// next block in the next period with expected revenue distribution
		err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(6))
		require.Nil(t, err)

		// make sure payment schedule is updated to the new period
		newPsi, err := keeper.GetPaymentScheduleI(ctx)
		require.Nil(t, err)
		newPs := newPsi.(*revenuetypes.BlockBasedPaymentSchedule)
		require.Equal(t, uint64(5), newPs.BlocksPerPeriod)
		require.Equal(t, uint64(6), newPs.CurrentPeriodStartBlock)

		// make sure validators' info is reset
		info, err := keeper.GetAllValidatorInfo(ctx)
		require.Nil(t, err)
		require.Equal(t, 0, len(info))
	})

	t.Run("EarlyPeriodEndOnPaymentScheduleTypeChange", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set block-based payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
			PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
				BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
			},
		}
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		psi := (&revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 10, CurrentPeriodStartBlock: 1})
		require.Nil(t, keeper.SetPaymentSchedule(ctx, psi.IntoPaymentSchedule()))

		// init a validator with 100% performance (the next block will be the 6th one)
		val1Info := val1Info()
		val1Info.CommitedBlocksInPeriod = 5
		val1Info.CommitedOracleVotesInPeriod = 5
		val1Info.InActiveValsetForBlocksInPeriod = 5
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
			DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
			Base:       "untrn", Symbol: "NTRN",
		}, true).AnyTimes()

		err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
		require.Nil(t, err)

		// update payment schedule type to the empty one in module params
		g.Params.PaymentScheduleType.PaymentScheduleType = &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{
			EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
		}
		require.Nil(t, keeper.SetParams(ctx, g.Params))

		baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx)
		require.Nil(t, err)
		// 50% of revenue for 1/2 of the payment period (see ctx.WithBlockHeight(6) below)
		expectedRevenueAmount := baseRevenueAmount.Quo(math.NewInt(2))

		// expect one successful SendCoinsFromModuleToAccount call for val1
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
			gomock.Any(),
			revenuetypes.RevenueTreasuryPoolName,
			sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
			sdktypes.NewCoins(sdktypes.NewCoin(
				g.Params.RewardAsset,
				expectedRevenueAmount)),
		).Times(1).Return(nil)

		// next block is not in the next period but still revenue distribution is expected
		// due to the change of the payment schedule type
		err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(6))
		require.Nil(t, err)

		// make sure payment schedule is updated in accordance with module params
		newPsi, err := keeper.GetPaymentScheduleI(ctx)
		require.Nil(t, err)
		require.IsType(t, &revenuetypes.EmptyPaymentSchedule{}, newPsi)

		// make sure validators' info is reset
		info, err := keeper.GetAllValidatorInfo(ctx)
		require.Nil(t, err)
		require.Equal(t, 0, len(info))
	})
}

func val1Info() revenuetypes.ValidatorInfo {
	return revenuetypes.ValidatorInfo{
		ValOperAddress: val1OperAddr,
	}
}

func mustGetFromBech32(
	t *testing.T,
	bech32str string, //nolint:unparam
	prefix string, //nolint:unparam
) []byte {
	b, err := sdktypes.GetFromBech32(bech32str, prefix)
	require.Nil(t, err)
	return b
}

func mustValAddressFromBech32(
	t *testing.T,
	address string,
) sdktypes.ValAddress {
	va, err := sdktypes.ValAddressFromBech32(address)
	require.Nil(t, err)
	return va
}
