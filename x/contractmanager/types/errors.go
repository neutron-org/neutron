package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/contractmanager module sentinel errors
var (
	IncorrectAckType           = errors.Register(ModuleName, 1100, "incorrect acknowledgement type")
	IncorrectFailureToResubmit = errors.Register(ModuleName, 1101, "incorrect failure to resubmit")
)
