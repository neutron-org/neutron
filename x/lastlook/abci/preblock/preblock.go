package preblock

import (
	"fmt"

	"cosmossdk.io/log"
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	types2 "github.com/cosmos/cosmos-sdk/x/consensus/types"

	lastlookabcitypes "github.com/neutron-org/neutron/v4/x/lastlook/abci/types"
	lastlookkeeper "github.com/neutron-org/neutron/v4/x/lastlook/keeper"
	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

// LastLookPreBlockHandler is responsible for writing the last look batch data into the store before any transactions
// are executed/finalized for a given block.
type LastLookPreBlockHandler struct { //golint:ignore
	logger log.Logger

	// keeper is the keeper for the last look module. This is utilized to write
	// last look data to state.
	keeper *lastlookkeeper.Keeper

	// consensusParamsKeeper is the keeper for the consensus module. This is utilized to get an info about a height
	// where vote extensions were enabled
	consensusParamsKeeper *consensuskeeper.Keeper
}

// NewLastLookPreBlockHandler returns a new LastLookPreBlockHandler. The handler
// is responsible for writing last look data included in a block to state.
func NewLastLookPreBlockHandler(
	logger log.Logger,
	lastlookKeeper *lastlookkeeper.Keeper,
	consensusParamsKeeper *consensuskeeper.Keeper,
) *LastLookPreBlockHandler {
	return &LastLookPreBlockHandler{
		logger:                logger,
		keeper:                lastlookKeeper,
		consensusParamsKeeper: consensusParamsKeeper,
	}
}

// PreBlocker is called by the base app before the block is finalized. It
// is responsible for writing the last look batch data into the store before any transactions
// are executed/finalized for a given block.
func (h *LastLookPreBlockHandler) PreBlocker() sdk.PreBlocker {
	return func(ctx sdk.Context, req *cometabci.RequestFinalizeBlock) (_ *sdk.ResponsePreBlock, err error) {
		if req == nil {
			ctx.Logger().Error(
				"received nil RequestFinalizeBlock in oracle preblocker",
				"height", ctx.BlockHeight(),
			)

			return &sdk.ResponsePreBlock{}, fmt.Errorf("received nil RequestFinalizeBlock in oracle preblocker: height %d", ctx.BlockHeight())
		}

		h.logger.Debug(
			"executing the pre-finalize block hook",
			"height", req.Height,
		)

		if len(req.Txs) < lastlookabcitypes.NumInjectedTxs {
			h.logger.Error("block doesn't contain TxBlob")
			return &sdk.ResponsePreBlock{}, nil
		}

		// by default, last look batch data is inserted as a first tx in a block (index 0)
		batchTxIndex := 0

		params, err := h.consensusParamsKeeper.Params(ctx, &types2.QueryParamsRequest{})
		if err != nil {
			return nil, err
		}

		// but if votes extensions are enabled, a block includes a slinky info as a first tx and we have last look batch info
		// as a second tx in a block (index 1)
		if ctx.BlockHeight() > params.GetParams().Abci.GetVoteExtensionsEnableHeight() {
			batchTxIndex = 1
		}

		// if for some reason vote extensions are enabled, but we don't have slinky tx included, we have last look tx batch
		// as a first tx in a block
		// mostly it happens in unit tests
		if len(req.Txs) <= batchTxIndex {
			batchTxIndex = 0
		}

		var txBlob types.Batch
		if h.keeper.GetCodec().Unmarshal(req.Txs[batchTxIndex], &txBlob) != nil {
			h.logger.Error("failed to unmarshal txs blob", "err", err)
			return &sdk.ResponsePreBlock{}, nil
		}

		if err := h.keeper.StoreBatch(ctx, ctx.BlockHeight()+1, sdk.AccAddress(txBlob.Proposer), txBlob.Txs); err != nil {
			h.logger.Error("failed to store txs", "err", err)
			return nil, err
		}

		return &sdk.ResponsePreBlock{}, nil
	}
}
