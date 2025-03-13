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
	PaymentScheduleKey        = []byte{0x02}
	PrefixValidatorInfoKey    = []byte{0x03}
	PrefixAccumulatedPriceKey = []byte{0x04}
)

func GetValidatorInfoKey(addr sdk.ValAddress) []byte {
	return append(PrefixValidatorInfoKey, addr.Bytes()...)
}

func GetAccumulatedPriceKey(time int64) []byte {
	return append(PrefixAccumulatedPriceKey, types.Uint64ToBigEndian(uint64(time))...) //nolint:gosec
}
