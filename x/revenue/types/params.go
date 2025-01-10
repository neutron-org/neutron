package types

import (
	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/v5/app/params"
	"gopkg.in/yaml.v2"
)

// TODO: comments for default values
var (
	DefaultDenomCompensation                   = params.DefaultDenom
	DefaultBaseCompensation             uint64 = 1500
	DefaultBlocksPerformanceRequirement        = &PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.005
		RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 0.9
	}
	DefaultOracleVotesPerformanceRequirement = &PerformanceRequirement{
		AllowedToMiss:   math.LegacyNewDecWithPrec(5, 3), // 0.005
		RequiredAtLeast: math.LegacyNewDecWithPrec(9, 1), // 0.9
	}
)

// NewParams creates a new Params instance
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

// DefaultParams returns a default set of parameters
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
