package types_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/assert"
)

func TestMonthlyPaymentSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	_, ctx := testkeeper.RevenueKeeper(t, stakingKeeper, bankKeeper, "")

	// a monthly schedule for January with first block height = 1
	s := &revenuetypes.MonthlyPaymentSchedule{
		CurrentMonth:           1,
		CurrentMonthStartBlock: 1,
	}
	// such schedule's period is expected to be ended when January ends
	assert.False(t, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC))))
	assert.False(t, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.January, 31, 0, 0, 0, 0, time.UTC))))
	assert.True(t, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC))))
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
	assert.False(t, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC))))
	assert.False(t, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.February, 28, 0, 0, 0, 0, time.UTC))))
	assert.True(t, s.PeriodEnded(ctx.WithBlockTime(time.Date(2000, time.March, 1, 0, 0, 0, 0, time.UTC))))
	// blocks in the period are counted from the block with height 101
	assert.Equal(t, uint64(0), s.TotalBlocksInPeriod(ctx.WithBlockHeight(101)))
	assert.Equal(t, uint64(10), s.TotalBlocksInPeriod(ctx.WithBlockHeight(111)))
	assert.Equal(t, uint64(100), s.TotalBlocksInPeriod(ctx.WithBlockHeight(201)))
}

func TestBlockBasedPaymentSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	_, ctx := testkeeper.RevenueKeeper(t, stakingKeeper, bankKeeper, "")

	// a block based schedule of 100 blocks period and period start block = 1
	s := &revenuetypes.BlockBasedPaymentSchedule{
		BlocksPerPeriod:         100,
		CurrentPeriodStartBlock: 1,
	}
	// such schedule's period is expected to be ended after block 100
	assert.False(t, s.PeriodEnded(ctx.WithBlockHeight(1)))
	assert.False(t, s.PeriodEnded(ctx.WithBlockHeight(100)))
	assert.True(t, s.PeriodEnded(ctx.WithBlockHeight(101)))
	// blocks in the period are counted from the block with height 1
	assert.Equal(t, uint64(0), s.TotalBlocksInPeriod(ctx.WithBlockHeight(1)))
	assert.Equal(t, uint64(10), s.TotalBlocksInPeriod(ctx.WithBlockHeight(11)))
	assert.Equal(t, uint64(100), s.TotalBlocksInPeriod(ctx.WithBlockHeight(101)))

	// start a new period with the first block = 101
	s.StartNewPeriod(ctx.
		WithBlockHeight(101),
	)
	// such schedule's period is expected to be ended after block 200
	assert.False(t, s.PeriodEnded(ctx.WithBlockHeight(101)))
	assert.False(t, s.PeriodEnded(ctx.WithBlockHeight(200)))
	assert.True(t, s.PeriodEnded(ctx.WithBlockHeight(201)))
	// blocks in the period are counted from the block with height 101
	assert.Equal(t, uint64(0), s.TotalBlocksInPeriod(ctx.WithBlockHeight(101)))
	assert.Equal(t, uint64(10), s.TotalBlocksInPeriod(ctx.WithBlockHeight(111)))
	assert.Equal(t, uint64(100), s.TotalBlocksInPeriod(ctx.WithBlockHeight(201)))
}

func TestPaymentScheduleTypeMatch(t *testing.T) {
	eps := &revenuetypes.EmptyPaymentSchedule{}
	assert.True(t, eps.MatchesType(&revenuetypes.Params_EmptyPaymentScheduleType{EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{}}))
	assert.False(t, eps.MatchesType(&revenuetypes.Params_BlockBasedPaymentScheduleType{}))
	assert.False(t, eps.MatchesType(&revenuetypes.Params_MonthlyPaymentScheduleType{}))

	bbps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 1}
	assert.True(t, bbps.MatchesType(&revenuetypes.Params_BlockBasedPaymentScheduleType{
		BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 1},
	}))
	assert.False(t, bbps.MatchesType(&revenuetypes.Params_BlockBasedPaymentScheduleType{
		BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 10},
	}))
	assert.False(t, bbps.MatchesType(&revenuetypes.Params_MonthlyPaymentScheduleType{}))
	assert.False(t, bbps.MatchesType(&revenuetypes.Params_EmptyPaymentScheduleType{}))

	mps := &revenuetypes.MonthlyPaymentSchedule{CurrentMonth: 1, CurrentMonthStartBlock: 1}
	assert.True(t, mps.MatchesType(&revenuetypes.Params_MonthlyPaymentScheduleType{MonthlyPaymentScheduleType: &revenuetypes.MonthlyPaymentScheduleType{}}))
	assert.False(t, mps.MatchesType(&revenuetypes.Params_BlockBasedPaymentScheduleType{}))
	assert.False(t, mps.MatchesType(&revenuetypes.Params_EmptyPaymentScheduleType{}))
}
