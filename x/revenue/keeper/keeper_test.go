package keeper_test

import (
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"
	"github.com/stretchr/testify/require"

	appconfig "github.com/neutron-org/neutron/v5/app/config"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
)

const (
	val1OperAddr = "neutronvaloper18zawa74y4xv6xg3zv0cstmfl9y38ecurgt4e70"
	val1ConsAddr = "neutronvalcons18zawa74y4xv6xg3zv0cstmfl9y38ecurucx9jw"

	val2OperAddr = "neutronvaloper1x6hw4rnkj4ag97jkdz4srlxzkr7w6pny54qmda"
	val2ConsAddr = "neutronvalcons1x6hw4rnkj4ag97jkdz4srlxzkr7w6pnyqxn8pu"
)

func TestParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

	// assert default params
	params, err := keeper.GetParams(ctx)
	require.Nil(t, err)
	require.Equal(t, params, revenuetypes.DefaultParams())

	// set new params and assert they are changed
	newParams := revenuetypes.DefaultParams()
	newParams.DenomCompensation = "uibcatom"
	err = keeper.SetParams(ctx, newParams)
	require.Nil(t, err)
	params, err = keeper.GetParams(ctx)
	require.Nil(t, err)
	require.Equal(t, "uibcatom", params.DenomCompensation)
	require.Equal(t, revenuetypes.DefaultParams().BaseCompensation, params.BaseCompensation)
	require.Equal(t, revenuetypes.DefaultParams().BlocksPerformanceRequirement, params.BlocksPerformanceRequirement)
	require.Equal(t, revenuetypes.DefaultParams().OracleVotesPerformanceRequirement, params.OracleVotesPerformanceRequirement)
}

func TestValidatorInfo(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 1
	val1Info.CommitedOracleVotesInPeriod = 2

	val2Info := val2Info()
	val1Info.CommitedBlocksInPeriod = 100
	val1Info.CommitedOracleVotesInPeriod = 200

	// set validator infos
	err := keeper.SetValidatorInfo(ctx, []byte(val1Info.ConsensusAddress), val1Info)
	require.Nil(t, err)
	err = keeper.SetValidatorInfo(ctx, []byte(val2Info.ConsensusAddress), val2Info)
	require.Nil(t, err)

	// get all validator info
	valInfos, err := keeper.GetAllValidatorInfo(ctx)
	require.Nil(t, err)
	require.Equal(t, 2, len(valInfos))
	require.Equal(t, val1Info, valInfos[0])
	require.Equal(t, val2Info, valInfos[1])
}

func TestProcessRevenue(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 1000
	val1Info.CommitedOracleVotesInPeriod = 1000

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, []byte(val1Info.ConsensusAddress), val1Info)
	require.Nil(t, err)

	stakingKeeper.EXPECT().GetValidatorByConsAddr(
		gomock.Any(),
		mustConsAddressFromBech32(t, val1Info.ConsensusAddress),
	).Return(stakingtypes.Validator{OperatorAddress: val1OperAddr}, nil)

	// expect one successful SendCoinsFromModuleToAccount call
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			revenuetypes.DefaultDenomCompensation,
			math.NewInt(keeper.CalcBaseRevenueAmount(ctx)))),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx, revenuetypes.DefaultParams(), 1000)
	require.Nil(t, err)
}

func TestProcessRevenueNoReward(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

	// set val1 info as if they haven't committed any blocks and prices
	val1Info := val1Info()

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, []byte(val1Info.ConsensusAddress), val1Info)
	require.Nil(t, err)

	// no SendCoinsFromModuleToAccount calls expected
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err = keeper.ProcessRevenue(ctx, revenuetypes.DefaultParams(), 1000)
	require.Nil(t, err)
}

