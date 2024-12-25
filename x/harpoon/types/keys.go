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
	ParamsKey           = []byte("harpoon_params")
	HookSubscriptionKey = []byte("harpoon_subscriptions")
)

func GetHookSubscriptionKeyPrefix() []byte {
	return HookSubscriptionKey
}

// GetHookSubscriptionKey returns the store key for the hooks for the contractAddress
func GetHookSubscriptionKey(
	contractAddress string,
) []byte {
	return append(GetHookSubscriptionKeyPrefix(), []byte(contractAddress)...)
}
