package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"

	math_utils "github.com/neutron-org/neutron/v4/utils/math"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyFeeTiers               = []byte("FeeTiers")
	DefaultFeeTiers           = []uint64{0, 1, 2, 3, 4, 5, 10, 20, 50, 100, 150, 200}
	DefaultMaxTrueTakerSpread = math_utils.MustNewPrecDecFromStr("0.005")
	KeyPaused                 = []byte("Paused")
	DefaultPaused             = false
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(feeTiers []uint64, maxTrueTakerSpread math_utils.PrecDec, paused bool) Params {
	return Params{
		FeeTiers:           feeTiers,
		MaxTrueTakerSpread: maxTrueTakerSpread,
		Paused:             paused,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultFeeTiers, DefaultMaxTrueTakerSpread, DefaultPaused)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyFeeTiers, &p.FeeTiers, validateFeeTiers),
		paramtypes.NewParamSetPair(KeyPaused, &p.Paused, validatePaused),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
func (p Params) Validate() error {
	err := validateFeeTiers(p.FeeTiers)
	if err != nil {
		return fmt.Errorf("invalid fee tiers: %w", err)
	}

	err = validatePaused(p.Paused)
	if err != nil {
		return fmt.Errorf("invalid paused: %w", err)
	}

	return nil
}

func validateFeeTiers(v interface{}) error {
	feeTiers, ok := v.([]uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	feeTierMap := make(map[uint64]bool)
	for _, f := range feeTiers {
		if _, ok := feeTierMap[f]; ok {
			return fmt.Errorf("duplicate fee tier found")
		}
		feeTierMap[f] = true
	}
	return nil
}

func validatePaused(v interface{}) error {
	_, ok := v.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return nil
}
