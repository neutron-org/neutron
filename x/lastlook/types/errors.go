package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/lastlook module sentinel errors
var (
	ErrProtoMarshal   = errors.Register(ModuleName, 1201, "failed to marshal protobuf bytes")
	ErrProtoUnmarshal = errors.Register(ModuleName, 1202, "failed to unmarshal protobuf bytes")
	ErrNoBlob         = errors.Register(ModuleName, 1203, "no blob found for a block")
)
