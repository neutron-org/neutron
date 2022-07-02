package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/interchaintxs module sentinel errors
var (
	ErrSample          = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidICAOwner = sdkerrors.Register(ModuleName, 1101, "invalid interchain account owner")
)
