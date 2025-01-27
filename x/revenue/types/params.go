package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
	"gopkg.in/yaml.v2"

	"github.com/neutron-org/neutron/v5/app/params"
)

var (
	// DefaultDenomCompensation represents the default denom compensation.
	DefaultDenomCompensation = params.DefaultDenom
	// DefaultBaseCompensation represents the default compensation amount in USD.
	DefaultBaseCompensation uint64 = 2500
	// DefaultPaymentScheduleType represents the default payment schedule type.
	DefaultPaymentScheduleType = PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_UNSPECIFIED
)

// NewParams creates a new Params instance.
func NewParams(
	denomCompensation string,
	baseCompensation uint64,
	blocksPerformanceRequirement *PerformanceRequirement,
	oraclePricesPerformanceRequirement *PerformanceRequirement,
	paymentScheduleType PaymentScheduleType,
) Params {
	return Params{
		DenomCompensation:                 denomCompensation,
		BaseCompensation:                  baseCompensation,
		BlocksPerformanceRequirement:      blocksPerformanceRequirement,
		OracleVotesPerformanceRequirement: oraclePricesPerformanceRequirement,
		PaymentScheduleType:               paymentScheduleType,
	}
}

// DefaultParams returns the default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultDenomCompensation,
		DefaultBaseCompensation,
		DefaultBlocksPerformanceRequirement(),
		DefaultOracleVotesPerformanceRequirement(),
		DefaultPaymentScheduleType,
	)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate performs basic validation of the revenue module parameters.
func (p Params) Validate() error {
	if _, ex := PaymentScheduleType_name[int32(p.PaymentScheduleType)]; !ex {
		return fmt.Errorf("invalid payment schedule type %s", p.PaymentScheduleType)
	}

	if p.BlocksPerformanceRequirement.AllowedToMiss.LT(math.LegacyZeroDec()) ||
		p.BlocksPerformanceRequirement.AllowedToMiss.GT(math.LegacyOneDec()) {
		return fmt.Errorf("blocks allowed to miss must be between 0.0 and 1.0")
	}
	if p.BlocksPerformanceRequirement.RequiredAtLeast.LT(math.LegacyZeroDec()) ||
		p.BlocksPerformanceRequirement.RequiredAtLeast.GT(math.LegacyOneDec()) {
		return fmt.Errorf("blocks required at least must be between 0.0 and 1.0")
	}

	if p.OracleVotesPerformanceRequirement.AllowedToMiss.LT(math.LegacyZeroDec()) ||
		p.OracleVotesPerformanceRequirement.AllowedToMiss.GT(math.LegacyOneDec()) {
		return fmt.Errorf("oracle votes allowed to miss must be between 0.0 and 1.0")
	}
	if p.OracleVotesPerformanceRequirement.RequiredAtLeast.LT(math.LegacyZeroDec()) ||
		p.OracleVotesPerformanceRequirement.RequiredAtLeast.GT(math.LegacyOneDec()) {
		return fmt.Errorf("oracle votes required at least must be between 0.0 and 1.0")
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
