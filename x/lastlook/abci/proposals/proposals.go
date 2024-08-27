package proposals

import (
	"bytes"
	"fmt"

	"cosmossdk.io/log"
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"golang.org/x/exp/slices"

	lastlookabcitypes "github.com/neutron-org/neutron/v4/x/lastlook/abci/types"
	"github.com/neutron-org/neutron/v4/x/lastlook/keeper"
	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

// ProposalHandler is responsible primarily for:
//  1. Filling a proposal with transactions.
//  2. Injecting last look batch into the proposal.
//  3. Verifying that the last look batch injected is valid.
type ProposalHandler struct {
	logger log.Logger

	// the keeper of the last look module
	lastlookKeeper *keeper.Keeper
}

// NewProposalHandler returns a new ProposalHandler.
func NewProposalHandler(
	logger log.Logger,
	lastlookKeeper *keeper.Keeper,
) *ProposalHandler {
	handler := &ProposalHandler{
		logger:         logger,
		lastlookKeeper: lastlookKeeper,
	}

	return handler
}

// PrepareProposalHandler returns a PrepareProposalHandler that will be called
// by base app when a new block proposal is requested. The PrepareProposalHandler
// will first get txs batch from the queue. Then the handler will inject the new batch info (composed with txs from CometBFT mempool) into the proposal.
func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *cometabci.RequestPrepareProposal) (*cometabci.ResponsePrepareProposal, error) {
		if req == nil {
			h.logger.Error("PrepareProposalHandler received a nil request")
			return nil, sdkerrors.ErrInvalidRequest
		}

		currentBatch, err := h.lastlookKeeper.GetBatch(ctx, req.Height)
		if err != nil {
			h.logger.Error("GetBatch returned an error", "err", err)
			return nil, err
		}

		consumedBytes := int64(0)
		// increase consumedBytes counter by a size of all transaction from a batch that must be inserted to a block
		for _, tx := range req.Txs {
			consumedBytes += int64(len(tx))
		}

		newBatch := types.Batch{
			Proposer: req.ProposerAddress,
			Txs:      make([][]byte, 0, len(req.Txs)),
		}

		for _, mempoolTx := range req.Txs {
			containsFunc := func(item []byte) bool {
				return bytes.Equal(item, mempoolTx)
			}

			// An inclusion of a tx is delayed for 1 block and because of that CometBFT does not see a tx in a block immediately
			// and tries to include a tx in two blocks in a row.
			// At a block X a tx is added into the queue, at a block X+1 a tx is finally included into a block.
			// BUT IT IS PROPOSED BY CometBFT two times, thus we need to filter it out second time to be sure it's not introduced on chain twice.
			if slices.ContainsFunc(currentBatch.Txs, containsFunc) {
				continue
			}

			newBatch.Txs = append(newBatch.Txs, mempoolTx)

			// if final size of a serialised tx is too big to insert into a block, remove last inserted tx and continue
			// iterating over txs from mempool, maybe there is some small tx we can still fit in
			if int64(newBatch.XXX_Size())+consumedBytes > req.MaxTxBytes {
				newBatch.Txs = newBatch.Txs[:len(newBatch.Txs)-1]
			}
		}

		newBatchBz, err := h.lastlookKeeper.GetCodec().Marshal(&newBatch)
		if err != nil {
			h.logger.Error("failed to marshal txs batch", "err", err)
			return nil, err
		}

		resp := cometabci.ResponsePrepareProposal{
			// preallocate memory for all transactions from the current batch + 1 place for a new batch
			Txs: make([][]byte, 0, len(currentBatch.Txs)+1),
		}

		// a new batch must the first
		resp.Txs = append(resp.Txs, newBatchBz)

		// and then all txs from the current batch
		resp.Txs = append(resp.Txs, currentBatch.Txs...)

		return &resp, nil
	}
}

// ProcessProposalHandler returns a ProcessProposalHandler that will be called
// by base app when a new block proposal needs to be verified. The ProcessProposalHandler
// will verify that the last look batch included in the proposal is valid.
func (h *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *cometabci.RequestProcessProposal) (resp *cometabci.ResponseProcessProposal, err error) {
		// this should never happen, but just in case
		if req == nil {
			h.logger.Error("ProcessProposalHandler received a nil request")
			return nil, sdkerrors.ErrInvalidRequest
		}

		var txBlob types.Batch
		if err := h.lastlookKeeper.GetCodec().Unmarshal(req.Txs[lastlookabcitypes.TxBlobIndex], &txBlob); err != nil {
			h.logger.Error("failed to unmarshal txs blob", "err", err)
			return nil, err
		}

		req.Txs = req.Txs[lastlookabcitypes.NumInjectedTxs:]

		currentBlob, err := h.lastlookKeeper.GetBatch(ctx, req.Height)
		if err != nil {
			h.logger.Error("GetBatch returned an error", "err", err)
			return nil, err
		}

		if len(req.Txs) != len(currentBlob.Txs) {
			return &cometabci.ResponseProcessProposal{Status: cometabci.ResponseProcessProposal_REJECT}, fmt.Errorf("len(req.Txs) != len(currentBlob.Txs): %v != %v", len(req.Txs), len(currentBlob.Txs))
		}

		for i := 0; i < len(currentBlob.Txs); i++ {
			if !bytes.Equal(req.Txs[i], currentBlob.Txs[i]) {
				return &cometabci.ResponseProcessProposal{Status: cometabci.ResponseProcessProposal_REJECT}, fmt.Errorf("req.Txs[i] != currentBlob.Txs[i]: %v != %v", req.Txs[i], currentBlob.Txs[i])
			}
		}

		if !bytes.Equal(req.ProposerAddress, txBlob.Proposer) {
			return &cometabci.ResponseProcessProposal{Status: cometabci.ResponseProcessProposal_REJECT}, fmt.Errorf("req.ProposerAddress != proposer in current blob: %v != %v", req.ProposerAddress, txBlob.Proposer)
		}

		return &cometabci.ResponseProcessProposal{Status: cometabci.ResponseProcessProposal_ACCEPT}, nil
	}
}