func TestProcessRevenueMultipleValidators(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

	// define test specific performance requirements
	params := revenuetypes.DefaultParams()
	params.BlocksPerformanceRequirement = &revenuetypes.PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(1, 1), // 0.1 allowed to miss without a fine
		RequiredAtLeast: math.LegacyNewDecWithPrec(8, 1), // not more than 0.2 allowed to miss to get rewards
	}
	params.OracleVotesPerformanceRequirement = &revenuetypes.PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(1, 1), // 0.1 allowed to miss without a fine
		RequiredAtLeast: math.LegacyNewDecWithPrec(8, 1), // not more than 0.2 allowed to miss to get rewards
	}

	// set val1 info as if they have missed 0.15 blocks and prices
	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 850
	val1Info.CommitedOracleVotesInPeriod = 850
	// val2 haven't missed a thing
	val2Info := val2Info()
	val2Info.CommitedBlocksInPeriod = 1000
	val2Info.CommitedOracleVotesInPeriod = 1000

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, []byte(val1Info.ConsensusAddress), val1Info)
	require.Nil(t, err)
	err = keeper.SetValidatorInfo(ctx, []byte(val2Info.ConsensusAddress), val2Info)
	require.Nil(t, err)

	stakingKeeper.EXPECT().GetValidatorByConsAddr(
		gomock.Any(),
		mustConsAddressFromBech32(t, val1Info.ConsensusAddress),
	).Return(stakingtypes.Validator{OperatorAddress: val1OperAddr}, nil)
	stakingKeeper.EXPECT().GetValidatorByConsAddr(
		gomock.Any(),
		mustConsAddressFromBech32(t, val2Info.ConsensusAddress),
	).Return(stakingtypes.Validator{OperatorAddress: val2OperAddr}, nil)

	// expect one successful SendCoinsFromModuleToAccount call for val1 75% of rewards
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			revenuetypes.DefaultDenomCompensation,
			math.LegacyNewDecWithPrec(75, 2).MulInt(math.NewInt(keeper.CalcBaseRevenueAmount(ctx))).RoundInt(),
		)),
	).Times(1).Return(nil)

	// expect one successful SendCoinsFromModuleToAccount call for val2 with full rewards
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val2OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			revenuetypes.DefaultDenomCompensation,
			math.NewInt(keeper.CalcBaseRevenueAmount(ctx)))),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx, params, 1000)
	require.Nil(t, err)
}

func TestProcessSignaturesAndPrices(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

	// known validator (set in keeper below) with 100% performance
	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 1000
	val1Info.CommitedOracleVotesInPeriod = 1000
	// new validator (doesn't exist in keeper state)
	val2Info := val2Info()

	ca1 := mustConsAddressFromBech32(t, val1Info.ConsensusAddress)
	ca2 := mustConsAddressFromBech32(t, val2Info.ConsensusAddress)

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, ca1, val1Info)
	require.Nil(t, err)

	// add vote info from the validator
	ctx = ctx.WithVoteInfos([]abcitypes.VoteInfo{
		// known validator misses a block
		{
			Validator:   abcitypes.Validator{Address: ca1, Power: 10},
			BlockIdFlag: tmtypes.BlockIDFlagAbsent,
		},
		// new validator commits a block
		{
			Validator:   abcitypes.Validator{Address: ca2, Power: 10},
			BlockIdFlag: tmtypes.BlockIDFlagCommit,
		},
	})
	// known validator misses oracle prices update
	voteAggregator.EXPECT().GetPriceForValidator(ca1).Return(nil)
	// new validator commits oracle prices (content doesn't matter, the len of the map does)
	voteAggregator.EXPECT().GetPriceForValidator(ca2).Return(map[slinkytypes.CurrencyPair]*big.Int{{}: big.NewInt(0)})

	err = keeper.RecordValidatorsParticipation(ctx)
	require.Nil(t, err)

	// make sure that the validator votes are processed and recorded
	storedVal1Info, err := keeper.GetValidatorInfo(ctx, ca1) // known val
	require.Nil(t, err)
	require.Equal(t, val1Info.ConsensusAddress, storedVal1Info.ConsensusAddress)
	require.Equal(t, uint64(1000), storedVal1Info.CommitedBlocksInPeriod)      // never missed a block but the last one
	require.Equal(t, uint64(1000), storedVal1Info.CommitedOracleVotesInPeriod) // never missed a block but the last one

	storedVal2Info, err := keeper.GetValidatorInfo(ctx, ca2) // new val
	require.Nil(t, err)
	require.Equal(t, val2Info.ConsensusAddress, storedVal2Info.ConsensusAddress)
	require.Equal(t, uint64(1), storedVal2Info.CommitedBlocksInPeriod)      // all but the last one are missed
	require.Equal(t, uint64(1), storedVal2Info.CommitedOracleVotesInPeriod) // all but the last one are missed
}

