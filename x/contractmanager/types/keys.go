package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "contractmanager"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_contractmanager"
)

const (
	prefixContractFailures = iota + 1
	prefixNextFailureId
)

var (
	ContractFailuresKey = []byte{prefixContractFailures}
	NextFailureIdKey    = []byte{prefixNextFailureId}
)

// GetFailureKeyPrefix returns the store key for the failures of the specific address
func GetFailureKeyPrefix(
	address string,
) []byte {
	key := append(ContractFailuresKey, []byte(address)...)
	return append(key, []byte("/")...)
}

// GetFailureKey returns the store key to retrieve a Failure from the index fields
func GetFailureKey(
	address string,
	offset uint64,
) []byte {
	key := GetFailureKeyPrefix(address)
	return append(key, sdk.Uint64ToBigEndian(offset)...)
}

// GetFailureKeyPrefix returns the store key for the failures of the specific address
func GetNextFailureIdKey(
	address string,
) []byte {
	return append(NextFailureIdKey, []byte(address)...)
}
