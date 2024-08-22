package keeper

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

func (k *Keeper) StoreTxs(ctx sdk.Context, proposer sdk.Address, txs [][]byte) error {
	store := ctx.KVStore(k.storeKey)
	blob := types.TxsBlob{
		Proposer: proposer.Bytes(),
		Txs:      txs,
	}

	bz, err := k.cdc.Marshal(&blob)
	if err != nil {
		return errors.Wrapf(types.ErrProtoMarshal, "failed to marshal txs blob: %v", err)
	}

	store.Set(types.GetTxsQueueKey(ctx.BlockHeight()+1), bz)

	return nil
}

func (k *Keeper) GetTxsBlob(ctx sdk.Context, blockHeight int64) (*types.TxsBlob, error) {
	store := ctx.KVStore(k.storeKey)

	var blob types.TxsBlob

	if blockHeight == 1 {
		return &blob, nil
	}

	bz := store.Get(types.GetTxsQueueKey(blockHeight))
	if bz == nil {
		return nil, errors.Wrapf(types.ErrNoBlob, "no txs blob found for a block %d", blockHeight)
	}

	if err := k.cdc.Unmarshal(bz, &blob); err != nil {
		return nil, errors.Wrapf(types.ErrProtoUnmarshal, "failed to unmarshal tx blob: %v", err)
	}

	return &blob, nil
}

func (k *Keeper) GetProposerForBlock(ctx sdk.Context, blockHeight int64) (sdk.Address, error) {
	blob, err := k.GetTxsBlob(ctx, blockHeight)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get txs blob for block height %d: %v", blockHeight, err)
	}

	return sdk.AccAddress(blob.Proposer), nil
}

func (k *Keeper) RemoveTxsBlob(ctx sdk.Context, blockHeight int64) error {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.GetTxsQueueKey(blockHeight))

	return nil
}
