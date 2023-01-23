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
	KeyQuerySubmitTimeout               = []byte("QuerySubmitTimeout")
	DefaultQuerySubmitTimeout           = uint64(1036800) // One month, with block_time = 2.5s
	KeyQueryDeposit                     = []byte("QueryDeposit")
	DefaultQueryDeposit       sdk.Coins = sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdk.NewInt(int64(1_000_000))))
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable(
		paramtypes.NewParamSetPair(KeyQuerySubmitTimeout, DefaultQuerySubmitTimeout, func(value interface{}) error { return nil }),
		paramtypes.NewParamSetPair(KeyQueryDeposit, sdk.Coins{}, validateCoins),
	)
}

// NewParams creates a new Params instance
func NewParams(querySubmitTimeout uint64, queryDeposit sdk.Coins) Params {
	return Params{
		QuerySubmitTimeout: querySubmitTimeout,
		QueryDeposit:       queryDeposit,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultQuerySubmitTimeout, DefaultQueryDeposit)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyQuerySubmitTimeout, &p.QuerySubmitTimeout, func(value interface{}) error { return nil }),
		paramtypes.NewParamSetPair(KeyQueryDeposit, &p.QueryDeposit, validateCoins),
	}
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

func validateCoins(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !v.IsValid() {
		return fmt.Errorf("invalid coins parameter: %s", v)
	}

	return nil
}
