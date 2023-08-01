package types

import (
	"cosmossdk.io/errors"
)

// x/interchainqueries module sentinel errors
var (
	ErrInvalidQueryID             = errors.Register(ModuleName, 1100, "invalid query id")
	ErrEmptyResult                = errors.Register(ModuleName, 1101, "empty result")
	ErrInvalidClientID            = errors.Register(ModuleName, 1102, "invalid client id")
	ErrInvalidUpdatePeriod        = errors.Register(ModuleName, 1103, "invalid update period")
	ErrInvalidConnectionID        = errors.Register(ModuleName, 1104, "invalid connection id")
	ErrInvalidQueryType           = errors.Register(ModuleName, 1105, "invalid query type")
	ErrInvalidTransactionsFilter  = errors.Register(ModuleName, 1106, "invalid transactions filter")
	ErrInvalidSubmittedResult     = errors.Register(ModuleName, 1107, "invalid result")
	ErrProtoMarshal               = errors.Register(ModuleName, 1108, "failed to marshal protobuf bytes")
	ErrProtoUnmarshal             = errors.Register(ModuleName, 1109, "failed to unmarshal protobuf bytes")
	ErrInvalidType                = errors.Register(ModuleName, 1110, "invalid type")
	ErrInternal                   = errors.Register(ModuleName, 1111, "internal error")
	ErrInvalidProof               = errors.Register(ModuleName, 1112, "merkle proof is invalid")
	ErrInvalidHeader              = errors.Register(ModuleName, 1113, "header is invalid")
	ErrInvalidHeight              = errors.Register(ModuleName, 1114, "height is invalid")
	ErrNoQueryResult              = errors.Register(ModuleName, 1115, "no query result")
	ErrNotContract                = errors.Register(ModuleName, 1116, "not a contract")
	ErrEmptyKeys                  = errors.Register(ModuleName, 1117, "keys are empty")
	ErrEmptyKeyPath               = errors.Register(ModuleName, 1118, "key path is empty")
	ErrEmptyKeyID                 = errors.Register(ModuleName, 1119, "key id is empty")
	ErrTooManyKVQueryKeys         = errors.Register(ModuleName, 1120, "too many keys")
	ErrUnexpectedQueryTypeGenesis = errors.Register(ModuleName, 1121, "unexpected query type")
)
