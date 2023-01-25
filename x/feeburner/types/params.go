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
	KeyNeutronDenom        = []byte("NeutronDenom")
	DefaultNeutronDenom    = params.DefaultDenom
	KeyTreasuryAddress     = []byte("TreasuryAddress")
	DefaultTreasuryAddress = "neutron1vguuxez2h5ekltfj9gjd62fs5k4rl2zy5hfrncasykzw08rezpfsd2rhm7"
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
			KeyTreasuryAddress,
			DefaultTreasuryAddress,
			validateTreasuryAddress,
		),
	)
}

// NewParams creates a new Params instance
func NewParams(neutronDenom, treasuryAddress string) Params {
	return Params{
		NeutronDenom:    neutronDenom,
		TreasuryAddress: treasuryAddress,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultNeutronDenom, DefaultTreasuryAddress)
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
			KeyTreasuryAddress,
			&p.TreasuryAddress,
			validateTreasuryAddress,
		),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	err := validateNeutronDenom(p.NeutronDenom)
	if err != nil {
		return err
	}

	err = validateTreasuryAddress(p.TreasuryAddress)
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

func validateTreasuryAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid treasury address: %w", err)
	}

	return nil
}
