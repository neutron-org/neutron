package revenue

import (
	"fmt"

	cometabcitypes "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	revenuekeeper "github.com/neutron-org/neutron/v5/x/revenue/keeper"
)

func NewPreBlockHandler(keeper *revenuekeeper.Keeper) *PreBlockHandler {
	return &PreBlockHandler{keeper: keeper}
}

// PreBlockHandler is responsible for aggregating oracle data from each
// validator and writing the oracle data into the store before any transactions
// are executed/finalized for a given block.
type PreBlockHandler struct { //golint:ignore
	// keeper is the keeper for the revenue module.
	keeper *revenuekeeper.Keeper
}

// WrappedPreBlocker is called by the base app before the block is finalized. It
// is responsible for calling the module manager's PreBlock method, aggregating oracle data from each validator and
// writing the oracle data to the store.
func (h *PreBlockHandler) WrappedPreBlocker(oraclePreBlock sdktypes.PreBlocker) sdktypes.PreBlocker {
	return func(ctx sdktypes.Context, req *cometabcitypes.RequestFinalizeBlock) (response *sdktypes.ResponsePreBlock, err error) {
		response, err = oraclePreBlock(ctx, req)
		if err != nil {
			return response, fmt.Errorf("oracle module PreBlock failed: %w", err)
		}

		if err := h.keeper.PreBlock(ctx); err != nil {
			return response, fmt.Errorf("revenue module PreBlock failed: %w", err)
		}

		return response, nil
	}
}
