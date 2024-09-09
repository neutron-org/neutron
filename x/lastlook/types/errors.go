package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/lastlook module sentinel errors
var (
	ErrProtoMarshal    = errors.Register(ModuleName, 1201, "failed to marshal protobuf bytes")
	ErrProtoUnmarshal  = errors.Register(ModuleName, 1202, "failed to unmarshal protobuf bytes")
	ErrNoBatch         = errors.Register(ModuleName, 1203, "no batch found for a block")
	ErrInvalidProposal = errors.New(ModuleName, 1204, "invalid proposal")
)
