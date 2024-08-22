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

type PreBlockHandler struct { //golint:ignore
	logger log.Logger

	keeper                *lastlookkeeper.Keeper
	consensusParamsKeeper *consensuskeeper.Keeper
}

func NewOraclePreBlockHandler(
	logger log.Logger,
	lastlookKeeper *lastlookkeeper.Keeper,
	consensusParamsKeeper *consensuskeeper.Keeper,
) *PreBlockHandler {
	return &PreBlockHandler{
		logger:                logger,
		keeper:                lastlookKeeper,
		consensusParamsKeeper: consensusParamsKeeper,
	}
}

func (h *PreBlockHandler) PreBlocker() sdk.PreBlocker {
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

		txBlobIndex := 0

		params, err := h.consensusParamsKeeper.Params(ctx, &types2.QueryParamsRequest{})
		if err != nil {
			return nil, err
		}

		if ctx.BlockHeight() > params.GetParams().Abci.GetVoteExtensionsEnableHeight() {
			txBlobIndex = 1
		}

		var txBlob types.TxsBlob
		if h.keeper.GetCodec().Unmarshal(req.Txs[txBlobIndex], &txBlob) != nil {
			h.logger.Error("failed to unmarshal txs blob", "err", err)
			return &sdk.ResponsePreBlock{}, nil
		}

		if err := h.keeper.StoreTxs(ctx, sdk.AccAddress(txBlob.Proposer), txBlob.Txs); err != nil {
			h.logger.Error("failed to store txs", "err", err)
			return nil, err
		}

		return &sdk.ResponsePreBlock{}, nil
	}
}
