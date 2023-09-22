package types

// ModuleName defines the name for the swap middleware.
const ModuleName = "swap-middleware"

// ProcessedKey is used to signal to the swap middleware that a packet has already been processed by some other
// middleware and so invoking the transfer modules OnRecvPacket callback should be avoided.
type ProcessedKey struct{}
