package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"

	appconfig "github.com/neutron-org/neutron/v5/app/config"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	vetypes "github.com/skip-mev/slinky/abci/ve/types"
	"github.com/stretchr/testify/require"
)

const (
	val1OperAddr = "neutronvaloper18zawa74y4xv6xg3zv0cstmfl9y38ecurgt4e70"
	val1ConsAddr = "neutronvalcons18zawa74y4xv6xg3zv0cstmfl9y38ecurucx9jw"

	val2OperAddr = "neutronvaloper1x6hw4rnkj4ag97jkdz4srlxzkr7w6pny54qmda"
	val2ConsAddr = "neutronvalcons1x6hw4rnkj4ag97jkdz4srlxzkr7w6pnyqxn8pu"
)

func TestParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, stakingKeeper, bankKeeper, oracleKeeper, "")

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
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, stakingKeeper, bankKeeper, oracleKeeper, "")

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
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, stakingKeeper, bankKeeper, oracleKeeper, "")

	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 1000
	val1Info.CommitedOracleVotesInPeriod = 1000

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, []byte(val1Info.ConsensusAddress), val1Info)
	require.Nil(t, err)

	err = keeper.SaveCumulativePrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
	require.Nil(t, err)

	params, err := keeper.GetParams(ctx)
	require.Nil(t, err)

	baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx, int64(params.BaseCompensation))
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
			baseRevenueAmount)),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx, revenuetypes.DefaultParams(), 1000)
	require.Nil(t, err)
}

func TestProcessRevenueNoReward(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, stakingKeeper, bankKeeper, oracleKeeper, "")

	// set val1 info as if they haven't committed any blocks and prices
	val1Info := val1Info()

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, []byte(val1Info.ConsensusAddress), val1Info)
	require.Nil(t, err)

	err = keeper.SaveCumulativePrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
	require.Nil(t, err)

	// no SendCoinsFromModuleToAccount calls expected
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err = keeper.ProcessRevenue(ctx, revenuetypes.DefaultParams(), 1000)
	require.Nil(t, err)
}

func TestProcessRevenueMultipleValidators(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, stakingKeeper, bankKeeper, oracleKeeper, "")

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

	err = keeper.SaveCumulativePrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
	require.Nil(t, err)

	baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx, int64(params.BaseCompensation))
	require.Nil(t, err)

	// expect one successful SendCoinsFromModuleToAccount call for val1 75% of rewards
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			revenuetypes.DefaultDenomCompensation,
			math.LegacyNewDecWithPrec(75, 2).MulInt(baseRevenueAmount).RoundInt(),
		)),
	).Times(1).Return(nil)

	// expect one successful SendCoinsFromModuleToAccount call for val2 with full rewards
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val2OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			revenuetypes.DefaultDenomCompensation,
			baseRevenueAmount)),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx, params, 1000)
	require.Nil(t, err)
}

func TestProcessSignaturesAndPrices(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, stakingKeeper, bankKeeper, oracleKeeper, "")

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

	// vote info from the validators
	votes := []revenuetypes.ValidatorParticipation{
		// known validator misses a block
		{
			ConsAddress:         ca1,
			BlockVote:           tmtypes.BlockIDFlagAbsent,
			OracleVoteExtension: vetypes.OracleVoteExtension{Prices: map[uint64][]byte{}},
		},
		// new validator commits a block and oracle prices
		{
			ConsAddress: ca2,
			BlockVote:   tmtypes.BlockIDFlagCommit,
			// content doesn't matter, the len of the map does
			OracleVoteExtension: vetypes.OracleVoteExtension{Prices: map[uint64][]byte{0: {}}},
		},
	}

	err = keeper.RecordValidatorsParticipation(ctx, votes)
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

func mustConsAddressFromBech32(
	t *testing.T,
	address string,
) sdktypes.ConsAddress {
	ca, err := sdktypes.ConsAddressFromBech32(address)
	require.Nil(t, err)
	return ca
}
