package types

import (
	fmt "fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	params := DefaultParams()
	paymentSchedule, err := codectypes.NewAnyWithValue(PaymentScheduleByType(params.PaymentScheduleType))
	if err != nil {
		panic(fmt.Sprintf("failed to create Any payment schedule for the default payment schedule type %s: %v", params.PaymentScheduleType, err))
	}

	return &GenesisState{
		Params: params,
		State: State{
			PaymentSchedule: paymentSchedule,
		},
		Validators: nil,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	ps, ok := gs.State.PaymentSchedule.GetCachedValue().(PaymentSchedule)
	if !ok {
		return fmt.Errorf("expected State.PaymentSchedule to be of type PaymentSchedule: %T", gs.State.PaymentSchedule.GetCachedValue())
	}
	if !PaymentScheduleMatchesType(ps, gs.Params.PaymentScheduleType) {
		return fmt.Errorf("payment schedule type %s does not match payment schedule of type %T in genesis state", gs.Params.PaymentScheduleType, ps)
	}

	return nil
}
