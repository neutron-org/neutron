package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/interchainqueries module sentinel errors
var (
	ErrSample = sdkerrors.Register(ModuleName, 1101, "sample error")
)
