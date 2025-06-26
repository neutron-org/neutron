package stateverification

import (
	"cosmossdk.io/errors"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v10/modules/core/exported"
	ics23 "github.com/cosmos/ics23/go"

	"github.com/neutron-org/neutron/v7/x/interchainqueries/types"
)

// ExtraVerifier is a type describing a function that will be called for each storage value from the []StorageValue slice with an index of the value in the slice
// to do any additional user-defined checks of storage values during the slice verification
type ExtraVerifier func(index int) error

// VerifyStorageValues verifies stValues slice against proof using proofSpecs
// A caller can provide extraVerification method that will be called for each storage value from the slice with an index of the value in the slice
// to do any additional user-defined checks of storage values
func VerifyStorageValues(stValues []*types.StorageValue, root exported.Root, proofSpecs []*ics23.ProofSpec, extraVerification ExtraVerifier) error {
	for index, value := range stValues {
		proof, err := ibccommitmenttypes.ConvertProofs(value.Proof)
		if err != nil {
			return errors.Wrapf(ErrInvalidType, "failed to convert crypto.ProofOps to MerkleProof: %v", err)
		}

		if extraVerification != nil {
			if err := extraVerification(index); err != nil {
				return errors.Wrapf(ErrInvalidStorageValue, "extra verification failed at index %d: %v", index, err)
			}
		}

		path := ibccommitmenttypes.NewMerklePath([]byte(value.StoragePrefix), value.Key)
		// identify what kind proofs (non-existence proof always has *ics23.CommitmentProof_Nonexist as the first item) we got
		// and call corresponding method to verify it
		switch proof.GetProofs()[0].GetProof().(type) {
		// we can get non-existence proof if someone queried some key which is not exists in the storage on remote chain
		case *ics23.CommitmentProof_Nonexist:
			if err := proof.VerifyNonMembership(proofSpecs, root, path); err != nil {
				return errors.Wrapf(ErrInvalidProof, "non-membership verification failed for value at index %d with path %s: %v", index, path, err)
			}
		case *ics23.CommitmentProof_Exist:
			if err := proof.VerifyMembership(proofSpecs, root, path, value.Value); err != nil {
				return errors.Wrapf(ErrInvalidProof, "membership verification failed for value at index %d with path %s: %v", index, path, err)
			}
		default:
			return errors.Wrapf(ErrInvalidProof, "unknown proof type %T", proof.GetProofs()[0].GetProof())
		}
	}

	return nil
}
