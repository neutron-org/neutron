package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/neutron-org/neutron/app/params"
)

// Parameter store keys.
var (
	KeyDenomCreationFee              = []byte("DenomCreationFee")
	DefaultNeutronDenom              = params.DefaultDenom
	DefaultFeeAmount           int64 = 1_000_000
	KeyFeeCollectorAddress           = []byte("FeeCollectorAddress")
	DefaultFeeCollectorAddress       = ""
)

// ParamTable for tokenfactory module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(denomCreationFee sdk.Coins, feeCollectorAddress string) Params {
	return Params{
		DenomCreationFee:    denomCreationFee,
		FeeCollectorAddress: feeCollectorAddress,
	}
}

// default tokenfactory module parameters.
func DefaultParams() Params {
	return Params{
		DenomCreationFee:    sdk.NewCoins(sdk.NewInt64Coin(DefaultNeutronDenom, DefaultFeeAmount)),
		FeeCollectorAddress: DefaultFeeCollectorAddress,
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validateDenomCreationFee(p.DenomCreationFee); err != nil {
		return err
	}

	return validateFeeCollectorAddress(p.FeeCollectorAddress)
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDenomCreationFee, &p.DenomCreationFee, validateDenomCreationFee),
		paramtypes.NewParamSetPair(KeyFeeCollectorAddress, &p.FeeCollectorAddress, validateFeeCollectorAddress),
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

func validateFeeCollectorAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Fee collector address might be explicitly empty in test environments
	if len(v) == 0 {
		return nil
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid fee collector address: %w", err)
	}

	return nil
}
