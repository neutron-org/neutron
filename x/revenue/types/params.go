package types

import (
	"cosmossdk.io/math"
	"github.com/neutron-org/neutron/v5/app/params"
	"gopkg.in/yaml.v2"
)

// TODO: comments for default values
var (
	DefaultDenomCompensation           = params.DefaultDenom
	DefaultBaseCompensation     uint64 = 1500
	DefaultPerformanceThreshold        = math.LegacyNewDecWithPrec(1, 1)
	DefaultAllowedMissed               = math.LegacyNewDecWithPrec(5, 3)
)

// NewParams creates a new Params instance
func NewParams(
	denomCompensation string,
	baseCompensation uint64,
	performanceThreshold math.LegacyDec,
	allowedMissed math.LegacyDec,
) Params {
	return Params{
		DenomCompensation:    denomCompensation,
		BaseCompensation:     baseCompensation,
		PerformanceThreshold: performanceThreshold,
		AllowedMissed:        allowedMissed,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultDenomCompensation,
		DefaultBaseCompensation,
		DefaultPerformanceThreshold,
		DefaultAllowedMissed,
	)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
