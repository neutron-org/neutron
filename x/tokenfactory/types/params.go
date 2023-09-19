package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyDenomCreationFee        = []byte("DenomCreationFee")
	KeyDenomCreationGasConsume = []byte("DenomCreationGasConsume")
	KeyFeeCollectorAddress     = []byte("FeeCollectorAddress")
	// We don't want to charge users for denom creation
	DefaultDenomCreationFee        sdk.Coins = nil
	DefaultDenomCreationGasConsume uint64    = 0
	DefaultFeeCollectorAddress               = ""
)

// ParamKeyTable the param key table for tokenfactory module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(denomCreationFee sdk.Coins, denomCreationGasConsume uint64, feeCollectorAddress string) Params {
	return Params{
		DenomCreationFee:        denomCreationFee,
		DenomCreationGasConsume: denomCreationGasConsume,
		FeeCollectorAddress:     feeCollectorAddress,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDenomCreationFee, DefaultDenomCreationGasConsume, DefaultFeeCollectorAddress)
}

// Validate validates params
func (p Params) Validate() error {
	// DenomCreationFee and FeeCollectorAddress must be both set, or both unset
	hasDenomCreationFee := len(p.DenomCreationFee) > 0
	hasFeeCollectorAddress := p.FeeCollectorAddress != ""

	if hasDenomCreationFee != hasFeeCollectorAddress {
		return fmt.Errorf("DenomCreationFee and FeeCollectorAddr must be both set or both unset")
	}

	if err := validateDenomCreationFee(p.DenomCreationFee); err != nil {
		return err
	}

	if err := validateAddress(p.FeeCollectorAddress); err != nil {
		return err
	}

	return nil
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDenomCreationFee, &p.DenomCreationFee, validateDenomCreationFee),
		paramtypes.NewParamSetPair(KeyDenomCreationGasConsume, &p.DenomCreationGasConsume, validateDenomCreationGasConsume),
		paramtypes.NewParamSetPair(KeyFeeCollectorAddress, &p.FeeCollectorAddress, validateAddress),
	}
}

func validateDenomCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid denom creation fee: %+v", i)
	}

	return nil
}

func validateDenomCreationGasConsume(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(v) == 0 {
		return nil
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	return nil
}
