package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/neutron-org/neutron/app/params"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyNeutronDenom       = []byte("NeutronDenom")
	DefaultNeutronDenom   = params.DefaultDenom
	KeyReserveAddress     = []byte("ReserveAddress")
	DefaultReserveAddress = ""
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable(
		paramtypes.NewParamSetPair(
			KeyNeutronDenom,
			DefaultNeutronDenom,
			validateNeutronDenom,
		),
		paramtypes.NewParamSetPair(
			KeyReserveAddress,
			DefaultReserveAddress,
			validateReserveAddress,
		),
	)
}

// NewParams creates a new Params instance
func NewParams(neutronDenom, reserveAddress string) Params {
	return Params{
		NeutronDenom:   neutronDenom,
		ReserveAddress: reserveAddress,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultNeutronDenom, DefaultReserveAddress)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			KeyNeutronDenom,
			&p.NeutronDenom,
			validateNeutronDenom,
		),
		paramtypes.NewParamSetPair(
			KeyReserveAddress,
			&p.ReserveAddress,
			validateReserveAddress,
		),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	err := validateNeutronDenom(p.NeutronDenom)
	if err != nil {
		return err
	}

	err = validateReserveAddress(p.ReserveAddress)
	if err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateNeutronDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == "" {
		return fmt.Errorf("NeutronDenom must not be empty")
	}

	return nil
}

func validateReserveAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Reserve might be explicitly empty in test environments
	if len(v) == 0 {
		return nil
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid Reserve address: %w", err)
	}

	return nil
}
