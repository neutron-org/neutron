package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"

	"github.com/neutron-org/neutron/v4/app/params"
)

// NewParams creates a new Params instance
func NewParams(prices sdk.DecCoins) Params {
	return Params{
		NtrnPrices: prices,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		sdk.NewDecCoins(sdk.NewDecCoin(params.DefaultDenom, math.OneInt())),
	)
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
