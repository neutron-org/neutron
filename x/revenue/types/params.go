package types

import (
	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/v5/app/params"
	"gopkg.in/yaml.v2"
)

var (
	// DefaultDenomCompensation represents the default denom compensation.
	DefaultDenomCompensation = params.DefaultDenom
	// DefaultBaseCompensation represents the default compensation amount in USD.
	DefaultBaseCompensation uint64 = 2500
	// DefaultBlocksPerformanceRequirement represents the default blocks performance requirement.
	DefaultBlocksPerformanceRequirement = &PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.005
		RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 0.9
	}
	// DefaultOracleVotesPerformanceRequirement represents the default oracle votes performance
	// requirement.
	DefaultOracleVotesPerformanceRequirement = &PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.005
		RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 0.9
	}
)

// NewParams creates a new Params instance.
func NewParams(
	denomCompensation string,
	baseCompensation uint64,
	blocksPerformanceRequirement *PerformanceRequirement,
	oraclePricesPerformanceRequirement *PerformanceRequirement,
) Params {
	return Params{
		DenomCompensation:                 denomCompensation,
		BaseCompensation:                  baseCompensation,
		BlocksPerformanceRequirement:      blocksPerformanceRequirement,
		OracleVotesPerformanceRequirement: oraclePricesPerformanceRequirement,
	}
}

// DefaultParams returns the default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultDenomCompensation,
		DefaultBaseCompensation,
		DefaultBlocksPerformanceRequirement,
		DefaultOracleVotesPerformanceRequirement,
	)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
