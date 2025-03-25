package revenue

import (
	"fmt"

	cometabcitypes "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/skip-mev/slinky/abci/strategies/codec"
	slinkyabcitypes "github.com/skip-mev/slinky/abci/types"
	slinkyve "github.com/skip-mev/slinky/abci/ve"

	revenuekeeper "github.com/neutron-org/neutron/v6/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

// NewPreBlockHandler returns a new PreBlockHandler. The handler is responsible for recording
// validators' participation in network operations and distribute revenue to validators.
func NewPreBlockHandler(
	revenueKeeper *revenuekeeper.Keeper,
	stakingKeeper revenuetypes.StakingKeeper,
	veCodec codec.VoteExtensionCodec,
	ecCodec codec.ExtendedCommitCodec,
) *PreBlockHandler {
	return &PreBlockHandler{
		revenueKeeper: revenueKeeper,
		stakingKeeper: stakingKeeper,
		veCodec:       veCodec,
		ecCodec:       ecCodec,
	}
}

// PreBlockHandler is responsible for recording validators' participation in network operations and
// distribute revenue to validators.
type PreBlockHandler struct { //golint:ignore
	revenueKeeper *revenuekeeper.Keeper
	stakingKeeper revenuetypes.StakingKeeper

	// codecs
	veCodec codec.VoteExtensionCodec
	ecCodec codec.ExtendedCommitCodec
}

// WrappedPreBlocker is called by the base app before the block is finalized. It is responsible for
// calling the oraclePreBlocker, maintaining data for proper reward asset TWAP calculation,
// distributing revenue to validators, and recording validators' participation in network operations.
func (h *PreBlockHandler) WrappedPreBlocker(oraclePreBlocker sdktypes.PreBlocker) sdktypes.PreBlocker {
	return func(ctx sdktypes.Context, req *cometabcitypes.RequestFinalizeBlock) (response *sdktypes.ResponsePreBlock, err error) {
		response, err = oraclePreBlocker(ctx, req)
		if err != nil {
			return response, fmt.Errorf("oracle module PreBlock failed: %w", err)
		}

		// If vote extensions are not enabled, then we don't need to do anything.
		if !slinkyve.VoteExtensionsEnabled(ctx) {
			h.revenueKeeper.Logger(ctx).Info("vote extensions are not enabled", "height", ctx.BlockHeight())
			return response, nil
		}

		err = h.revenueKeeper.UpdateRewardAssetPrice(ctx)
		if err != nil {
			h.revenueKeeper.Logger(ctx).Error("failed to update reward asset price", "err", err)
		}

		if len(req.Txs) < slinkyabcitypes.NumInjectedTxs {
			return response, fmt.Errorf("the number of transactions is less than the expected number of injected transactions: %d < %d", len(req.Txs), slinkyabcitypes.NumInjectedTxs)
		}
		extendedCommitInfo, err := h.ecCodec.Decode(req.Txs[slinkyabcitypes.OracleInfoIndex])
		if err != nil {
			return response, fmt.Errorf("failed to decode oracle info indexed tx[%d] as extended commit info: %w", slinkyabcitypes.OracleInfoIndex, err)
		}
		if err := h.ProcessExtendedCommitInfo(ctx, extendedCommitInfo); err != nil {
			return response, fmt.Errorf("error processing extended commit info: %w", err)
		}

		if err := h.PaymentScheduleCheck(ctx); err != nil {
			return response, fmt.Errorf("error checking payment schedule: %w", err)
		}

		return response, nil
	}
}

// PaymentScheduleCheck maintains payment schedule state and consistency, and ensures revenue is
// distributed across validators according to the payment schedule.
func (h *PreBlockHandler) PaymentScheduleCheck(ctx sdktypes.Context) error {
	params, err := h.revenueKeeper.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module params: %w", err)
	}

	ps, err := h.revenueKeeper.GetPaymentScheduleI(ctx)
	if err != nil {
		return fmt.Errorf("failed to get payment schedule: %w", err)
	}

	var psRequiresUpdate bool
	switch {
	// a mismatch means that the payment schedule type has been changed in the previous block by
	// a MsgUpdateParams submission.
	// in this case, we need to distribute revenue for passed part of the payment period, reflect the
	// change in the payment schedule by storing the corresponding payment schedule implementation
	// in the module's store, and prepare to the a new period
	case !ps.MatchesType(params.PaymentScheduleType.PaymentScheduleType):
		h.revenueKeeper.Logger(ctx).Debug("payment schedule type module parameter has changed",
			"new_payment_schedule_type", fmt.Sprintf("%+v", params.PaymentScheduleType),
			"old_payment_schedule_value", ps.String(),
			"effective_period_progress", ps.EffectivePeriodProgress(ctx).String(),
		)

		if err := h.revenueKeeper.ProcessRevenue(ctx, params, ps); err != nil {
			return fmt.Errorf("failed to process revenue: %w", err)
		}

		if err := h.revenueKeeper.ResetValidatorsInfo(ctx); err != nil {
			return fmt.Errorf("failed to reset validators info on payment schedule change: %w", err)
		}

		ps = revenuetypes.PaymentScheduleIByType(params.PaymentScheduleType.PaymentScheduleType)
		ps.StartNewPeriod(ctx)
		psRequiresUpdate = true

	// if the period has ended, revenue needs to be processed and payment schedule set to the next period
	case ps.PeriodEnded(ctx):
		h.revenueKeeper.Logger(ctx).Debug("payment period has ended, processing revenue")
		if err := h.revenueKeeper.ProcessRevenue(ctx, params, ps); err != nil {
			return fmt.Errorf("failed to process revenue: %w", err)
		}
		if err := h.revenueKeeper.ResetValidatorsInfo(ctx); err != nil {
			return fmt.Errorf("failed to reset validators info on revenue distribution: %w", err)
		}
		ps.StartNewPeriod(ctx)
		psRequiresUpdate = true
	}

	if psRequiresUpdate {
		if err := h.revenueKeeper.SetPaymentScheduleI(ctx, ps); err != nil {
			return fmt.Errorf("failed to set payment schedule after changing: %w", err)
		}
		h.revenueKeeper.Logger(ctx).Debug("module payment schedule updated", "new_payment_schedule", ps.String())
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

		validator, err := h.stakingKeeper.GetValidatorByConsAddr(ctx, voteInfo.Validator.Address)
		if err != nil {
			return fmt.Errorf("error getting validator by consensus address: %w", err)
		}
		valoperAddr, err := sdktypes.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			return fmt.Errorf("error converting bech32 validator operator address to sdktypes.ValAddress: %w", err)
		}

		votes = append(votes, revenuetypes.ValidatorParticipation{
			ValOperAddress:      valoperAddr,
			BlockVote:           voteInfo.BlockIdFlag,
			OracleVoteExtension: voteExtension,
		})
	}

	if err := h.revenueKeeper.RecordValidatorsParticipation(ctx, votes); err != nil {
		return fmt.Errorf("failed to record validators participation for current block: %w", err)
	}

	return nil
}
