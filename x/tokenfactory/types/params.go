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
	KeyWhitelistedHooks        = []byte("WhitelistedHooks")
	// We don't want to charge users for denom creation
	DefaultDenomCreationFee        sdk.Coins
	DefaultDenomCreationGasConsume uint64
	DefaultFeeCollectorAddress     = ""
	DefaultWhitelistedHooks        = []*WhitelistedHook{}
)

// ParamKeyTable the param key table for tokenfactory module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(denomCreationFee sdk.Coins, denomCreationGasConsume uint64, feeCollectorAddress string, whitelistedHooks []*WhitelistedHook) Params {
	return Params{
		DenomCreationFee:        denomCreationFee,
		DenomCreationGasConsume: denomCreationGasConsume,
		FeeCollectorAddress:     feeCollectorAddress,
		WhitelistedHooks:        whitelistedHooks,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDenomCreationFee, DefaultDenomCreationGasConsume, DefaultFeeCollectorAddress, DefaultWhitelistedHooks)
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
		return fmt.Errorf("failed to validate DenomCreationFee: %w", err)
	}

	if err := validateFeeCollectorAddress(p.FeeCollectorAddress); err != nil {
		return fmt.Errorf("failed to validate FeeCollectorAddress: %w", err)
	}

	if err := validateWhitelistedHooks(p.WhitelistedHooks); err != nil {
		return fmt.Errorf("failed to validate WhitelistedHooks: %w", err)
	}

	return nil
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDenomCreationFee, &p.DenomCreationFee, validateDenomCreationFee),
		paramtypes.NewParamSetPair(KeyDenomCreationGasConsume, &p.DenomCreationGasConsume, validateDenomCreationGasConsume),
		paramtypes.NewParamSetPair(KeyFeeCollectorAddress, &p.FeeCollectorAddress, validateFeeCollectorAddress),
		paramtypes.NewParamSetPair(KeyWhitelistedHooks, &p.WhitelistedHooks, validateWhitelistedHooks),
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

func validateFeeCollectorAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(v) == 0 {
		return nil
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid fee collector address: %w", err)
	}

	return nil
}

func validateWhitelistedHooks(i interface{}) error {
	hooks, ok := i.([]*WhitelistedHook)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	seenHooks := map[string]bool{}
	for _, hook := range hooks {
		hookStr := hook.String()
		if seenHooks[hookStr] {
			return fmt.Errorf("duplicate whitelisted hook: %s", hookStr)
		}
		seenHooks[hookStr] = true
		_, err := sdk.AccAddressFromBech32(hook.DenomCreator)
		if err != nil {
			return fmt.Errorf("invalid denom creator address: %w", err)
		}
	}
	return nil
}
