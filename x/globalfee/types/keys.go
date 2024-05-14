package types

var ParamsKey = []byte{0x00}

const (
	// ModuleName is the name of the this module
	ModuleName = "globalfee"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	QuerierRoute = ModuleName
)
