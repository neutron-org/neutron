package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyAdminAddress    = []byte("AdminAddress")
	KeySecurityAddress = []byte("SecurityAddress")
	KeyLimit           = []byte("Limit")

	DefaultAdminAddress    = ""
	DefaultSecurityAddress = ""
	DefaultLimit           = uint64(5)
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(adminAddress, securityAddress string, limit uint64) Params {
	return Params{
		AdminAddress:    adminAddress,
		SecurityAddress: securityAddress,
		Limit:           limit,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultAdminAddress, DefaultSecurityAddress, DefaultLimit)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			KeyAdminAddress,
			&p.AdminAddress,
			validateAddress,
		),
		paramtypes.NewParamSetPair(
			KeySecurityAddress,
			&p.SecurityAddress,
			validateAddress,
		),
		paramtypes.NewParamSetPair(
			KeyLimit,
			&p.Limit,
			validateLimit,
		),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	err := validateAddress(p.AdminAddress)
	if err != nil {
		return err
	}

	err = validateAddress(p.SecurityAddress)
	if err != nil {
		return err
	}

	err = validateLimit(p.Limit)
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

func validateAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// treasury might be explicitly empty in test environments
	if len(v) == 0 {
		return nil
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	return nil
}

func validateLimit(i interface{}) error {
	l, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if l == 0 {
		return fmt.Errorf("limit cannot be zero")
	}

	return nil
}
