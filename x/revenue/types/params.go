package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
	"gopkg.in/yaml.v2"

	"github.com/neutron-org/neutron/v6/app/params"
)

var (
	// DefaultRewardAsset represents the default reward asset.
	DefaultRewardAsset = params.DefaultDenom
	// DefaultRewardQuoteAmount represents the default reward amount measured in quote asset.
	DefaultRewardQuoteAmount uint64 = 2500
	// DefaultRewardQuoteAsset represents the default reward quote asset.
	DefaultRewardQuoteAsset = "USD"
	// DefaultTWAPWindow represents default time span to calculate TWAP for. Measured in seconds.
	DefaultTWAPWindow int64 = 24 * 3600

	// MaxTWAPWindow represents the maximum window size allowed to set with params
	// to reduce inordinate state growth
	MaxTWAPWindow int64 = 30 * 24 * 3600
)

// NewParams creates a new Params instance.
func NewParams(
	rewardAsset string,
	rewardQuote *RewardQuote,
	blocksPerformanceRequirement *PerformanceRequirement,
	oraclePricesPerformanceRequirement *PerformanceRequirement,
	paymentScheduleType isPaymentScheduleType_PaymentScheduleType,
) Params {
	return Params{
		RewardAsset:                       rewardAsset,
		RewardQuote:                       rewardQuote,
		BlocksPerformanceRequirement:      blocksPerformanceRequirement,
		OracleVotesPerformanceRequirement: oraclePricesPerformanceRequirement,
		PaymentScheduleType:               &PaymentScheduleType{PaymentScheduleType: paymentScheduleType},
		TwapWindow:                        DefaultTWAPWindow,
	}
}

// DefaultParams returns the default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultRewardAsset,
		DefaultRewardQuote(),
		DefaultBlocksPerformanceRequirement(),
		DefaultOracleVotesPerformanceRequirement(),
		DefaultPaymentScheduleType(),
	)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate performs basic validation of the revenue module parameters.
func (p Params) Validate() error {
	// shorthands
	bpr := p.BlocksPerformanceRequirement
	ovpr := p.OracleVotesPerformanceRequirement

	if p.TwapWindow < 1 || p.TwapWindow > MaxTWAPWindow {
		return fmt.Errorf("twap window must be between 1 and %d", MaxTWAPWindow)
	}

	if bpr.AllowedToMiss.LT(math.LegacyZeroDec()) || bpr.AllowedToMiss.GT(math.LegacyOneDec()) {
		return fmt.Errorf("blocks allowed to miss must be between 0.0 and 1.0")
	}
	if bpr.RequiredAtLeast.LT(math.LegacyZeroDec()) || bpr.RequiredAtLeast.GT(math.LegacyOneDec()) {
		return fmt.Errorf("blocks required at least must be between 0.0 and 1.0")
	}

	if ovpr.AllowedToMiss.LT(math.LegacyZeroDec()) || ovpr.AllowedToMiss.GT(math.LegacyOneDec()) {
		return fmt.Errorf("oracle votes allowed to miss must be between 0.0 and 1.0")
	}
	if ovpr.RequiredAtLeast.LT(math.LegacyZeroDec()) || ovpr.RequiredAtLeast.GT(math.LegacyOneDec()) {
		return fmt.Errorf("oracle votes required at least must be between 0.0 and 1.0")
	}

	if bpr.AllowedToMiss.Add(bpr.RequiredAtLeast).GT(math.LegacyOneDec()) {
		return fmt.Errorf("sum of blocks allowed to miss and required at least must not be greater than 1.0")
	}
	if ovpr.AllowedToMiss.Add(ovpr.RequiredAtLeast).GT(math.LegacyOneDec()) {
		return fmt.Errorf("sum of oracle votes allowed to miss and required at least must not be greater than 1.0")
	}

	if err := ValidatePaymentScheduleType(p.PaymentScheduleType.PaymentScheduleType); err != nil {
		return fmt.Errorf("validation of payment schedule type failed: %w", err)
	}

	return nil
}

// DefaultBlocksPerformanceRequirement returns the default blocks performance requirement.
func DefaultBlocksPerformanceRequirement() *PerformanceRequirement {
	return &PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.005
		RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 0.9
	}
}

// DefaultOracleVotesPerformanceRequirement returns the default oracle votes performance requirement.
func DefaultOracleVotesPerformanceRequirement() *PerformanceRequirement {
	return &PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.005
		RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 0.9
	}
}

// DefaultPaymentScheduleType returns the default payment schedule type.
func DefaultPaymentScheduleType() isPaymentScheduleType_PaymentScheduleType {
	return &PaymentScheduleType_EmptyPaymentScheduleType{
		EmptyPaymentScheduleType: &EmptyPaymentScheduleType{},
	}
}

// DefaultRewardQuote returns the default reward quote.
func DefaultRewardQuote() *RewardQuote {
	return &RewardQuote{
		Asset:  DefaultRewardQuoteAsset,
		Amount: DefaultRewardQuoteAmount,
	}
}
