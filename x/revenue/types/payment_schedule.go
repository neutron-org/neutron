package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

var (
	_ PaymentScheduleI = (*MonthlyPaymentSchedule)(nil)
	_ PaymentScheduleI = (*BlockBasedPaymentSchedule)(nil)
	_ PaymentScheduleI = (*EmptyPaymentSchedule)(nil)
)

var (
	// PeriodCompletenessZero represents zero payment period completeness.
	PeriodCompletenessZero math.LegacyDec = math.LegacyZeroDec()
	// PeriodCompletenessFull represents full payment period completeness.
	PeriodCompletenessFull math.LegacyDec = math.LegacyOneDec()
)

// The PaymentScheduleI interface defines the structure and behavior of different payment schedule
// types for distributing validator compensation. It provides methods to manage and track payment
// periods, ensuring rewards are distributed accurately based on the defined schedule type.
type PaymentScheduleI interface {
	proto.Message

	// PeriodCompleteness returns the completeness of the current payment period as a decimal value
	// in the range [0,1], where 0 indicates a just started period and 1 indicates a complete period.
	PeriodCompleteness(ctx sdktypes.Context) math.LegacyDec
	// TotalBlocksInPeriod returns the amount of blocks created within the current payment period.
	// The check is made against the passed context.
	TotalBlocksInPeriod(ctx sdktypes.Context) uint64
	// StartNewPeriod resets the current payment period to the start of the new one. The passed
	// context is used to define the new period's start conditions.
	StartNewPeriod(ctx sdktypes.Context)
	// MatchesType checks whether the payment schedule matches a given payment schedule type.
	MatchesType(t isPaymentScheduleType_PaymentScheduleType) bool
	// IntoPaymentSchedule creates a PaymentSchedule with a oneof value populated accordingly.
	IntoPaymentSchedule() *PaymentSchedule
}

