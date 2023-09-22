package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/incentives module sentinel errors.
var (
	ErrNotStakeOwner = sdkerrors.Register(
		ModuleName,
		1,
		"msg sender is not the owner of specified stake",
	)
	ErrStakeNotFound = sdkerrors.Register(ModuleName, 2, "stake not found")
	ErrGaugeNotActive  = sdkerrors.Register(
		ModuleName,
		3,
		"cannot distribute from gauges when it is not active",
	)
	ErrInvalidGaugeStatus = sdkerrors.Register(
		ModuleName,
		4,
		"Gauge status filter must be one of: ACTIVE_UPCOMING, ACTIVE, UPCOMING, FINISHED",
	)
	ErrMaxGaugesReached = sdkerrors.Register(
		ModuleName,
		5,
		"Gauge limit has been reached; additional gauges may be created once the gauge limit has been raised via governance proposal",
	)
	ErrGaugePricingTickOutOfRange = sdkerrors.Register(
		ModuleName,
		6,
		"cannot use an invalid price tick",
	)
	ErrGaugeDistrToTickOutOfRange = sdkerrors.Register(ModuleName, 7, "cannot use an distrTo tick")
	ErrInvalidSigner              = sdkerrors.Register(
		ModuleName,
		8,
		"owner must be module authority",
	)
)
