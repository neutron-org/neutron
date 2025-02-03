package revenue

import (
	"fmt"

	cometabcitypes "github.com/cometbft/cometbft/abci/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	revenuekeeper "github.com/neutron-org/neutron/v5/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/skip-mev/slinky/abci/strategies/codec"
	slinkyabcitypes "github.com/skip-mev/slinky/abci/types"
	slinkyve "github.com/skip-mev/slinky/abci/ve"
)

// NewPreBlockHandler returns a new PreBlockHandler. The handler is responsible for recording
// validators' participation in network operations and distribute revenue to validators.
func NewPreBlockHandler(
	keeper *revenuekeeper.Keeper,
	veCodec codec.VoteExtensionCodec,
	ecCodec codec.ExtendedCommitCodec,
) *PreBlockHandler {
	return &PreBlockHandler{
		keeper:  keeper,
		veCodec: veCodec,
		ecCodec: ecCodec,
	}
}

// PreBlockHandler is responsible for recording validators' participation in network operations and
// distribute revenue to validators.
type PreBlockHandler struct { //golint:ignore
	// keeper is the keeper for the revenue module.
	keeper *revenuekeeper.Keeper

	// codecs
	veCodec codec.VoteExtensionCodec
	ecCodec codec.ExtendedCommitCodec
}

// WrappedPreBlocker is called by the base app before the block is finalized. It is responsible for
// calling the oraclePreBlocker, distributing revenue to validators, and recording validators'
// participation in network operations.
func (h *PreBlockHandler) WrappedPreBlocker(oraclePreBlocker sdktypes.PreBlocker) sdktypes.PreBlocker {
	return func(ctx sdktypes.Context, req *cometabcitypes.RequestFinalizeBlock) (response *sdktypes.ResponsePreBlock, err error) {
		response, err = oraclePreBlocker(ctx, req)
		if err != nil {
			return response, fmt.Errorf("oracle module PreBlock failed: %w", err)
		}

		// If vote extensions are not enabled, then we don't need to do anything.
		if !slinkyve.VoteExtensionsEnabled(ctx) {
			h.keeper.Logger(ctx).Info("vote extensions are not enabled", "height", ctx.BlockHeight())
			return response, nil
		}

		if err := h.PaymentScheduleCheck(ctx); err != nil {
			return response, err
		}

		if len(req.Txs) < slinkyabcitypes.NumInjectedTxs {
			return response, slinkyabcitypes.MissingCommitInfoError{}
		}
		extendedCommitInfo, err := h.ecCodec.Decode(req.Txs[slinkyabcitypes.OracleInfoIndex])
		if err != nil {
			return response, slinkyabcitypes.CodecError{Err: fmt.Errorf("error decoding extended-commit-info: %w", err)}
		}
		if err := h.ProcessExtendedCommitInfo(ctx, extendedCommitInfo); err != nil {
			return response, err
		}

		return response, nil
	}
}

// PaymentScheduleCheck maintains payment schedule state and consistency, and ensures revenue is
// distributed across validators according to the payment schedule.
func (h *PreBlockHandler) PaymentScheduleCheck(ctx sdktypes.Context) error {
	state, err := h.keeper.GetState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module state: %w", err)
	}
	params, err := h.keeper.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module params: %w", err)
	}

	pscv := state.PaymentSchedule.GetCachedValue()
	ps, ok := pscv.(revenuetypes.PaymentSchedule)
	if !ok {
		return fmt.Errorf("expected state.PaymentSchedule to be of type PaymentSchedule, but got %T", pscv)
	}

	var stateRequiresUpdate bool

	// if the period has ended, revenue needs to be processed and module's state set to the next period
	if ps.PeriodEnded(ctx) {
		h.keeper.Logger(ctx).Debug("payment period has ended, processing revenue")
		if err := h.keeper.ProcessRevenue(ctx, params, ps.TotalBlocksInPeriod(ctx)); err != nil {
			return fmt.Errorf("failed to process revenue: %w", err)
		}
		if err := h.keeper.ResetValidatorsInfo(ctx); err != nil {
			return fmt.Errorf("failed to reset validators info on revenue distribution: %w", err)
		}
		ps.StartNewPeriod(ctx)
		stateRequiresUpdate = true
	}

	// a mismatch means that the payment schedule type has been changed in the current block by
	// a MsgUpdateParams submission
	// in this case, we need to reflect the change in the module's State by storing the corresponding
	// payment schedule implementation in the module's State and prepare for the a new period
	if !revenuetypes.PaymentScheduleMatchesType(ps, params.PaymentScheduleType) {
		h.keeper.Logger(ctx).Debug("payment schedule type module parameter has changed",
			"new_payment_schedule_type", params.PaymentScheduleType.String(),
			"old_payment_schedule_value", ps.String(),
		)
		if err := h.keeper.ResetValidatorsInfo(ctx); err != nil {
			return fmt.Errorf("failed to reset validators info on payment schedule change: %w", err)
		}

		ps = revenuetypes.PaymentScheduleByType(params.PaymentScheduleType)
		ps.StartNewPeriod(ctx)
		stateRequiresUpdate = true
	}

	if stateRequiresUpdate {
		packedPs, err := codectypes.NewAnyWithValue(ps)
		if err != nil {
			return fmt.Errorf("failed to pack new payment schedule %+v: %w", ps, err)
		}
		state.PaymentSchedule = packedPs
		if err := h.keeper.SetState(ctx, state); err != nil {
			return fmt.Errorf("failed to set module state after changing payment schedule: %w", err)
		}
		h.keeper.Logger(ctx).Debug("module state updated", "new_state", state.String())
	}

	return nil
}

// ProcessExtendedCommitInfo decodes the extended commit info and records validators' participation
// in the block creation and oracle prices provision.
func (h *PreBlockHandler) ProcessExtendedCommitInfo(ctx sdktypes.Context, extendedCommitInfo cometabcitypes.ExtendedCommitInfo) error {
	votes := make([]revenuetypes.ValidatorParticipation, 0, len(extendedCommitInfo.Votes))
	for _, voteInfo := range extendedCommitInfo.Votes {
		voteExtension, err := h.veCodec.Decode(voteInfo.VoteExtension)
		if err != nil {
			return slinkyabcitypes.CodecError{Err: fmt.Errorf("error decoding vote-extension: %w", err)}
		}

		votes = append(votes, revenuetypes.ValidatorParticipation{
			ConsAddress:         voteInfo.Validator.Address,
			BlockVote:           voteInfo.BlockIdFlag,
			OracleVoteExtension: voteExtension,
		})
	}

	if err := h.keeper.RecordValidatorsParticipation(ctx, votes); err != nil {
		return fmt.Errorf("failed to record validators participation for current block: %w", err)
	}

	return nil
}
