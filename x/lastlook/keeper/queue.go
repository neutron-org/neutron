package keeper

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

// StoreBatch stores a slice of txs and an address of a proposer who proposed a block into a queue
func (k *Keeper) StoreBatch(ctx sdk.Context, height int64, proposer sdk.Address, txs [][]byte) error {
	store := ctx.KVStore(k.storeKey)
	blob := types.Batch{
		Proposer: proposer.Bytes(),
		Txs:      txs,
	}

	bz, err := k.cdc.Marshal(&blob)
	if err != nil {
		return errors.Wrapf(types.ErrProtoMarshal, "failed to marshal txs blob: %v", err)
	}

	store.Set(types.GetTxsQueueKey(height), bz)

	return nil
}

// GetBatch returns a batch of txs that must be inserted into a block at `blockHeight`
func (k *Keeper) GetBatch(ctx sdk.Context, blockHeight int64) (*types.Batch, error) {
	store := ctx.KVStore(k.storeKey)

	var blob types.Batch

	if blockHeight <= 1 {
		return &blob, nil
	}

	bz := store.Get(types.GetTxsQueueKey(blockHeight))
	if bz == nil {
		return nil, errors.Wrapf(types.ErrNoBlob, "no txs batch found for a block %d", blockHeight)
	}

	if err := k.cdc.Unmarshal(bz, &blob); err != nil {
		return nil, errors.Wrapf(types.ErrProtoUnmarshal, "failed to unmarshal tx batch: %v", err)
	}

	return &blob, nil
}

// GetProposerForBlock returns a proposer who proposed to insert a batch of txs into a block at height `blockHeight`
func (k *Keeper) GetProposerForBlock(ctx sdk.Context, blockHeight int64) (sdk.Address, error) {
	blob, err := k.GetBatch(ctx, blockHeight)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get txs blob for block height %d: %v", blockHeight, err)
	}

	return sdk.AccAddress(blob.Proposer), nil
}

// RemoveBatch removes a batch from a queue
func (k *Keeper) RemoveBatch(ctx sdk.Context, blockHeight int64) error {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.GetTxsQueueKey(blockHeight))

	return nil
}
