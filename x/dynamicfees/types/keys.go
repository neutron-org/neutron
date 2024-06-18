package types

const (
	// ModuleName defines the module name
	ModuleName = "dynamicfees"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

const (
	prefixParamsKey = iota + 1
)

var ParamsKey = []byte{prefixParamsKey}
