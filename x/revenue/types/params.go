package types

import (
	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/v5/app/params"
	"gopkg.in/yaml.v2"
)

var (
	DefaultDenomCompensation      = params.DefaultDenom
	DefaultOracleLegacyVoteWeight = math.LegacyOneDec()
	DefaultPerformanceThreshold   = math.LegacyNewDecWithPrec(1, 1)
	DefaultAllowedMissed          = math.LegacyNewDecWithPrec(5, 3)
)

// NewParams creates a new Params instance
func NewParams(
	denomCompensation string,
	oracleVoteWeight math.LegacyDec,
	performanceThreshold math.LegacyDec,
	allowedMissed math.LegacyDec,
) Params {
	return Params{
		DenomCompensation:    denomCompensation,
		OracleVoteWeight:     oracleVoteWeight,
		PerformanceThreshold: performanceThreshold,
		AllowedMissed:        allowedMissed,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultDenomCompensation,
		DefaultOracleLegacyVoteWeight,
		DefaultPerformanceThreshold,
		DefaultAllowedMissed,
	)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
