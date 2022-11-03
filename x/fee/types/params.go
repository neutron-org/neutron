package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyFees     = []byte("FEES")
	DefaultFees = Fee{
		RecvFee:    sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))),
		AckFee:     sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))),
		TimeoutFee: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))),
	}
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable(paramtypes.NewParamSetPair(KeyFees, DefaultFees, validateFee))
}

// NewParams creates a new Params instance
func NewParams(minfee Fee) Params {
	return Params{MinFee: minfee}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultFees)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{paramtypes.NewParamSetPair(KeyFees, &p.MinFee, validateFee)}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateFee(i interface{}) error {
	v, ok := i.(Fee)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.RecvFee.IsZero() || v.AckFee.IsZero() || v.TimeoutFee.IsZero() {
		return fmt.Errorf("fee can't be zero: %s", v)
	}

	return nil
}
