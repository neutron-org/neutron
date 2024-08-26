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

		currentBlob, err := h.lastlookKeeper.GetBatch(ctx, req.Height)
		if err != nil {
			h.logger.Error("GetBatch returned an error", "err", err)
			return nil, err
		}

		// An inclusion of a tx is delayed for 1 block and because of that CometBFT does not see a tx in a block immediately
		// and tries to include a tx in two blocks in a row.
		// At a block X a tx is added into the queue, at a block X+1 a tx is finally is included into a block.
		// BUT IT PROPOSED BY CometBFT two times, thus we need to filter it out second time to be sure it's not introduced on chain twice.
		filteredTxs := make([][]byte, 0)
		for _, mempoolTx := range req.Txs {
			containsFunc := func(item []byte) bool {
				return bytes.Equal(item, mempoolTx)
			}
			if !slices.ContainsFunc(currentBlob.Txs, containsFunc) {
				filteredTxs = append(filteredTxs, mempoolTx)
			}
		}

		newBlob := types.Batch{
			Proposer: req.ProposerAddress,
			Txs:      filteredTxs,
		}

		newBlobBz, err := h.lastlookKeeper.GetCodec().Marshal(&newBlob)
		if err != nil {
			h.logger.Error("failed to marshal txs blob", "err", err)
			return nil, err
		}

		resp := cometabci.ResponsePrepareProposal{}

		resp.Txs = h.injectAndResize(currentBlob.Txs, newBlobBz, req.MaxTxBytes+int64(len(newBlobBz)))

		return &resp, nil
	}
}

// injectAndResize returns a tx array containing the injectTx at the beginning followed by appTxs.
// The returned transaction array is bounded by maxSizeBytes, and the function is idempotent meaning the
// injectTx will only appear once regardless of how many times you attempt to inject it.
// If injectTx is large enough, all originalTxs may end up being excluded from the returned tx array.
func (h *ProposalHandler) injectAndResize(appTxs [][]byte, injectTx []byte, maxSizeBytes int64) [][]byte {
	//nolint: prealloc
	var (
		returnedTxs   [][]byte
		consumedBytes int64
	)

	// If VEs are enabled and our VE Tx isn't already in the appTxs, inject it here
	if len(injectTx) != 0 && (len(appTxs) < 1 || !bytes.Equal(appTxs[0], injectTx)) {
		injectBytes := int64(len(injectTx))
		// Ensure the VE Tx is in the response if we have room.
		// We may want to be more aggressive in the future about dedicating block space for application-specific Txs.
		// However, the VE Tx size should be relatively stable so MaxTxBytes should be set w/ plenty of headroom.
		if injectBytes <= maxSizeBytes {
			consumedBytes += injectBytes
			returnedTxs = append(returnedTxs, injectTx)
		}
	}
	// Add as many appTxs to the returned proposal as possible given our maxSizeBytes constraint
	for _, tx := range appTxs {
		consumedBytes += int64(len(tx))
		if consumedBytes > maxSizeBytes {
			return returnedTxs
		}
		returnedTxs = append(returnedTxs, tx)
	}
	return returnedTxs
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
