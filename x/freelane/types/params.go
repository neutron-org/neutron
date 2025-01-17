package types

import (
	fmt "fmt"

	"gopkg.in/yaml.v2"
)

var (
	DefaultBlockSpace     = 0.1
	DefaultSequenceNumber = uint64(2)
)

// NewParams creates a new Params instance
func NewParams(blockSpace float64, sequenceNumber uint64) Params {
	return Params{
		BlockSpace:     blockSpace,
		SequenceNumber: sequenceNumber,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultBlockSpace, DefaultSequenceNumber)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.BlockSpace < 0 || p.BlockSpace > 1 {
		return fmt.Errorf("invalid block space")
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
