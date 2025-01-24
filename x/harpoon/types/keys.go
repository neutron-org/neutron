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

func GetHookSubscriptionKeyPrefix() []byte {
	return HookSubscriptionKey
}

func GetHookSubscriptionKey(hookType HookType) []byte {
	var arr []byte = make([]byte, 4)
	binary.BigEndian.PutUint32(arr[0:4], uint32(hookType))
	return append(GetHookSubscriptionKeyPrefix(), arr...)
}
