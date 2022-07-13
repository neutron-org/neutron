package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/interchaintxs module sentinel errors
var (
	ErrInvalidICAOwner       = sdkerrors.Register(ModuleName, 1100, "invalid interchain account interchainAccountID")
	ErrInvalidAccountAddress = sdkerrors.Register(ModuleName, 1101, "invalid account address")
)
