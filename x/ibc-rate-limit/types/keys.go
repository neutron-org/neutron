package types

import "strings"

const (
	prefixParamsKey = iota + 1
)

const (
	ModuleName = "rate-limited-ibc" // IBC at the end to avoid conflicts with the ibc prefix

)

var ParamsKey = []byte{prefixParamsKey}

// RouterKey is the message route. Can only contain
// alphanumeric characters.
var RouterKey = strings.ReplaceAll(ModuleName, "-", "")
