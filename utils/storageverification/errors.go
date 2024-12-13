package storageverification

import "cosmossdk.io/errors"

const StateVerificationCodespace = "state_verification"

var (
	ErrInvalidType         = errors.Register(StateVerificationCodespace, 1, "invalid type")
	ErrInvalidStorageValue = errors.Register(StateVerificationCodespace, 2, "failed to check storage value")
	ErrInvalidProof        = errors.Register(StateVerificationCodespace, 3, "merkle proof is invalid")
)
