package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "feerefunder"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_feerefunder"
)

const (
	prefixFeeKey = iota + 1

	Separator = ";"
)

var FeeKey = []byte{prefixFeeKey}

func GetFeePacketKey(channelID, portID string, sequenceID uint64) []byte {
	return append(append(FeeKey, []byte(channelID+Separator+portID+Separator)...), sdk.Uint64ToBigEndian(sequenceID)...)
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
