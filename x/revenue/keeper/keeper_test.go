package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/golang/mock/gomock"
	vetypes "github.com/skip-mev/slinky/abci/ve/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appconfig "github.com/neutron-org/neutron/v6/app/config"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v6/testutil/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

const (
	val1OperAddr = "neutronvaloper18zawa74y4xv6xg3zv0cstmfl9y38ecurgt4e70"
	val2OperAddr = "neutronvaloper1x6hw4rnkj4ag97jkdz4srlxzkr7w6pny54qmda"
)

func TestParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")

	// assert default params
	params, err := keeper.GetParams(ctx)
	require.Nil(t, err)
	require.Equal(t, params, revenuetypes.DefaultParams())

	// set new params and assert they are changed
	newParams := revenuetypes.DefaultParams()
	newParams.TwapWindow = revenuetypes.DefaultTWAPWindow + 10
	err = keeper.SetParams(ctx, newParams)
	require.Nil(t, err)
	params, err = keeper.GetParams(ctx)
	require.Nil(t, err)
	require.Equal(t, revenuetypes.DefaultTWAPWindow+10, params.TwapWindow)
	require.Equal(t, revenuetypes.DefaultParams().RewardQuote.Amount, params.RewardQuote.Amount)
	require.Equal(t, revenuetypes.DefaultParams().RewardQuote.Asset, params.RewardQuote.Asset)
	require.Equal(t, revenuetypes.DefaultParams().BlocksPerformanceRequirement, params.BlocksPerformanceRequirement)
	require.Equal(t, revenuetypes.DefaultParams().OracleVotesPerformanceRequirement, params.OracleVotesPerformanceRequirement)
}

func TestValidatorInfo(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")

	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 1
	val1Info.CommitedOracleVotesInPeriod = 2
	val1Info.InActiveValsetForBlocksInPeriod = 3

	val2Info := val2Info()
	val2Info.CommitedBlocksInPeriod = 100
	val2Info.CommitedOracleVotesInPeriod = 200
	val2Info.InActiveValsetForBlocksInPeriod = 300

	// set validator infos
	err := keeper.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1Info.ValOperAddress), val1Info)
	require.Nil(t, err)
	err = keeper.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val2Info.ValOperAddress), val2Info)
	require.Nil(t, err)

	// get all validator info
	valInfos, err := keeper.GetAllValidatorInfo(ctx)
	require.Nil(t, err)
	require.Equal(t, 2, len(valInfos))
	require.Equal(t, val1Info, valInfos[1])
	require.Equal(t, val2Info, valInfos[0])
}

func TestProcessRevenue(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)

	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         1000,
		CurrentPeriodStartBlock: 1,
	}
	ctx = ctx.WithBlockHeight(1001)

	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 1000
	val1Info.CommitedOracleVotesInPeriod = 1000
	val1Info.InActiveValsetForBlocksInPeriod = 1000

	// prepare keeper state
	err = keeper.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1Info.ValOperAddress), val1Info)
	require.Nil(t, err)

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
	require.Nil(t, err)

	bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
		Base:       "untrn", Symbol: "NTRN",
	}, true).AnyTimes()

	baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx)
	require.Nil(t, err)

	// expect one successful SendCoinsFromModuleToAccount call
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			params.RewardAsset,
			baseRevenueAmount)),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx, revenuetypes.DefaultParams(), ps)
	require.Nil(t, err)
}

func TestProcessRevenueNoReward(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")

	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         1000,
		CurrentPeriodStartBlock: 1,
	}
	ctx = ctx.WithBlockHeight(1001)

	// set val1 info as if they haven't committed any blocks and prices
	val1Info := val1Info()
	val1Info.InActiveValsetForBlocksInPeriod = 1000

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1Info.ValOperAddress), val1Info)
	require.Nil(t, err)

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
	require.Nil(t, err)

	bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
		Base:       "untrn", Symbol: "NTRN",
	}, true).AnyTimes()

	// no SendCoinsFromModuleToAccount calls expected
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err = keeper.ProcessRevenue(ctx, revenuetypes.DefaultParams(), ps)
	require.Nil(t, err)
}

