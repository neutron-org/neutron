# Overview

The State Verifier module allows to verify that some `storage values` were indeed present on a particular `block height` in the chain.

The idea is the similar how Neutron's KV ICQ works: each `StorageValue` in Cosmos SDK is stored in [KV-IAVL storage](https://github.com/cosmos/iavl).
And to be more precise it's stored in a structure called [`MerkleTree`](https://github.com/cosmos/cosmos-sdk/blob/ae77f0080a724b159233bd9b289b2e91c0de21b5/docs/interfaces/lite/specification.md).
The tree allows to compose `Proof` for `key` and `value` pairs that can prove two things using `RootHash` of the tree:
* `key` and `value` are present in the tree;
* `key` is not present in the tree.

Cosmos blockchain's storage is stored as a different tree for each block.
That means we can prove that a particular `KV` pair is really present (or not present) in the storage at a particular block height.

# Implementation

### BeginBlocker
In each block the module's `BeginBlocker` is being called and it saves `ConsensusState` of the current block height in the storage to use it for verification of storage values later:

```go
consensusState := tendermint.ConsensusState{
    Timestamp:          ctx.BlockTime(), // current block time
    Root:               ibccommitmenttypes.NewMerkleRoot(headerInfo.AppHash), // .AppHash for the previous block
    NextValidatorsHash: cometInfo.GetValidatorsHash(), // hash of the validator set for the next block
}
```

For verification only `.Root` (`header.AppHash`) is used, but it's good to save all the values just in case and do not leave them empty.

### VerifyStateValues query
The main query of the module that accepts slice of `[]StorageValue` structures and `blockHeight` on which those `StorageValues` are present.
The module verifies the values and returns an error if values cannot be verified `{valid: true}` structure if values are valid.