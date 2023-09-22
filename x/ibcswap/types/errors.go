package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrInvalidSwapMetadata = sdkerrors.Register(ModuleName, 2, "invalid swap metadata")
	ErrSwapFailed          = sdkerrors.Register(ModuleName, 3, "ibc swap failed")
	ErrMsgHandlerInvalid   = sdkerrors.Register(ModuleName, 4, "msg service handler not found")
)
