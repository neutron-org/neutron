package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/dynamicfees module sentinel errors
var (
	ErrUnknownDenom = errors.Register(ModuleName, 1100, "unknown denom")
)