func TestProcessRevenuePaymentScheduleTypeChange(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)

	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         1000,
		CurrentPeriodStartBlock: 1,
	}
	ctx = ctx.WithBlockHeight(501) // 500/1000 == 1/2 of the payment period

	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 500
	val1Info.CommitedOracleVotesInPeriod = 500
	val1Info.InActiveValsetForBlocksInPeriod = 500

	// prepare keeper state
	err = keeper.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1Info.ValOperAddress), val1Info)
	require.Nil(t, err)

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
	require.Nil(t, err)

	bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
		Base:       "untrn", Symbol: "NTRN",
	}, true).AnyTimes()

	baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx)
	require.Nil(t, err)
	expectedRevenueAmount := baseRevenueAmount.Quo(math.NewInt(2)) // for 1/2 of the payment period

	// expect one successful SendCoinsFromModuleToAccount call
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			params.RewardAsset,
			expectedRevenueAmount)),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx, revenuetypes.DefaultParams(), ps)
	require.Nil(t, err)
}

func TestProcessRevenueLateValsetJoin(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")
	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)

	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         1000,
		CurrentPeriodStartBlock: 1,
	}
	ctx = ctx.WithBlockHeight(1001)

	// set val1 info as if they have committed all blocks and prices they could but joined
	// the valset 300 blocks after the start of the payment period
	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 700          // 700/1000
	val1Info.CommitedOracleVotesInPeriod = 700     // 700/1000
	val1Info.InActiveValsetForBlocksInPeriod = 700 // 700/1000

	// prepare keeper state
	err = keeper.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1Info.ValOperAddress), val1Info)
	require.Nil(t, err)

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
	require.Nil(t, err)

	bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
		Base:       "untrn", Symbol: "NTRN",
	}, true).AnyTimes()

	baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx)
	require.Nil(t, err)

	// val is only eligible of 70% of rewards due to joining valset 30% of period late
	expectedRevenueAmount := math.LegacyNewDecFromInt(baseRevenueAmount).
		Mul(math.LegacyNewDecWithPrec(7, 1)).
		TruncateInt()

	// expect one successful SendCoinsFromModuleToAccount call
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			params.RewardAsset,
			expectedRevenueAmount)),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx, revenuetypes.DefaultParams(), ps)
	require.Nil(t, err)
}

func TestProcessRevenueReducedByAllFactors(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")

	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         2000,
		CurrentPeriodStartBlock: 1,
	}
	ctx = ctx.WithBlockHeight(1501) // 1500/2000 = 3/4 of the payment period

	// set test specific performance requirements
	params := revenuetypes.DefaultParams()
	params.BlocksPerformanceRequirement = &revenuetypes.PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(1, 1), // 0.1 allowed to miss without a fine
		RequiredAtLeast: math.LegacyNewDecWithPrec(8, 1), // not more than 0.2 allowed to miss to get rewards
	}
	params.OracleVotesPerformanceRequirement = &revenuetypes.PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(1, 1), // 0.1 allowed to miss without a fine
		RequiredAtLeast: math.LegacyNewDecWithPrec(8, 1), // not more than 0.2 allowed to miss to get rewards
	}
	require.Nil(t, keeper.SetParams(ctx, params))

	// set val1 info as if they have committed 850/1000 blocks and prices and joined the valset
	// 1000 blocks after the start of the payment period
	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 850
	val1Info.CommitedOracleVotesInPeriod = 850
	val1Info.InActiveValsetForBlocksInPeriod = 1000 // 1000/1500

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1Info.ValOperAddress), val1Info)
	require.Nil(t, err)

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
	require.Nil(t, err)

	bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
		Base:       "untrn", Symbol: "NTRN",
	}, true).AnyTimes()

	baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx)
	require.Nil(t, err)

	expectedRevenueAmount := math.LegacyNewDecFromInt(baseRevenueAmount).
		Mul(math.LegacyNewDecWithPrec(666666667, 9)). // 0.666666667, been in valset for 2/3 of payment period
		Mul(math.LegacyNewDecWithPrec(75, 2)).        // 0.75, missed 15% of blocks and oracle votes
		Mul(math.LegacyNewDecWithPrec(75, 2)).        // 0.75, for 3/4 of the payment period
		TruncateInt()

	// expect one successful SendCoinsFromModuleToAccount call
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			params.RewardAsset,
			expectedRevenueAmount)),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx, params, ps)
	require.Nil(t, err)
}

