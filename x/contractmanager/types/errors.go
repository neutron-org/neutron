package types

import (
	"cosmossdk.io/errors"
)

// x/contractmanager module sentinel errors
var (
	ErrIncorrectFailureToResubmit = errors.Register(ModuleName, 1101, "incorrect failure to resubmit")
	ErrFailedToResubmitFailure    = errors.Register(ModuleName, 1102, "failed to resubmit failure")
	ErrSudoOutOfGas               = errors.Register(ModuleName, 1103, "sudo handling went beyond the gas limit allowed by the module")
	ErrNotContractResubmission    = errors.Register(ModuleName, 1104, "failures resubmission is only allowed to be called by a smart contract")
)
