package types

import (
	"cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "revenue"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	RevenueFeeRedistributePoolName = "revenue-fee-redistribute"
	RevenueTreasuryPoolName        = "revenue-treasury"
	RevenueStakingRewardsPoolName  = "revenue-staking-rewards"
)

var (
	ParamsKey                 = []byte{0x01}
	StateKey                  = []byte{0x02}
	PrefixValidatorInfoKey    = []byte{0x03}
	PrefixAccumulatedPriceKey = []byte{0x04}
)

func GetValidatorInfoKey(addr sdk.ConsAddress) []byte {
	return append(PrefixValidatorInfoKey, addr.Bytes()...)
}

func GetAccumulatedPriceKey(time uint64) []byte {
	return append(PrefixAccumulatedPriceKey, types.Uint64ToBigEndian(time)...)
}

func GetTimeFromAccumulatedPriceKey(key []byte) uint64 {
	return types.BigEndianToUint64(key[1:])
}
