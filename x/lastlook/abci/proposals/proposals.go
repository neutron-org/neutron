package proposals

import (
	"bytes"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	lastlookabcitypes "github.com/neutron-org/neutron/v4/x/lastlook/abci/types"
	"github.com/neutron-org/neutron/v4/x/lastlook/keeper"
	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

type ProposalHandler struct {
	logger         log.Logger
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

func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *cometabci.RequestPrepareProposal) (*cometabci.ResponsePrepareProposal, error) {
		if req == nil {
			h.logger.Error("PrepareProposalHandler received a nil request")
			return nil, sdkerrors.ErrInvalidRequest
		}

		currentBlob, err := h.lastlookKeeper.GetTxsBlob(ctx, req.Height)
		if err != nil {
			h.logger.Error("GetTxsBlob returned an error", "err", err)
			return nil, err
		}

		newBlob := types.TxsBlob{
			Proposer: req.ProposerAddress,
			Txs:      req.Txs,
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

func (h *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *cometabci.RequestProcessProposal) (resp *cometabci.ResponseProcessProposal, err error) {
		// this should never happen, but just in case
		if req == nil {
			h.logger.Error("ProcessProposalHandler received a nil request")
			return nil, sdkerrors.ErrInvalidRequest
		}

		var txBlob types.TxsBlob
		if err := h.lastlookKeeper.GetCodec().Unmarshal(req.Txs[lastlookabcitypes.TxBlobIndex], &txBlob); err != nil {
			h.logger.Error("failed to unmarshal txs blob", "err", err)
			return nil, err
		}

		req.Txs = req.Txs[lastlookabcitypes.NumInjectedTxs:]

		currentBlob, err := h.lastlookKeeper.GetTxsBlob(ctx, req.Height)
		if err != nil {
			h.logger.Error("GetTxsBlob returned an error", "err", err)
			return nil, err
		}

		if len(req.Txs) != len(currentBlob.Txs) {
			return &cometabci.ResponseProcessProposal{Status: cometabci.ResponseProcessProposal_REJECT}, errors.New(fmt.Sprintf("len(req.Txs) != len(currentBlob.Txs): %v != %v", len(req.Txs), len(currentBlob.Txs)))
		}

		for i := 0; i < len(currentBlob.Txs); i++ {
			if !bytes.Equal(req.Txs[i], currentBlob.Txs[i]) {
				return &cometabci.ResponseProcessProposal{Status: cometabci.ResponseProcessProposal_REJECT}, errors.New(fmt.Sprintf("req.Txs[i] != currentBlob.Txs[i]: %v != %v", req.Txs[i], currentBlob.Txs[i]))
			}

		}

		if !bytes.Equal(req.ProposerAddress, txBlob.Proposer) {
			return &cometabci.ResponseProcessProposal{Status: cometabci.ResponseProcessProposal_REJECT}, errors.New(fmt.Sprintf("req.ProposerAddress != proposer in current blob: %v != %v", req.ProposerAddress, txBlob.Proposer))
		}

		return &cometabci.ResponseProcessProposal{Status: cometabci.ResponseProcessProposal_ACCEPT}, nil
	}
}
