package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "lastlook"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

const (
	prefixParamsKey   = iota + 1
	prefixTxsQueueKey = iota + prefixParamsKey
)

var (
	ParamsKey   = []byte{prefixParamsKey}
	TxsQueueKey = []byte{prefixTxsQueueKey}
)

func GetTxsQueueKey(blockHeight int64) []byte {
	return append(TxsQueueKey, sdk.Uint64ToBigEndian(uint64(blockHeight))...)
}
