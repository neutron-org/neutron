package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

// NewParams creates a new Params instance
func NewParams(prices sdk.DecCoins) Params {
	return Params{
		NtrnPrices: prices,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(sdk.DecCoins{})
}

// Validate validates the set of params
func (p Params) Validate() error {
	// if p.NtrnPrices.Len() ==0{
	//	return
	//}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
