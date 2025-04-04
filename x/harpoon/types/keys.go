package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "harpoon"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

var HookSubscriptionKey = []byte("subscriptions")

// GetHookSubscriptionKeyPrefix returns the store key for hook subscriptions.
func GetHookSubscriptionKeyPrefix() []byte {
	return HookSubscriptionKey
}

// GetHookSubscriptionKey returns the store key for a specific hook subscription.
func GetHookSubscriptionKey(hookType HookType) []byte {
	arr := make([]byte, 4)
	binary.BigEndian.PutUint32(arr[0:4], uint32(hookType)) //nolint:gosec
	return append(GetHookSubscriptionKeyPrefix(), arr...)
}
