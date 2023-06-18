package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyMsgSubmitTxMaxMessages     = []byte("MsgSubmitTxMaxMessages")
	DefaultMsgSubmitTxMaxMessages = uint64(16)
)

// ParamKeyTable the param key table for launch module
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
func NewParams(msgSubmitTxMaxMessages uint64) Params {
	return Params{
		MsgSubmitTxMaxMessages: msgSubmitTxMaxMessages,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMsgSubmitTxMaxMessages)
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
