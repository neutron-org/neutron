package types_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v6/testutil/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

func TestMonthlyPaymentSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	_, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")

	// a monthly schedule for January with first block height = 1
	s := &revenuetypes.MonthlyPaymentSchedule{
		CurrentMonthStartBlock:   1,
		CurrentMonthStartBlockTs: uint64(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()), //nolint:gosec
	}
	// such schedule's period is expected to be ended when January ends
	assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC))))
	assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.January, 31, 23, 59, 59, 0, time.UTC))))
	assert.Equal(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC))))
	// blocks in the period are counted from the block with height 1
	assert.Equal(t, uint64(0), s.TotalBlocksInPeriod(ctx.WithBlockHeight(1)))
	assert.Equal(t, uint64(10), s.TotalBlocksInPeriod(ctx.WithBlockHeight(11)))
	assert.Equal(t, uint64(100), s.TotalBlocksInPeriod(ctx.WithBlockHeight(101)))

	// start a new period in February with the first block = 101
	s.StartNewPeriod(ctx.
		WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC)).
		WithBlockHeight(101),
	)
	// such schedule's period is expected to be ended when February ends
	assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC))))
	assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.February, 28, 23, 59, 59, 0, time.UTC))))
	assert.Equal(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.March, 1, 0, 0, 0, 0, time.UTC))))
	// blocks in the period are counted from the block with height 101
	assert.Equal(t, uint64(0), s.TotalBlocksInPeriod(ctx.WithBlockHeight(101)))
	assert.Equal(t, uint64(10), s.TotalBlocksInPeriod(ctx.WithBlockHeight(111)))
	assert.Equal(t, uint64(100), s.TotalBlocksInPeriod(ctx.WithBlockHeight(201)))

	// 4 full days passed in Jan = (4*24) / (31*24) ≈ 12.9%
	s.StartNewPeriod(ctx.WithBlockTime(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, math.LegacyNewDecWithPrec(129032258064516129, 18), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.January, 5, 0, 0, 0, 0, time.UTC))))
	// 14.5 days passed in Feb = (14*24 + 12) / (29*24) = 50%
	s.StartNewPeriod(ctx.WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, math.LegacyNewDecWithPrec(5, 1), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.February, 15, 12, 0, 0, 0, time.UTC))))
	// 24 days, 12 hours and 30 minutes passed in March = (24*24 + 12.5) / (31*24) ≈ 79%
	s.StartNewPeriod(ctx.WithBlockTime(time.Date(2000, time.March, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, math.LegacyNewDecWithPrec(790994623655913978, 18), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.March, 25, 12, 30, 0, 0, time.UTC))))
	// 59 minutes and 59 seconds passed in March = 3599/3600 / (31*24) ≈ 0.13%
	s.StartNewPeriod(ctx.WithBlockTime(time.Date(2000, time.March, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, math.LegacyNewDecWithPrec(1343712664277180, 18), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.March, 1, 0, 59, 59, 0, time.UTC))))
	// 1 hour passed in March = 1 / (31*24) = 0.0013%
	s.StartNewPeriod(ctx.WithBlockTime(time.Date(2000, time.March, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, math.LegacyNewDecWithPrec(1344086021505376, 18), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.March, 1, 1, 0, 0, 0, time.UTC))))
	// 30 days, 23 hours and 59 minutes passed in March = (30*24 + 23 + 59/60) / (31*24) ≈ 99.99%
	s.StartNewPeriod(ctx.WithBlockTime(time.Date(2000, time.March, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, math.LegacyNewDecWithPrec(999977598566308244, 18), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.March, 31, 23, 59, 0, 0, time.UTC))))
	// more than month passed, still = 100%, no overflow
	s.StartNewPeriod(ctx.WithBlockTime(time.Date(2000, time.March, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.May, 1, 0, 0, 0, 0, time.UTC))))

	// border cases for all months
	for startMonth := range []time.Month{
		time.January,
		time.February,
		time.March,
		time.April,
		time.May,
		time.June,
		time.July,
		time.August,
		time.September,
		time.October,
		time.November,
		time.December,
	} {
		daysInCurrentMonth := time.Date(
			2000,
			time.Month(startMonth)+1,
			0, 0, 0, 0, 0,
			time.UTC,
		).Day()

		s.StartNewPeriod(ctx.WithBlockTime(time.Date(2000, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)))
		assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC))))
		assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.Month(startMonth), daysInCurrentMonth, 23, 59, 59, 0, time.UTC))))
		assert.Equal(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.Month(startMonth)+1, 1, 0, 0, 0, 0, time.UTC))))
	}

	for _, loc := range []*time.Location{
		time.FixedZone("BakerIsland", -12*3600),
		time.FixedZone("AzoresIslands", -1*3600),
		time.UTC,
		time.FixedZone("Berlin", 1*3600),
		time.FixedZone("LineIslands", 14*3600),
	} {
		t.Run("TimezoneOperations"+loc.String(), func(t *testing.T) {
			time.Local = loc // set the timezone for the test

			t.Run("BeginningOfMonth", func(t *testing.T) {
				// timestamp is in the beginning of the month: 2000-01-01 00:00:01
				s = &revenuetypes.MonthlyPaymentSchedule{
					CurrentMonthStartBlockTs: uint64(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()), //nolint:gosec
				}

				// check that the current month is January regardless of the timezone
				assert.Equal(t, time.January, s.CurrentMonth())

				// make sure period end edge cases result in the same regardless of the timezone
				assert.Equal(t, false, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC))))
				assert.Equal(t, false, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.January, 31, 23, 59, 59, 0, time.UTC))))
				assert.Equal(t, true, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 1, 0, time.UTC))))

				// make sure effective period progress edge cases result in the same regardless of the timezone
				assert.Equal(t, math.LegacyZeroDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC))))
				assert.Equal(t, math.LegacyNewDecWithPrec(5, 1), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.January, 16, 12, 0, 0, 0, time.UTC))))
				assert.Equal(t, math.LegacyNewDecWithPrec(1344086021505376, 18), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.January, 1, 1, 0, 0, 0, time.UTC))))
				assert.Equal(t, math.LegacyNewDecWithPrec(999977598566308244, 18), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.January, 31, 23, 59, 0, 0, time.UTC))))
				assert.Equal(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC))))

				// make sure the next period start timestamp and month are the same regardless of the timezone
				nextPeriodStartDate := time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC)
				s.StartNewPeriod(ctx.WithBlockTime(nextPeriodStartDate))
				assert.Equal(t, time.February, s.CurrentMonth())
				assert.Equal(t, uint64(nextPeriodStartDate.Unix()), s.CurrentMonthStartBlockTs) //nolint:gosec
			})

			t.Run("EndOfMonth", func(t *testing.T) {
				// timestamp is in the end of the month: 2000-01-31 23:59:00
				s = &revenuetypes.MonthlyPaymentSchedule{
					CurrentMonthStartBlockTs: uint64(time.Date(2000, time.January, 31, 23, 59, 0, 0, time.UTC).Unix()), //nolint:gosec
				}

				// check that the current month is January regardless of the timezone
				assert.Equal(t, time.January, s.CurrentMonth())

				// make sure period end edge cases result in the same regardless of the timezone
				assert.Equal(t, false, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.January, 31, 23, 59, 0, 0, time.UTC))))
				assert.Equal(t, false, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.January, 31, 23, 59, 59, 0, time.UTC))))
				assert.Equal(t, true, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 1, 0, time.UTC))))

				// make sure effective period progress edge cases result in the same regardless of the timezone
				assert.Equal(t, math.LegacyZeroDec(), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.January, 31, 23, 59, 0, 0, time.UTC))))
				assert.Equal(t, math.LegacyNewDecWithPrec(22028076463560, 18), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.January, 31, 23, 59, 59, 0, time.UTC))))
				assert.Equal(t, math.LegacyNewDecWithPrec(22401433691756, 18), s.EffectivePeriodProgress(ctx.WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC))))

				// make sure the next period start timestamp and month are the same regardless of the timezone
				nextPeriodStartDate := time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC)
				s.StartNewPeriod(ctx.WithBlockTime(nextPeriodStartDate))
				assert.Equal(t, time.February, s.CurrentMonth())
				assert.Equal(t, uint64(nextPeriodStartDate.Unix()), s.CurrentMonthStartBlockTs) //nolint:gosec
			})
		})
	}
}

func TestBlockBasedPaymentSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	_, ctx := testkeeper.RevenueKeeper(t, bankKeeper, oracleKeeper, "")

	// a block based schedule of 100 blocks period and period start block = 1
	s := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         100,
		CurrentPeriodStartBlock: 1,
	}
	// such schedule's period is expected to be ended after block 100
	assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockHeight(1)))
	assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockHeight(100)))
	assert.Equal(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockHeight(101)))
	// blocks in the period are counted from the block with height 1
	assert.Equal(t, uint64(0), s.TotalBlocksInPeriod(ctx.WithBlockHeight(1)))
	assert.Equal(t, uint64(10), s.TotalBlocksInPeriod(ctx.WithBlockHeight(11)))
	assert.Equal(t, uint64(100), s.TotalBlocksInPeriod(ctx.WithBlockHeight(101)))

	// start a new period with the first block = 101
	s.StartNewPeriod(ctx.WithBlockHeight(101))
	// such schedule's period is expected to be ended after block 200
	assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockHeight(101)))
	assert.NotEqual(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockHeight(200)))
	assert.Equal(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockHeight(201)))
	// blocks in the period are counted from the block with height 101
	assert.Equal(t, uint64(0), s.TotalBlocksInPeriod(ctx.WithBlockHeight(101)))
	assert.Equal(t, uint64(10), s.TotalBlocksInPeriod(ctx.WithBlockHeight(111)))
	assert.Equal(t, uint64(100), s.TotalBlocksInPeriod(ctx.WithBlockHeight(201)))

	// block 151 in period 101-201 = 50/100 = 50%
	assert.Equal(t, math.LegacyNewDecWithPrec(5, 1), s.EffectivePeriodProgress(ctx.WithBlockHeight(151)))
	// block 102 in period 101-201 = 1/100 = 1%
	assert.Equal(t, math.LegacyNewDecWithPrec(1, 2), s.EffectivePeriodProgress(ctx.WithBlockHeight(102)))
	// block 200 in period 101-201 = 99/100 = 99%
	assert.Equal(t, math.LegacyNewDecWithPrec(99, 2), s.EffectivePeriodProgress(ctx.WithBlockHeight(200)))
	// set much longer period
	s = &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 100000, CurrentPeriodStartBlock: 1}
	// block 2 in period 1-100001 = 1/100000 = 0.001%
	assert.Equal(t, math.LegacyNewDecWithPrec(1, 5), s.EffectivePeriodProgress(ctx.WithBlockHeight(2)))
	// block 100000 in period 1-100001 = 99999/100000 = 99.999%
	assert.Equal(t, math.LegacyNewDecWithPrec(99999, 5), s.EffectivePeriodProgress(ctx.WithBlockHeight(100000)))
	// block 100500 in period 1-100001, still 100%, no overflow
	assert.Equal(t, math.LegacyOneDec(), s.EffectivePeriodProgress(ctx.WithBlockHeight(100500)))
}

func TestPaymentScheduleTypeMatch(t *testing.T) {
	eps := &revenuetypes.EmptyPaymentSchedule{}
	assert.True(t, eps.MatchesType(&revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{}}))
	assert.False(t, eps.MatchesType(&revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{}))
	assert.False(t, eps.MatchesType(&revenuetypes.PaymentScheduleType_MonthlyPaymentScheduleType{}))

	bbps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 1}
	assert.True(t, bbps.MatchesType(&revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
		BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 1},
	}))
	assert.False(t, bbps.MatchesType(&revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
		BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 10},
	}))
	assert.False(t, bbps.MatchesType(&revenuetypes.PaymentScheduleType_MonthlyPaymentScheduleType{}))
	assert.False(t, bbps.MatchesType(&revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{}))

	mps := &revenuetypes.MonthlyPaymentSchedule{}
	assert.True(t, mps.MatchesType(&revenuetypes.PaymentScheduleType_MonthlyPaymentScheduleType{MonthlyPaymentScheduleType: &revenuetypes.MonthlyPaymentScheduleType{}}))
	assert.False(t, mps.MatchesType(&revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{}))
	assert.False(t, mps.MatchesType(&revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{}))
}
