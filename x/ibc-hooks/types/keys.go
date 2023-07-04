package types

const (
	ModuleName     = "ibchooks"
	RouteKey       = ModuleName
	StoreKey       = "hooks-for-ibc" // not using the module name because of collisions with key "ibc"
	IBCCallbackKey = "ibc_callback"
	SenderPrefix   = "ibc-wasm-hook-intermediary"
)