// PeriodCompleteness returns the completeness of the current payment period as a decimal value in
// the range [0,1]. The current period ends when the month of the block creation is different from
// the current month defined in the payment schedule. Otherwise the result completeness is equal to
// the ratio of hours passed since the start of the current month to the total number of hours in
// the month.
func (s *MonthlyPaymentSchedule) PeriodCompleteness(ctx sdktypes.Context) math.LegacyDec {
	if s.CurrentMonth != uint64(ctx.BlockTime().Month()) {
		return PeriodCompletenessFull
	}

	// source: https://www.brandur.org/fragments/go-days-in-month
	daysInCurrentMonth := time.Date(
		ctx.BlockTime().Year(),
		ctx.BlockTime().Month()+1,
		0, 0, 0, 0, 0,
		ctx.BlockTime().Location(),
	).Day()
	hoursInCurrentMonth := int64(daysInCurrentMonth * 24)
	hoursPassed := int64((ctx.BlockTime().Day()-1)*24 + ctx.BlockTime().Hour())
	return math.LegacyNewDec(hoursPassed).Quo(math.LegacyNewDec(hoursInCurrentMonth))
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

// MatchesType checks whether the payment schedule matches a given payment schedule type.
func (s *MonthlyPaymentSchedule) MatchesType(t isPaymentScheduleType_PaymentScheduleType) bool {
	_, ok := t.(*PaymentScheduleType_MonthlyPaymentScheduleType)
	return ok
}

// IntoPaymentSchedule creates a PaymentSchedule with a oneof value populated accordingly.
func (s *MonthlyPaymentSchedule) IntoPaymentSchedule() *PaymentSchedule {
	return &PaymentSchedule{PaymentSchedule: &PaymentSchedule_MonthlyPaymentSchedule{MonthlyPaymentSchedule: s}}
}

// PeriodCompleteness returns the completeness of the current payment period as a decimal value in
// the range [0,1]. The current period is complete when there has been at least BlocksPerPeriod since
// CurrentPeriodStartBlock. Otherwise the result completeness is equal to the ratio of the blocks
// created during the period to the number of blocks per period defined in the schedule.
func (s *BlockBasedPaymentSchedule) PeriodCompleteness(ctx sdktypes.Context) math.LegacyDec {
	switch {
	case uint64(ctx.BlockHeight()) >= s.CurrentPeriodStartBlock+s.BlocksPerPeriod:
		return PeriodCompletenessFull
	case uint64(ctx.BlockHeight()) <= s.CurrentPeriodStartBlock:
		return PeriodCompletenessZero
	default:
		return math.LegacyNewDec(ctx.BlockHeight()).
			Sub(math.LegacyNewDecFromInt(math.NewIntFromUint64(s.CurrentPeriodStartBlock))).
			QuoInt(math.NewIntFromUint64(s.BlocksPerPeriod))
	}
}

// TotalBlocksInPeriod returns the amount of blocks created from the beginning of the current period.
func (s *BlockBasedPaymentSchedule) TotalBlocksInPeriod(ctx sdktypes.Context) uint64 {
	return uint64(ctx.BlockHeight()) - s.CurrentPeriodStartBlock
}

// StartNewPeriod sets the current payment period start block to the current block height.
func (s *BlockBasedPaymentSchedule) StartNewPeriod(ctx sdktypes.Context) {
	s.CurrentPeriodStartBlock = uint64(ctx.BlockHeight())
}

// MatchesType checks whether the payment schedule matches a given payment schedule type.
func (s *BlockBasedPaymentSchedule) MatchesType(t isPaymentScheduleType_PaymentScheduleType) bool {
	v, ok := t.(*PaymentScheduleType_BlockBasedPaymentScheduleType)
	return ok &&
		v.BlockBasedPaymentScheduleType != nil &&
		s.BlocksPerPeriod == v.BlockBasedPaymentScheduleType.BlocksPerPeriod
}

// IntoPaymentSchedule creates a PaymentSchedule with a oneof value populated accordingly.
func (s *BlockBasedPaymentSchedule) IntoPaymentSchedule() *PaymentSchedule {
	return &PaymentSchedule{PaymentSchedule: &PaymentSchedule_BlockBasedPaymentSchedule{BlockBasedPaymentSchedule: s}}
}

// PeriodCompleteness always returns zero completeness for the EmptyPaymentSchedule.
func (s *EmptyPaymentSchedule) PeriodCompleteness(ctx sdktypes.Context) math.LegacyDec {
	return PeriodCompletenessZero
}

// TotalBlocksInPeriod always returns 0 for the EmptyPaymentSchedule.
func (s *EmptyPaymentSchedule) TotalBlocksInPeriod(_ sdktypes.Context) uint64 {
	return 0
}

// StartNewPeriod does nothing for the EmptyPaymentSchedule.
func (s *EmptyPaymentSchedule) StartNewPeriod(_ sdktypes.Context) {
}

// MatchesType checks whether the payment schedule matches a given payment schedule type.
func (s *EmptyPaymentSchedule) MatchesType(t isPaymentScheduleType_PaymentScheduleType) bool {
	_, ok := t.(*PaymentScheduleType_EmptyPaymentScheduleType)
	return ok
}

// IntoPaymentSchedule creates a PaymentSchedule with a oneof value populated accordingly.
func (s *EmptyPaymentSchedule) IntoPaymentSchedule() *PaymentSchedule {
	return &PaymentSchedule{PaymentSchedule: &PaymentSchedule_EmptyPaymentSchedule{EmptyPaymentSchedule: s}}
}

// PaymentScheduleIByType returns a PaymentScheduleI that corresponds to the given
// PaymentScheduleType.
func PaymentScheduleIByType(paymentScheduleType isPaymentScheduleType_PaymentScheduleType) PaymentScheduleI {
	switch v := paymentScheduleType.(type) {
	case *PaymentScheduleType_BlockBasedPaymentScheduleType:
		return &BlockBasedPaymentSchedule{BlocksPerPeriod: v.BlockBasedPaymentScheduleType.BlocksPerPeriod}
	case *PaymentScheduleType_MonthlyPaymentScheduleType:
		return &MonthlyPaymentSchedule{}
	case *PaymentScheduleType_EmptyPaymentScheduleType:
		return &EmptyPaymentSchedule{}
	default:
		panic(fmt.Sprintf("invalid payment schedule type: %T", paymentScheduleType))
	}
}

// ValidatePaymentScheduleType checks whether a given payment schedule type implementation is
// properly initialized.
func ValidatePaymentScheduleType(paymentScheduleType isPaymentScheduleType_PaymentScheduleType) error {
	switch v := paymentScheduleType.(type) {
	case *PaymentScheduleType_BlockBasedPaymentScheduleType:
		if v.BlockBasedPaymentScheduleType == nil {
			return fmt.Errorf("inner block based payment schedule is nil")
		}
		if v.BlockBasedPaymentScheduleType.BlocksPerPeriod == 0 {
			return fmt.Errorf("block based payment schedule type has zero blocks per period")
		}
		return nil
	case *PaymentScheduleType_MonthlyPaymentScheduleType:
		if v.MonthlyPaymentScheduleType == nil {
			return fmt.Errorf("inner monthly payment schedule is nil")
		}
		return nil
	case *PaymentScheduleType_EmptyPaymentScheduleType:
		if v.EmptyPaymentScheduleType == nil {
			return fmt.Errorf("inner empty payment schedule is nil")
		}
		return nil
	default:
		panic(fmt.Sprintf("invalid payment schedule type: %T", paymentScheduleType))
	}
}

// IntoPaymentScheduleI returns the oneof value populated in a given PaymentSchedule as a
// PaymentScheduleI.
func (s *PaymentSchedule) IntoPaymentScheduleI() (PaymentScheduleI, error) {
	switch v := s.PaymentSchedule.(type) {
	case *PaymentSchedule_BlockBasedPaymentSchedule:
		return v.BlockBasedPaymentSchedule, nil
	case *PaymentSchedule_MonthlyPaymentSchedule:
		return v.MonthlyPaymentSchedule, nil
	case *PaymentSchedule_EmptyPaymentSchedule:
		return v.EmptyPaymentSchedule, nil
	default:
		return nil, fmt.Errorf("no set oneof field found in payment schedule %+v", s)
	}
}