func TestEndBlockEmptyPaymentSchedule(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

	g := revenuetypes.DefaultGenesis()
	require.Nil(t, keeper.SetParams(ctx, g.Params))
	require.Nil(t, keeper.SetState(ctx, g.State))

	// init a fresh validator
	val1Info := val1Info()
	ca1 := mustConsAddressFromBech32(t, val1Info.ConsensusAddress)

	// prepare keeper state by setting validator info
	require.Nil(t, keeper.SetValidatorInfo(ctx, ca1, val1Info))

	// add vote info from the validator
	ctx = ctx.WithVoteInfos([]abcitypes.VoteInfo{
		{
			Validator:   abcitypes.Validator{Address: ca1, Power: 10},
			BlockIdFlag: tmtypes.BlockIDFlagCommit,
		},
	})
	// val commits oracle prices (content doesn't matter, the len of the map does)
	voteAggregator.EXPECT().GetPriceForValidator(ca1).Return(map[slinkytypes.CurrencyPair]*big.Int{{}: big.NewInt(0)})

	err := keeper.EndBlock(ctx)
	require.Nil(t, err)

	// make sure that the validator votes are processed and recorded
	storedVal1Info, err := keeper.GetValidatorInfo(ctx, ca1)
	require.Nil(t, err)
	require.Equal(t, val1Info.ConsensusAddress, storedVal1Info.ConsensusAddress)
	require.Equal(t, uint64(1), storedVal1Info.CommitedBlocksInPeriod)
	require.Equal(t, uint64(1), storedVal1Info.CommitedOracleVotesInPeriod)

	// no state updates are expected ever for the empty payment schedule
	state, err := keeper.GetState(ctx)
	require.Nil(t, err)
	require.Equal(t, state.PaymentSchedule.GetCachedValue(), g.State.PaymentSchedule.GetCachedValue())
}

