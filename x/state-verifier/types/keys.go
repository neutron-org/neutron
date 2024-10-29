package types

import "strconv"

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
	return append(ConsensusStateKey, []byte(strconv.FormatInt(height, 10))...)
}

func ExtractHeightFromConsensusStateKey(key []byte) (int64, error) {
	return strconv.ParseInt(string(key[len(ConsensusStateKey):]), 10, 64)
}
