package revenue_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"
	appconfig "github.com/neutron-org/neutron/v5/app/config"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	"github.com/neutron-org/neutron/v5/x/revenue"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	compression "github.com/skip-mev/slinky/abci/strategies/codec"
	"github.com/stretchr/testify/require"
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
	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, "")
	preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

	g := revenuetypes.DefaultGenesis()
	require.Nil(t, keeper.SetParams(ctx, g.Params))
	require.Nil(t, keeper.SetState(ctx, g.State))

	// init a fresh validator
	val1Info := val1Info()
	va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

	// prepare keeper state by setting validator info
	require.Nil(t, keeper.SetValidatorInfo(ctx, va1, val1Info))

	// no state updates are expected ever for the empty payment schedule
	for _, height := range []int64{1, 10, 1000, 100000, 1000000000} {
		err := preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(height))
		require.Nil(t, err)

		state, err := keeper.GetState(ctx)
		require.Nil(t, err)
		require.Equal(t, state.PaymentSchedule.GetCachedValue(), g.State.PaymentSchedule.GetCachedValue())
	}
}

func TestPaymentScheduleCheckMonthlyPaymentSchedule(t *testing.T) {
	t.Run("WithinTheSameMonth", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set monthly payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.Params_MonthlyPaymentScheduleType{
			MonthlyPaymentScheduleType: &revenuetypes.MonthlyPaymentScheduleType{},
		}
		g.State.PaymentSchedule = mustNewAnyWithValue(t, &revenuetypes.MonthlyPaymentSchedule{
			CurrentMonth: 1, CurrentMonthStartBlock: 1,
		})
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		require.Nil(t, keeper.SetState(ctx, g.State))

		// init a fresh validator
		val1Info := val1Info()
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		// expect no revenue distribution within the same period
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// no state updates are expected since the current period (month) hasn't ended yet
		for _, day := range []int{1, 10, 20, 31} {
			err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(2).WithBlockTime(time.Date(2000, 1, day, 0, 0, 0, 0, time.UTC)))
			require.Nil(t, err)

			state, err := keeper.GetState(ctx)
			require.Nil(t, err)
			require.Equal(t, state, g.State)
		}
	})

	t.Run("MonthChange", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set monthly payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.Params_MonthlyPaymentScheduleType{
			MonthlyPaymentScheduleType: &revenuetypes.MonthlyPaymentScheduleType{},
		}
		g.State.PaymentSchedule = mustNewAnyWithValue(t, &revenuetypes.MonthlyPaymentSchedule{
			CurrentMonth: 1, CurrentMonthStartBlock: 1,
		})
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		require.Nil(t, keeper.SetState(ctx, g.State))

		// init a validator with 100% performance (the next block will be the 6th one)
		val1Info := val1Info()
		val1Info.CommitedBlocksInPeriod = 5
		val1Info.CommitedOracleVotesInPeriod = 5
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		// expect one successful SendCoinsFromModuleToAccount call for val1 with full rewards
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
			gomock.Any(),
			revenuetypes.RevenueTreasuryPoolName,
			sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
			sdktypes.NewCoins(sdktypes.NewCoin(
				revenuetypes.DefaultDenomCompensation,
				math.NewInt(keeper.CalcBaseRevenueAmount(ctx)))),
		).Times(1).Return(nil)

		// next block in the next month with expected revenue distribution
		err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(6).WithBlockTime(time.Date(2000, 2, 1, 0, 0, 0, 0, time.UTC)))
		require.Nil(t, err)

		// make sure state is updated to the new period (month)
		state, err := keeper.GetState(ctx)
		require.Nil(t, err)
		newPs := state.PaymentSchedule.GetCachedValue().(*revenuetypes.MonthlyPaymentSchedule)
		require.Equal(t, uint64(2), newPs.CurrentMonth)
		require.Equal(t, uint64(6), newPs.CurrentMonthStartBlock)

		// make sure validators' info is reset
		info, err := keeper.GetValidatorInfo(ctx, va1)
		require.Nil(t, err)
		require.Equal(t, uint64(0), info.CommitedBlocksInPeriod)
		require.Equal(t, uint64(0), info.CommitedOracleVotesInPeriod)
	})
}

func TestPaymentScheduleCheckBasedPaymentSchedule(t *testing.T) {
	t.Run("WithinTheSamePeriod", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set block-based payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.Params_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
		}
		g.State.PaymentSchedule = mustNewAnyWithValue(t, &revenuetypes.BlockBasedPaymentSchedule{
			BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1,
		})
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		require.Nil(t, keeper.SetState(ctx, g.State))

		// init a fresh validator
		val1Info := val1Info()
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		// no SendCoinsFromModuleToAccount calls expected
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		// no state updates are expected since the current period hasn't ended yet
		for _, height := range []int64{2, 3, 4, 5} {
			err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(height))
			require.Nil(t, err)

			state, err := keeper.GetState(ctx)
			require.Nil(t, err)
			require.Equal(t, state, g.State)
		}
	})

	t.Run("PeriodChange", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, "")
		preBlock := revenue.NewPreBlockHandler(keeper, stakingKeeper, veCodec, ecCodec)

		// set block-based payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = &revenuetypes.Params_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
		}
		g.State.PaymentSchedule = mustNewAnyWithValue(t, &revenuetypes.BlockBasedPaymentSchedule{
			BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1,
		})
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		require.Nil(t, keeper.SetState(ctx, g.State))

		// init a validator with 100% performance (the next block will be the 6th one)
		val1Info := val1Info()
		val1Info.CommitedBlocksInPeriod = 5
		val1Info.CommitedOracleVotesInPeriod = 5
		va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, va1, val1Info)
		require.Nil(t, err)

		// expect one successful SendCoinsFromModuleToAccount call for val1 with full rewards
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
			gomock.Any(),
			revenuetypes.RevenueTreasuryPoolName,
			sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
			sdktypes.NewCoins(sdktypes.NewCoin(
				revenuetypes.DefaultDenomCompensation,
				math.NewInt(keeper.CalcBaseRevenueAmount(ctx)))),
		).Times(1).Return(nil)

		// next block in the next period with expected revenue distribution
		err = preBlock.PaymentScheduleCheck(ctx.WithBlockHeight(6))
		require.Nil(t, err)

		// make sure state is updated to the new period
		state, err := keeper.GetState(ctx)
		require.Nil(t, err)
		newPs := state.PaymentSchedule.GetCachedValue().(*revenuetypes.BlockBasedPaymentSchedule)
		require.Equal(t, uint64(5), newPs.BlocksPerPeriod)
		require.Equal(t, uint64(6), newPs.CurrentPeriodStartBlock)

		// make sure validators' info is reset
		info, err := keeper.GetValidatorInfo(ctx, va1)
		require.Nil(t, err)
		require.Equal(t, uint64(0), info.CommitedBlocksInPeriod)
		require.Equal(t, uint64(0), info.CommitedOracleVotesInPeriod)
	})
}

func val1Info() revenuetypes.ValidatorInfo {
	return revenuetypes.ValidatorInfo{
		ValOperAddress: val1OperAddr,
	}
}

func mustNewAnyWithValue(
	t *testing.T,
	m proto.Message,
) *codectypes.Any {
	v, err := codectypes.NewAnyWithValue(m)
	require.Nil(t, err)
	return v
}

func mustGetFromBech32(
	t *testing.T,
	bech32str string,
	prefix string,
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
