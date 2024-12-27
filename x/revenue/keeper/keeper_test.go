package keeper_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	appconfig "github.com/neutron-org/neutron/v5/app/config"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper)

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
	require.Equal(t, revenuetypes.DefaultParams().OracleVoteWeight, params.OracleVoteWeight)
	require.Equal(t, revenuetypes.DefaultParams().PerformanceThreshold, params.PerformanceThreshold)
	require.Equal(t, revenuetypes.DefaultParams().AllowedMissed, params.AllowedMissed)
}

func TestValidatorInfo(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper)

	val1Info := val1Info()
	val1Info.MissedBlocksInMonth = 1
	val1Info.MissedOracleVotesInMonth = 2

	val2Info := val2Info()
	val1Info.MissedBlocksInMonth = 100
	val1Info.MissedOracleVotesInMonth = 200

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

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper)

	val1Info := val1Info()

	// prepare keeper state
	err := keeper.SetState(ctx, revenuetypes.State{BlockCounter: 1000})
	require.Nil(t, err)
	err = keeper.SetValidatorInfo(ctx, []byte(val1Info.ConsensusAddress), val1Info)
	require.Nil(t, err)

	// expect one successful SendCoinsFromModuleToAccount call
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		ctx,
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1Info.OperatorAddress, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			revenuetypes.DefaultDenomCompensation,
			math.NewInt(keeper.GetBaseNTRNAmount(ctx)))),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx)
	require.Nil(t, err)
}

func TestProcessRevenueNoReward(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper)

	// set val1 info as if they have missed all blocks and prices
	val1Info := val1Info()
	val1Info.MissedBlocksInMonth = 1000
	val1Info.MissedOracleVotesInMonth = 1000

	// prepare keeper state
	err := keeper.SetState(ctx, revenuetypes.State{BlockCounter: 1000})
	require.Nil(t, err)
	err = keeper.SetValidatorInfo(ctx, []byte(val1Info.ConsensusAddress), val1Info)
	require.Nil(t, err)

	// no SendCoinsFromModuleToAccount calls expected
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		ctx,
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Times(0)

	err = keeper.ProcessRevenue(ctx)
	require.Nil(t, err)
}

func TestProcessRevenueMultipleValidators(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper)

	// set test specific inactivity params
	newParams := revenuetypes.DefaultParams()
	newParams.AllowedMissed = math.LegacyNewDecWithPrec(1, 1)        // 0.1 allowed to miss without a fine
	newParams.PerformanceThreshold = math.LegacyNewDecWithPrec(2, 1) // not more than 0.2 allowed to miss to get rewards
	err := keeper.SetParams(ctx, newParams)
	require.Nil(t, err)

	// set val1 info as if they have missed 0.15 blocks and prices
	val1Info := val1Info()
	val1Info.MissedBlocksInMonth = 150
	val1Info.MissedOracleVotesInMonth = 150
	// val2 haven't missed a thing
	val2Info := val2Info()

	// prepare keeper state
	err = keeper.SetState(ctx, revenuetypes.State{BlockCounter: 1000})
	require.Nil(t, err)
	err = keeper.SetValidatorInfo(ctx, []byte(val1Info.ConsensusAddress), val1Info)
	require.Nil(t, err)
	err = keeper.SetValidatorInfo(ctx, []byte(val2Info.ConsensusAddress), val2Info)
	require.Nil(t, err)

	// expect one successful SendCoinsFromModuleToAccount call for val1 75% of rewards
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		ctx,
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1Info.OperatorAddress, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			revenuetypes.DefaultDenomCompensation,
			math.LegacyNewDecWithPrec(75, 2).MulInt(math.NewInt(keeper.GetBaseNTRNAmount(ctx))).RoundInt(),
		)),
	).Times(1).Return(nil)

	// expect one successful SendCoinsFromModuleToAccount call for val2 with full rewards
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		ctx,
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val2Info.OperatorAddress, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			revenuetypes.DefaultDenomCompensation,
			math.NewInt(keeper.GetBaseNTRNAmount(ctx)))),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx)
	require.Nil(t, err)
}

