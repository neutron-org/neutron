package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (s *State) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var ps PaymentSchedule
	return unpacker.UnpackAny(s.PaymentSchedule, &ps)
}
