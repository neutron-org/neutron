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

const (
	// parameters key
	prefixParamsKey = iota + 1
	// prefix of code id, starting from which we charge fee for ICA registration
	prefixICARegistrationFeeFirstCodeID = iota + 2
)

var (
	ParamsKey                     = []byte{prefixParamsKey}
	ICARegistrationFeeFirstCodeID = []byte{prefixICARegistrationFeeFirstCodeID}
)