func TestProcessSignaturesAndPrices(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper)

	val1Info := val1Info() // known validator (set in keeper below) with 100% performance
	val2Info := val2Info() // new validator (doesn't exist in keeper state)
	ca1 := mustConsAddressFromBech32(t, val1Info.ConsensusAddress)
	ca2 := mustConsAddressFromBech32(t, val2Info.ConsensusAddress)

	// prepare keeper state
	err := keeper.SetState(ctx, revenuetypes.State{BlockCounter: 1000})
	require.Nil(t, err)
	err = keeper.SetValidatorInfo(ctx, ca1, val1Info)
	require.Nil(t, err)

	// add vote info from the validator
	ctx = ctx.WithVoteInfos([]abcitypes.VoteInfo{
		// known validator misses a block
		{
			Validator: abcitypes.Validator{
				Address: ca1,
				Power:   10,
			},
			BlockIdFlag: tmtypes.BlockIDFlagAbsent,
		},
		// new validator commits a block
		{
			Validator: abcitypes.Validator{
				Address: ca2,
				Power:   10,
			},
			BlockIdFlag: tmtypes.BlockIDFlagCommit,
		},
	})
	// known validator misses oracle prices update
	voteAggregator.EXPECT().GetPriceForValidator(ca1).Return(nil)
	// new validator commits oracle prices (content doesn't matter, the len of the map does)
	voteAggregator.EXPECT().GetPriceForValidator(ca2).Return(map[slinkytypes.CurrencyPair]*big.Int{{}: big.NewInt(0)})

	stakingKeeper.EXPECT().GetValidatorByConsAddr(
		ctx,
		ca2,
	).Return(stakingtypes.Validator{OperatorAddress: val2Info.OperatorAddress}, nil)

	err = keeper.ProcessSignatures(ctx, 1000)
	require.Nil(t, err)
	err = keeper.ProcessOracleVotes(ctx, 1000)
	require.Nil(t, err)

	// make sure that the validator votes are processed and keeper's state is updated
	// TODO: refactor to get each val by address
	valInfos, err := keeper.GetAllValidatorInfo(ctx)
	require.Nil(t, err)
	require.Equal(t, 2, len(valInfos))

	// known val
	require.Equal(t, val1Info.ConsensusAddress, valInfos[1].ConsensusAddress)
	require.Equal(t, uint64(1), valInfos[1].MissedBlocksInMonth)      // never missed a block but the last one
	require.Equal(t, uint64(1), valInfos[1].MissedOracleVotesInMonth) // never missed a block but the last one
	// new val
	require.Equal(t, val2Info.ConsensusAddress, valInfos[0].ConsensusAddress)
	require.Equal(t, uint64(1000), valInfos[0].MissedBlocksInMonth)      // all but the last one are missed
	require.Equal(t, uint64(1000), valInfos[0].MissedOracleVotesInMonth) // all but the last one are missed
}

func val1Info() revenuetypes.ValidatorInfo {
	return revenuetypes.ValidatorInfo{
		OperatorAddress:  "neutronvaloper18zawa74y4xv6xg3zv0cstmfl9y38ecurgt4e70",
		ConsensusAddress: "neutronvalcons18zawa74y4xv6xg3zv0cstmfl9y38ecurucx9jw",
	}
}

func val2Info() revenuetypes.ValidatorInfo {
	return revenuetypes.ValidatorInfo{
		OperatorAddress:  "neutronvaloper1x6hw4rnkj4ag97jkdz4srlxzkr7w6pny54qmda",
		ConsensusAddress: "neutronvalcons1x6hw4rnkj4ag97jkdz4srlxzkr7w6pnyqxn8pu",
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
