package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "feerefunder"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName
)

const (
	prefixFeeKey = iota + 1

	Separator = ";"
)

var FeeKey = []byte{prefixFeeKey}

func GetFeePacketKey(packet PacketID) []byte {
	return append(append(FeeKey, []byte(packet.ChannelId+Separator+packet.PortId+Separator)...), sdk.Uint64ToBigEndian(packet.Sequence)...)
}
