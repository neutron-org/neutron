package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/harpoon module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 2, "expected gov account as only signer for proposal message")
)
