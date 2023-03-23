package types

const (
	// ModuleName defines the module name
	ModuleName = "cron"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_cron"
)

const (
	prefixFeeKey = iota + 1
)

var FeeKey = []byte{prefixFeeKey}

func GetScheduleKey(name string) []byte {
	return append(FeeKey, []byte(name)...)
}
