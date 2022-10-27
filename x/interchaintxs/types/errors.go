package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/interchaintxs module sentinel errors
var (
	ErrInvalidICAOwner           = sdkerrors.Register(ModuleName, 1100, "invalid interchain account interchainAccountID")
	ErrInvalidAccountAddress     = sdkerrors.Register(ModuleName, 1101, "invalid account address")
	ErrInterchainAccountNotFound = sdkerrors.Register(ModuleName, 1102, "interchain account not found")
	ErrNotContract               = sdkerrors.Register(ModuleName, 1103, "not a contract")
	ErrEmptyInterchainAccountID  = sdkerrors.Register(ModuleName, 1104, "empty interchain account id")
	ErrEmptyConnectionID         = sdkerrors.Register(ModuleName, 1105, "empty connection id")
	ErrNoMessages                = sdkerrors.Register(ModuleName, 1106, "no messages provided")
	ErrInvalidTimeout            = sdkerrors.Register(ModuleName, 1107, "invalid timeout")
	ErrInvalidPayerFee           = sdkerrors.Register(ModuleName, 1108, "invalid payer fee")
)
