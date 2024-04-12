package types

import (
	"cosmossdk.io/errors"
)

// x/interchaintxs module sentinel errors
var (
	ErrInvalidICAOwner           = errors.Register(ModuleName, 1100, "invalid interchain account interchainAccountID")
	ErrInvalidAccountAddress     = errors.Register(ModuleName, 1101, "invalid account address")
	ErrInterchainAccountNotFound = errors.Register(ModuleName, 1102, "interchain account not found")
	ErrNotContract               = errors.Register(ModuleName, 1103, "not a contract")
	ErrEmptyInterchainAccountID  = errors.Register(ModuleName, 1104, "empty interchain account id")
	ErrEmptyConnectionID         = errors.Register(ModuleName, 1105, "empty connection id")
	ErrNoMessages                = errors.Register(ModuleName, 1106, "no messages provided")
	ErrInvalidTimeout            = errors.Register(ModuleName, 1107, "invalid timeout")
	ErrInvalidPayerFee           = errors.Register(ModuleName, 1108, "invalid payer feerefunder")
	ErrLongInterchainAccountID   = errors.Register(ModuleName, 1109, "interchain account id is too long")
	ErrInvalidType               = errors.Register(ModuleName, 1110, "invalid type")
)