func TestEndBlockMonthlyPaymentSchedule(t *testing.T) {
	t.Run("WithinTheSameMonth", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

		// set monthly payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = revenuetypes.PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_MONTHLY
		g.State.PaymentSchedule = mustNewAnyWithValue(t, &revenuetypes.MonthlyPaymentSchedule{CurrentMonth: 1, CurrentMonthStartBlock: 1})
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		require.Nil(t, keeper.SetState(ctx, g.State))

		// init a fresh validator
		val1Info := val1Info()
		ca1 := mustConsAddressFromBech32(t, val1Info.ConsensusAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, ca1, val1Info)
		require.Nil(t, err)

		// add vote info from the validator
		ctx = ctx.WithVoteInfos([]abcitypes.VoteInfo{
			{
				Validator:   abcitypes.Validator{Address: ca1, Power: 10},
				BlockIdFlag: tmtypes.BlockIDFlagCommit,
			},
		})
		// val commits oracle prices (content doesn't matter, the len of the map does)
		voteAggregator.EXPECT().GetPriceForValidator(ca1).Return(map[slinkytypes.CurrencyPair]*big.Int{{}: big.NewInt(0)})
		// no SendCoinsFromModuleToAccount calls expected
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		err = keeper.EndBlock(ctx.WithBlockHeight(2).WithBlockTime(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
		require.Nil(t, err)

		// make sure that the validator votes are processed and recorded
		storedVal1Info, err := keeper.GetValidatorInfo(ctx, ca1)
		require.Nil(t, err)
		require.Equal(t, val1Info.ConsensusAddress, storedVal1Info.ConsensusAddress)
		require.Equal(t, uint64(1), storedVal1Info.CommitedBlocksInPeriod)
		require.Equal(t, uint64(1), storedVal1Info.CommitedOracleVotesInPeriod)

		// no state updates are expected since the current period (month) hasn't ended yet
		state, err := keeper.GetState(ctx)
		require.Nil(t, err)
		require.Equal(t, state, g.State)
	})

	t.Run("MonthChange", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

		// set monthly payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = revenuetypes.PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_MONTHLY
		g.State.PaymentSchedule = mustNewAnyWithValue(t, &revenuetypes.MonthlyPaymentSchedule{CurrentMonth: 1, CurrentMonthStartBlock: 1})
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		require.Nil(t, keeper.SetState(ctx, g.State))

		// init a validator with 100% performance (the next block will be the 6th one)
		val1Info := val1Info()
		val1Info.CommitedBlocksInPeriod = 5
		val1Info.CommitedOracleVotesInPeriod = 5
		ca1 := mustConsAddressFromBech32(t, val1Info.ConsensusAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, ca1, val1Info)
		require.Nil(t, err)

		// add vote info from the validator
		ctx = ctx.WithVoteInfos([]abcitypes.VoteInfo{
			{
				Validator:   abcitypes.Validator{Address: ca1, Power: 10},
				BlockIdFlag: tmtypes.BlockIDFlagCommit,
			},
		})
		// val commits oracle prices (content doesn't matter, the len of the map does)
		voteAggregator.EXPECT().GetPriceForValidator(ca1).Return(map[slinkytypes.CurrencyPair]*big.Int{{}: big.NewInt(0)})

		stakingKeeper.EXPECT().GetValidatorByConsAddr(
			gomock.Any(),
			mustConsAddressFromBech32(t, val1Info.ConsensusAddress),
		).Return(stakingtypes.Validator{OperatorAddress: val1OperAddr}, nil)

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
		err = keeper.EndBlock(ctx.WithBlockHeight(6).WithBlockTime(time.Date(2000, 2, 1, 0, 0, 0, 0, time.UTC)))
		require.Nil(t, err)

		// make sure that the validator commitments are reset at the new period (month)
		storedVal1Info, err := keeper.GetValidatorInfo(ctx, ca1)
		require.Nil(t, err)
		require.Equal(t, val1Info.ConsensusAddress, storedVal1Info.ConsensusAddress)
		require.Equal(t, uint64(1), storedVal1Info.CommitedBlocksInPeriod)
		require.Equal(t, uint64(1), storedVal1Info.CommitedOracleVotesInPeriod)

		// make sure state is updated to the new period (month)
		state, err := keeper.GetState(ctx)
		require.Nil(t, err)
		newPs := state.PaymentSchedule.GetCachedValue().(*revenuetypes.MonthlyPaymentSchedule)
		require.Equal(t, uint64(2), newPs.CurrentMonth)
		require.Equal(t, uint64(6), newPs.CurrentMonthStartBlock)
	})
}

