package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"

	"github.com/neutron-org/neutron/v7/app/params"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyFees       = []byte("FEES")
	KeyFeeEnabled = []byte("FEEENABLED")
	DefaultFees   = Fee{
		RecvFee:    nil,
		AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(1000))),
		TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(1000))),
	}
	DefaultFeeEnabled = true
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable(
		paramtypes.NewParamSetPair(KeyFees, DefaultFees, validateFee),
		paramtypes.NewParamSetPair(KeyFeeEnabled, DefaultFeeEnabled, validateFeeEnabled))
}

// NewParams creates a new Params instance
func NewParams(minFee Fee, feeEnabled bool) Params {
	return Params{MinFee: minFee, FeeEnabled: feeEnabled}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultFees, DefaultFeeEnabled)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyFees, &p.MinFee, validateFee),
		paramtypes.NewParamSetPair(KeyFeeEnabled, &p.FeeEnabled, validateFeeEnabled),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateFee(p.MinFee); err != nil {
		return fmt.Errorf("invalid minFee: %w", err)
	}

	if err := validateFeeEnabled(p.FeeEnabled); err != nil {
		return fmt.Errorf("invalid feeEnabled: %w", err)
	}

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

	return v.Validate()
}

func validateFeeEnabled(v interface{}) error {
	_, ok := v.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return nil
}
