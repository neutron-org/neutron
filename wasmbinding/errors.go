package wasmbinding

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/dex module sentinel errors

var ErrModuleNotSupported = sdkerrors.Register(
	"wasmbinding",
	100,
	"module is no longer supported",
)