func TestProcessRevenueMultipleValidators(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")

	ps := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         1000,
		CurrentPeriodStartBlock: 1,
	}
	ctx = ctx.WithBlockHeight(1001)

	// set test specific performance requirements
	params := revenuetypes.DefaultParams()
	params.BlocksPerformanceRequirement = &revenuetypes.PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(1, 1), // 0.1 allowed to miss without a fine
		RequiredAtLeast: math.LegacyNewDecWithPrec(8, 1), // not more than 0.2 allowed to miss to get rewards
	}
	params.OracleVotesPerformanceRequirement = &revenuetypes.PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(1, 1), // 0.1 allowed to miss without a fine
		RequiredAtLeast: math.LegacyNewDecWithPrec(8, 1), // not more than 0.2 allowed to miss to get rewards
	}
	require.Nil(t, keeper.SetParams(ctx, params))

	// set val1 info as if they have missed 0.15 blocks and prices
	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 850
	val1Info.CommitedOracleVotesInPeriod = 850
	val1Info.InActiveValsetForBlocksInPeriod = 1000
	// val2 haven't missed a thing
	val2Info := val2Info()
	val2Info.CommitedBlocksInPeriod = 1000
	val2Info.CommitedOracleVotesInPeriod = 1000
	val2Info.InActiveValsetForBlocksInPeriod = 1000

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val1Info.ValOperAddress), val1Info)
	require.Nil(t, err)
	err = keeper.SetValidatorInfo(ctx, mustValAddressFromBech32(t, val2Info.ValOperAddress), val2Info)
	require.Nil(t, err)

	err = keeper.CalcNewRewardAssetPrice(ctx, math.LegacyOneDec(), ctx.BlockTime().Unix())
	require.Nil(t, err)

	bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: 6, Aliases: []string{"NTRN"}}},
		Base:       "untrn", Symbol: "NTRN",
	}, true).AnyTimes()

	baseRevenueAmount, err := keeper.CalcBaseRevenueAmount(ctx)
	require.Nil(t, err)

	// expect one successful SendCoinsFromModuleToAccount call for val1 75% of rewards
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val1OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			params.RewardAsset,
			math.LegacyNewDecWithPrec(75, 2).MulInt(baseRevenueAmount).RoundInt(),
		)),
	).Times(1).Return(nil)

	// expect one successful SendCoinsFromModuleToAccount call for val2 with full rewards
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(),
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.AccAddress(mustGetFromBech32(t, val2OperAddr, "neutronvaloper")),
		sdktypes.NewCoins(sdktypes.NewCoin(
			params.RewardAsset,
			baseRevenueAmount)),
	).Times(1).Return(nil)

	err = keeper.ProcessRevenue(ctx, params, ps)
	require.Nil(t, err)
}

