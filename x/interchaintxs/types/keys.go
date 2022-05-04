package types

const (
	// ModuleName defines the module name
	ModuleName = "interchaintxs"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_interchaintxs"
)

// Prefix bytes for the epoch persistent store
const (
	PrefixHubAddress = "prefix_hub_address"
)

var (
	KeyHubAddress = []byte("hub_address")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
