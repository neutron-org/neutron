package types

import (
	fmt "fmt"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	params := DefaultParams()
	return &GenesisState{
		Params:          params,
		PaymentSchedule: PaymentScheduleIByType(params.PaymentScheduleType.PaymentScheduleType).IntoPaymentSchedule(),
		Validators:      []ValidatorInfo{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	ps, err := gs.PaymentSchedule.IntoPaymentScheduleI()
	if err != nil {
		return fmt.Errorf("invalid payment schedule: %w", err)
	}

	if !ps.MatchesType(gs.Params.PaymentScheduleType.PaymentScheduleType) {
		return fmt.Errorf("payment schedule type %T does not match payment schedule of type %T in genesis state", gs.Params.PaymentScheduleType.PaymentScheduleType, ps)
	}

	return nil
}
