package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
	ParamsKey              = []byte{0x01}
	StateKey               = []byte{0x02}
	PrefixValidatorInfoKey = []byte{0x03}
)

func GetValidatorInfoKey(addr sdk.ConsAddress) []byte {
	return append(PrefixValidatorInfoKey, addr.Bytes()...)
}
