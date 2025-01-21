package types

const (
	// ModuleName defines the module name
	ModuleName = "harpoon"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

var (
	HookSubscriptionKey = []byte("subscriptions")
)

func GetHookSubscriptionKeyPrefix() []byte {
	return HookSubscriptionKey
}

func GetHookSubscriptionKey(hookType string) []byte {
	return append(GetHookSubscriptionKeyPrefix(), []byte(hookType)...)
}
