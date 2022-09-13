package adminmodule

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/neutron-org/neutron/x/adminmodule/keeper"
	"github.com/neutron-org/neutron/x/adminmodule/types"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := keeper.Logger(ctx)

	// fetch active proposals whose voting periods have ended (are passed the block time)
	keeper.IterateActiveProposalsQueue(ctx, ctx.BlockHeader().Time, func(proposal govtypes.Proposal) bool {
		var logMsg string

		handler := keeper.Router().GetRoute(proposal.ProposalRoute())
		cacheCtx, writeCache := ctx.CacheContext()

		// The proposal handler may execute state mutating logic depending
		// on the proposal content. If the handler fails, no state mutation
		// is written and the error message is logged.
		err := handler(cacheCtx, proposal.GetContent())
		if err == nil {
			logMsg = "passed"
			// write state to the underlying multi-store
			writeCache()
		} else {
			logMsg = fmt.Sprintf("passed, but failed on execution: %s", err)
		}

		proposal.Status = govtypes.StatusPassed

		keeper.SetProposal(ctx, proposal)
		keeper.RemoveFromActiveProposalQueue(ctx, proposal.ProposalId, proposal.SubmitTime.Add(2*time.Second)) // TODO hardcode

		keeper.AddToArchive(ctx, proposal)

		logger.Info(
			"proposal tallied",
			"proposal", proposal.ProposalId,
			"title", proposal.GetTitle(),
			"result", logMsg,
		)

		// TODO event?
		return false
	})
}
