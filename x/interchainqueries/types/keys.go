package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "interchainqueries"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_interchainadapter"
)

const (
	prefixRegisteredQuery = iota + 1
	prefixRegisteredQueryResult
)

var (
	RegisteredQueryKey       = []byte{prefixRegisteredQuery}
	RegisteredQueryResultKey = []byte{prefixRegisteredQueryResult}

	LastRegisteredQueryIdKey = []byte{0x64}
)

func GetRegisteredQueryByIDKey(id uint64) []byte {
	return append(RegisteredQueryKey, sdk.Uint64ToBigEndian(id)...)
}

func GetRegisteredQueryResultByIDKey(id uint64) []byte {
	return append(RegisteredQueryResultKey, sdk.Uint64ToBigEndian(id)...)
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
