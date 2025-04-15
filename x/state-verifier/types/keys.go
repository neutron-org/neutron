package types

import (
	"github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "state-verifier"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

const (
	prefixConsensusStateKey = iota + 1
)

var ConsensusStateKey = []byte{prefixConsensusStateKey}

func GetConsensusStateKey(height int64) []byte {
	return append(ConsensusStateKey, types.Uint64ToBigEndian(uint64(height))...) //nolint:gosec
}
