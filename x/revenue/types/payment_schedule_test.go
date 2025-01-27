package types_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/revenue/types"
	testkeeper "github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
)

func TestMonthlyPaymentSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	_, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

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
	voteAggregator := mock_types.NewMockVoteAggregator(ctrl)
	stakingKeeper := mock_types.NewMockStakingKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	oracleKeeper := mock_types.NewMockOracleKeeper(ctrl)

	_, ctx := testkeeper.RevenueKeeper(t, voteAggregator, stakingKeeper, bankKeeper, oracleKeeper)

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
