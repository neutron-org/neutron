package types

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

var (
	_ PaymentSchedule = (*MonthlyPaymentSchedule)(nil)
	_ PaymentSchedule = (*BlockBasedPaymentSchedule)(nil)
	_ PaymentSchedule = (*EmptyPaymentSchedule)(nil)
)

// The PaymentSchedule interface defines the structure and behavior of different payment schedule
// types for distributing validator compensation. It provides methods to manage and track payment
// periods, ensuring rewards are distributed accurately based on the defined schedule type.
type PaymentSchedule interface {
	proto.Message

	// PeriodEnded checks whether the end of the current payment period has come. The check is made
	// against the passed context and the payment schedule's parameters and conditions.
	PeriodEnded(ctx sdktypes.Context) bool
	// TotalBlocksInPeriod returns the amount of blocks created within the current payment period.
	// The check is made against the passed context.
	TotalBlocksInPeriod(ctx sdktypes.Context) uint64
	// StartNewPeriod resets the current payment period to the start of the new one. The passed
	// context is used to define the new period's start conditions.
	StartNewPeriod(ctx sdktypes.Context)
}

// PeriodEnded checks whether the end of the current payment period has come. The current period
// ends when the month of the block creation is different from the current month of the payment
// schedule.
func (s *MonthlyPaymentSchedule) PeriodEnded(ctx sdktypes.Context) bool {
	return s.CurrentMonth != uint64(ctx.BlockTime().Month())
}

// TotalBlocksInPeriod returns the amount of blocks created from the beginning of the current month.
func (s *MonthlyPaymentSchedule) TotalBlocksInPeriod(ctx sdktypes.Context) uint64 {
	return uint64(ctx.BlockHeight()) - s.CurrentMonthStartBlock
}

// StartNewPeriod sets the current payment period to new month and block height.
func (s *MonthlyPaymentSchedule) StartNewPeriod(ctx sdktypes.Context) {
	s.CurrentMonth = uint64(ctx.BlockTime().Month())
	s.CurrentMonthStartBlock = uint64(ctx.BlockHeight())
}

// PeriodEnded checks whether the end of the current payment period has come. The current period
// ends when there has been at least BlocksPerPeriod since CurrentPeriodStartBlock.
func (s *BlockBasedPaymentSchedule) PeriodEnded(ctx sdktypes.Context) bool {
	return uint64(ctx.BlockHeight()) >= s.CurrentPeriodStartBlock+s.BlocksPerPeriod
}

// TotalBlocksInPeriod returns the amount of blocks created from the beginning of the current period.
func (s *BlockBasedPaymentSchedule) TotalBlocksInPeriod(ctx sdktypes.Context) uint64 {
	return uint64(ctx.BlockHeight()) - s.CurrentPeriodStartBlock
}

// StartNewPeriod sets the current payment period start block to the current block height.
func (s *BlockBasedPaymentSchedule) StartNewPeriod(ctx sdktypes.Context) {
	s.CurrentPeriodStartBlock = uint64(ctx.BlockHeight())
}

// PeriodEnded always returns false for the EmptyPaymentSchedule.
func (s *EmptyPaymentSchedule) PeriodEnded(_ sdktypes.Context) bool {
	return false
}

// TotalBlocksInPeriod always returns 0 for the EmptyPaymentSchedule.
func (s *EmptyPaymentSchedule) TotalBlocksInPeriod(_ sdktypes.Context) uint64 {
	return 0
}

// StartNewPeriod does nothing for the EmptyPaymentSchedule.
func (s *EmptyPaymentSchedule) StartNewPeriod(_ sdktypes.Context) {
}

// PaymentScheduleByType returns a PaymentSchedule instance that corresponds to the given
// PaymentScheduleType.
func PaymentScheduleByType(paymentScheduleType PaymentScheduleType) PaymentSchedule {
	switch paymentScheduleType {
	case PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_BLOCK_BASED:
		return &BlockBasedPaymentSchedule{}
	case PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_MONTHLY:
		return &MonthlyPaymentSchedule{}
	default:
		return &EmptyPaymentSchedule{}
	}
}

// PaymentScheduleMatchesType checks whether the given PaymentSchedule instance matches the given
// PaymentScheduleType.
func PaymentScheduleMatchesType(ps PaymentSchedule, t PaymentScheduleType) bool {
	switch ps.(type) {
	case *MonthlyPaymentSchedule:
		return t == PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_MONTHLY
	case *BlockBasedPaymentSchedule:
		return t == PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_BLOCK_BASED
	default:
		return t == PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_UNSPECIFIED
	}
}