func TestProcessSignaturesAndPrices(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")

	// known validator (set in keeper below) with 100% performance
	val1Info := val1Info()
	val1Info.CommitedBlocksInPeriod = 1000
	val1Info.CommitedOracleVotesInPeriod = 1000
	val1Info.InActiveValsetForBlocksInPeriod = 1000
	// new validator (doesn't exist in keeper state)
	val2Info := val2Info()

	va1 := mustValAddressFromBech32(t, val1Info.ValOperAddress)
	va2 := mustValAddressFromBech32(t, val2Info.ValOperAddress)

	// prepare keeper state
	err := keeper.SetValidatorInfo(ctx, va1, val1Info)
	require.Nil(t, err)

	// vote info from the validators
	votes := []revenuetypes.ValidatorParticipation{
		// known validator misses a block
		{
			ValOperAddress:      va1,
			BlockVote:           tmtypes.BlockIDFlagAbsent,
			OracleVoteExtension: vetypes.OracleVoteExtension{Prices: map[uint64][]byte{}},
		},
		// new validator commits a block and oracle prices
		{
			ValOperAddress: va2,
			BlockVote:      tmtypes.BlockIDFlagCommit,
			// content doesn't matter, the len of the map does
			OracleVoteExtension: vetypes.OracleVoteExtension{Prices: map[uint64][]byte{0: {}}},
		},
	}

	err = keeper.RecordValidatorsParticipation(ctx, votes)
	require.Nil(t, err)

	// make sure that the validator votes are processed and recorded
	storedVal1Info, err := keeper.GetValidatorInfo(ctx, va1) // known val
	require.Nil(t, err)
	require.Equal(t, val1Info.ValOperAddress, storedVal1Info.ValOperAddress)
	require.Equal(t, uint64(1000), storedVal1Info.CommitedBlocksInPeriod)          // never missed a block but the last one
	require.Equal(t, uint64(1000), storedVal1Info.CommitedOracleVotesInPeriod)     // never missed a block but the last one
	require.Equal(t, uint64(1001), storedVal1Info.InActiveValsetForBlocksInPeriod) // been in valset for all blocks

	storedVal2Info, err := keeper.GetValidatorInfo(ctx, va2) // new val
	require.Nil(t, err)
	require.Equal(t, val2Info.ValOperAddress, storedVal2Info.ValOperAddress)
	require.Equal(t, uint64(1), storedVal2Info.CommitedBlocksInPeriod)          // all but the last one are missed
	require.Equal(t, uint64(1), storedVal2Info.CommitedOracleVotesInPeriod)     // all but the last one are missed
	require.Equal(t, uint64(1), storedVal2Info.InActiveValsetForBlocksInPeriod) // just joined active valset
}

func TestPrecisionConversion(t *testing.T) {
	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)
	keeper, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")

	exp := uint32(6)
	// real NTRN<>USD price from slinky at height 21466053
	slinkyPrice := int64(15357076)
	pairDecimals := uint64(8)

	oracleKeeper.EXPECT().GetPriceForCurrencyPair(gomock.Any(), slinkytypes.CurrencyPair{
		Base:  "NTRN",
		Quote: "USD",
	}).Return(oracletypes.QuotePrice{Price: math.NewInt(slinkyPrice)}, nil)
	oracleKeeper.EXPECT().GetDecimalsForCurrencyPair(gomock.Any(), slinkytypes.CurrencyPair{
		Base:  "NTRN",
		Quote: "USD",
	}).Return(pairDecimals, nil)
	bankKeeper.EXPECT().GetDenomMetaData(gomock.Any(), "untrn").Return(banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{Denom: "ntrn", Exponent: exp, Aliases: []string{"NTRN"}}},
		Base:       "untrn", Symbol: "NTRN",
	}, true).AnyTimes()

	err := keeper.UpdateRewardAssetPrice(ctx)
	assert.Nil(t, err)

	twap, err := keeper.GetTWAP(ctx)
	assert.Nil(t, err)
	// expected_twap = slinky_price / 10^pair_decimals = 15357076 / 10^8 = 0.15357076
	assert.Equal(t, 0.15357076, twap.MustFloat64())

	// expected_base_revenue = reward_quote_amount / twap * 10^exp = 2500 / 0.15357076 * 10^6 ≈ 16279140638untrn
	expectedBaseRevenue := math.NewInt(16279140638)
	baseRevenue, err := keeper.CalcBaseRevenueAmount(ctx)
	assert.Nil(t, err)
	assert.Equal(t, expectedBaseRevenue, baseRevenue)
}

func val1Info() revenuetypes.ValidatorInfo {
	return revenuetypes.ValidatorInfo{
		ValOperAddress: val1OperAddr,
	}
}

func val2Info() revenuetypes.ValidatorInfo {
	return revenuetypes.ValidatorInfo{
		ValOperAddress: val2OperAddr,
	}
}

func mustGetFromBech32(
	t *testing.T,
	bech32str string,
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
