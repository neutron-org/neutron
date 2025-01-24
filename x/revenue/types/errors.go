package types

import (
	"cosmossdk.io/errors"
)

// x/revenue module sentinel errors
var (
	// ErrNoValidatorInfoFound error if there is no validator info found given a validator address.
	ErrNoValidatorInfoFound = errors.Register(ModuleName, 2, "validator info not found")
)