func TestEndBlockBlockBasedPaymentSchedule(t *testing.T) {
	t.Run("WithinTheSamePeriod", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

		// set block-based payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = revenuetypes.PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_BLOCK_BASED
		g.State.PaymentSchedule = mustNewAnyWithValue(t, &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1})
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		require.Nil(t, keeper.SetState(ctx, g.State))

		// init a fresh validator
		val1Info := val1Info()
		ca1 := mustConsAddressFromBech32(t, val1Info.ConsensusAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, ca1, val1Info)
		require.Nil(t, err)

		// add vote info from the validator
		ctx = ctx.WithVoteInfos([]abcitypes.VoteInfo{
			{
				Validator:   abcitypes.Validator{Address: ca1, Power: 10},
				BlockIdFlag: tmtypes.BlockIDFlagCommit,
			},
		})
		// val commits oracle prices (content doesn't matter, the len of the map does)
		voteAggregator.EXPECT().GetPriceForValidator(ca1).Return(map[slinkytypes.CurrencyPair]*big.Int{{}: big.NewInt(0)})
		// no SendCoinsFromModuleToAccount calls expected
		bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		err = keeper.EndBlock(ctx.WithBlockHeight(2))
		require.Nil(t, err)

		// make sure that the validator votes are processed and recorded
		storedVal1Info, err := keeper.GetValidatorInfo(ctx, ca1)
		require.Nil(t, err)
		require.Equal(t, val1Info.ConsensusAddress, storedVal1Info.ConsensusAddress)
		require.Equal(t, uint64(1), storedVal1Info.CommitedBlocksInPeriod)
		require.Equal(t, uint64(1), storedVal1Info.CommitedOracleVotesInPeriod)

		// no state updates are expected since the current period hasn't ended yet
		state, err := keeper.GetState(ctx)
		require.Nil(t, err)
		require.Equal(t, state, g.State)
	})

	t.Run("PeriodChange", func(t *testing.T) {
		appconfig.GetDefaultConfig()

		ctrl := gomock.NewController(t)
		voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
		stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
		bankKeeper := mock_types.NewMockBankKeeper(ctrl)
		oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

		keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

		// set block-based payment schedule to the module's state and params
		g := revenuetypes.DefaultGenesis()
		g.Params.PaymentScheduleType = revenuetypes.PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_BLOCK_BASED
		g.State.PaymentSchedule = mustNewAnyWithValue(t, &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1})
		require.Nil(t, keeper.SetParams(ctx, g.Params))
		require.Nil(t, keeper.SetState(ctx, g.State))

		// init a validator with 100% performance (the next block will be the 6th one)
		val1Info := val1Info()
		val1Info.CommitedBlocksInPeriod = 5
		val1Info.CommitedOracleVotesInPeriod = 5
		ca1 := mustConsAddressFromBech32(t, val1Info.ConsensusAddress)

		// prepare keeper state by setting validator info
		err := keeper.SetValidatorInfo(ctx, ca1, val1Info)
		require.Nil(t, err)

		// add vote info from the validator
		ctx = ctx.WithVoteInfos([]abcitypes.VoteInfo{
			{
				Validator:   abcitypes.Validator{Address: ca1, Power: 10},
				BlockIdFlag: tmtypes.BlockIDFlagCommit,
			},
		})
		// val commits oracle prices (content doesn't matter, the len of the map does)
		voteAggregator.EXPECT().GetPriceForValidator(ca1).Return(map[slinkytypes.CurrencyPair]*big.Int{{}: big.NewInt(0)})

		stakingKeeper.EXPECT().GetValidatorByConsAddr(
			gomock.Any(),
			mustConsAddressFromBech32(t, val1Info.ConsensusAddress),
		).Return(stakingtypes.Validator{OperatorAddress: val1OperAddr}, nil)

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
		err = keeper.EndBlock(ctx.WithBlockHeight(6))
		require.Nil(t, err)

		// make sure that the validator commitments are reset at the new period
		storedVal1Info, err := keeper.GetValidatorInfo(ctx, ca1)
		require.Nil(t, err)
		require.Equal(t, val1Info.ConsensusAddress, storedVal1Info.ConsensusAddress)
		require.Equal(t, uint64(1), storedVal1Info.CommitedBlocksInPeriod)
		require.Equal(t, uint64(1), storedVal1Info.CommitedOracleVotesInPeriod)

		// make sure state is updated to the new period
		state, err := keeper.GetState(ctx)
		require.Nil(t, err)
		newPs := state.PaymentSchedule.GetCachedValue().(*revenuetypes.BlockBasedPaymentSchedule)
		require.Equal(t, uint64(5), newPs.BlocksPerPeriod)
		require.Equal(t, uint64(6), newPs.CurrentPeriodStartBlock)
	})
}

func val1Info() revenuetypes.ValidatorInfo {
	return revenuetypes.ValidatorInfo{
		ConsensusAddress: val1ConsAddr,
	}
}

func val2Info() revenuetypes.ValidatorInfo {
	return revenuetypes.ValidatorInfo{
		ConsensusAddress: val2ConsAddr,
	}
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

func mustNewAnyWithValue(
	t *testing.T,
	m proto.Message,
) *codectypes.Any {
	v, err := codectypes.NewAnyWithValue(m)
	require.Nil(t, err)
	return v
}

func mustConsAddressFromBech32(
	t *testing.T,
	address string,
) sdktypes.ConsAddress {
	ca, err := sdktypes.ConsAddressFromBech32(address)
	require.Nil(t, err)
	return ca
}
