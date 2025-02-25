package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyFeeTiers                         = []byte("FeeTiers")
	DefaultFeeTiers                     = []uint64{0, 1, 2, 3, 4, 5, 10, 20, 50, 100, 150, 200}
	KeyPaused                           = []byte("Paused")
	DefaultPaused                       = false
	KeyMaxJITsPerBlock                  = []byte("MaxJITs")
	DefaultMaxJITsPerBlock       uint64 = 25
	KeyGoodTilPurgeAllowance            = []byte("PurgeAllowance")
	DefaultGoodTilPurgeAllowance uint64 = 540_000
	KeyWhitelistedLPs                   = []byte("WhiteListedLPs")
	DefaultKeyWhitelistedLPs     []string
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(feeTiers []uint64, paused bool, maxJITsPerBlock, goodTilPurgeAllowance uint64, whitelistedLPs []string) Params {
	return Params{
		FeeTiers:              feeTiers,
		Paused:                paused,
		MaxJitsPerBlock:       maxJITsPerBlock,
		GoodTilPurgeAllowance: goodTilPurgeAllowance,
		WhitelistedLps:        whitelistedLPs,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultFeeTiers, DefaultPaused, DefaultMaxJITsPerBlock, DefaultGoodTilPurgeAllowance, DefaultKeyWhitelistedLPs)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyFeeTiers, &p.FeeTiers, validateFeeTiers),
		paramtypes.NewParamSetPair(KeyPaused, &p.Paused, validatePaused),
		paramtypes.NewParamSetPair(KeyMaxJITsPerBlock, &p.MaxJitsPerBlock, validateMaxJITsPerBlock),
		paramtypes.NewParamSetPair(KeyGoodTilPurgeAllowance, &p.GoodTilPurgeAllowance, validatePurgeAllowance),
		paramtypes.NewParamSetPair(KeyWhitelistedLPs, &p.WhitelistedLps, validateWhitelistedLPs),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateFeeTiers(p.FeeTiers); err != nil {
		return fmt.Errorf("invalid fee tiers: %w", err)
	}

	if err := validatePaused(p.Paused); err != nil {
		return fmt.Errorf("invalid paused: %w", err)
	}

	if err := validateMaxJITsPerBlock(p.MaxJitsPerBlock); err != nil {
		return fmt.Errorf("invalid max JITs per block: %w", err)
	}

	if err := validatePurgeAllowance(p.GoodTilPurgeAllowance); err != nil {
		return fmt.Errorf("invalid good til purge allowance: %w", err)
	}

	if err := validateWhitelistedLPs(p.WhitelistedLps); err != nil {
		return fmt.Errorf("invalid whitelisted LPs: %w", err)
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

func validateMaxJITsPerBlock(v interface{}) error {
	_, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return nil
}

func validatePurgeAllowance(v interface{}) error {
	_, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return nil
}

func validateWhitelistedLPs(v interface{}) error {
	whitelistedLPs, ok := v.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	for _, addr := range whitelistedLPs {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return fmt.Errorf("invalid LP address (%s): %w, ", addr, err)
		}
	}

	return nil
}
