package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/neutron-org/neutron/v6/app/params"

	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyMsgSubmitTxMaxMessages     = []byte("MsgSubmitTxMaxMessages")
	DefaultMsgSubmitTxMaxMessages = uint64(16)
	DefaultRegisterFee            = sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(1_000_000)))
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable(
		paramtypes.NewParamSetPair(
			KeyMsgSubmitTxMaxMessages,
			DefaultMsgSubmitTxMaxMessages,
			validateMsgSubmitTxMaxMessages,
		),
	)
}

// NewParams creates a new Params instance
func NewParams(msgSubmitTxMaxMessages uint64, registerFee sdk.Coins) Params {
	return Params{
		MsgSubmitTxMaxMessages: msgSubmitTxMaxMessages,
		RegisterFee:            registerFee,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMsgSubmitTxMaxMessages, DefaultRegisterFee)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			KeyMsgSubmitTxMaxMessages,
			&p.MsgSubmitTxMaxMessages,
			validateMsgSubmitTxMaxMessages,
		),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return validateMsgSubmitTxMaxMessages(p.GetMsgSubmitTxMaxMessages())
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateMsgSubmitTxMaxMessages(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("MsgSubmitTxMaxMessages must be greater than zero")
	}

	return nil
}
