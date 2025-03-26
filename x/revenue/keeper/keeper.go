package keeper

import (
	"context"
	"errors"
	"fmt"
	stdmath "math"

	coretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bech32types "github.com/cosmos/cosmos-sdk/types/bech32"

	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService coretypes.KVStoreService
	bankKeeper   revenuetypes.BankKeeper
	oracleKeeper revenuetypes.OracleKeeper
	authority    string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService coretypes.KVStoreService,
	bankKeeper revenuetypes.BankKeeper,
	oracleKeeper revenuetypes.OracleKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:          cdc,
		storeService: storeService,
		bankKeeper:   bankKeeper,
		oracleKeeper: oracleKeeper,
		authority:    authority,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", revenuetypes.ModuleName))
}

// GetParams gets the revenue module parameters.
func (k *Keeper) GetParams(ctx context.Context) (params revenuetypes.Params, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(revenuetypes.ParamsKey)
	if err != nil {
		return params, fmt.Errorf("failed to read params from the store: %w", err)
	}
	if bz == nil {
		return params, nil
	}

	if err = k.cdc.Unmarshal(bz, &params); err != nil {
		return params, fmt.Errorf("failed to unmarshal params: %w", err)
	}
	return params, nil
}

// SetParams sets the revenue module parameters.
func (k *Keeper) SetParams(ctx context.Context, params revenuetypes.Params) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}
	return store.Set(revenuetypes.ParamsKey, bz)
}

// GetPaymentScheduleI reads the payment schedule from the module store and returns it as a
// PaymentScheduleI.
func (k *Keeper) GetPaymentScheduleI(ctx sdk.Context) (revenuetypes.PaymentScheduleI, error) {
	ps, err := k.getPaymentSchedule(ctx)
	if err != nil {
		return nil, err
	}

	return ps.IntoPaymentScheduleI()
}

// SetPaymentSchedule stores a payment schedule.
func (k *Keeper) SetPaymentSchedule(ctx sdk.Context, ps *revenuetypes.PaymentSchedule) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(ps)
	if err != nil {
		return fmt.Errorf("failed to marshal payment schedule: %w", err)
	}

	if err := store.Set(revenuetypes.PaymentScheduleKey, bz); err != nil {
		return fmt.Errorf("failed to write payment schedule to the store: %w", err)
	}
	return nil
}

// SetPaymentScheduleI wraps a given PaymentScheduleI into a PaymentSchedule and stores it.
func (k *Keeper) SetPaymentScheduleI(ctx sdk.Context, psi revenuetypes.PaymentScheduleI) error {
	return k.SetPaymentSchedule(ctx, psi.IntoPaymentSchedule())
}

func (k *Keeper) GetAllValidatorInfo(ctx sdk.Context) (infos []revenuetypes.ValidatorInfo, err error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(revenuetypes.PrefixValidatorInfoKey, storetypes.PrefixEndBytes(revenuetypes.PrefixValidatorInfoKey))
	if err != nil {
		return nil, fmt.Errorf("failed to iterate over validator info: %w", err)
	}
	defer iter.Close()

	infos = make([]revenuetypes.ValidatorInfo, 0)
	for ; iter.Valid(); iter.Next() {
		info := revenuetypes.ValidatorInfo{}
		if err = k.cdc.Unmarshal(iter.Value(), &info); err != nil {
			return nil, fmt.Errorf("failed to unmarshal a validator info: %w", err)
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (k *Keeper) GetValidatorInfo(ctx sdk.Context, addr sdk.ValAddress) (info revenuetypes.ValidatorInfo, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(revenuetypes.GetValidatorInfoKey(addr))
	if err != nil {
		return info, fmt.Errorf("failed to read validator info from the store: %w", err)
	}
	if bz == nil {
		return info, revenuetypes.ErrNoValidatorInfoFound
	}

	if err = k.cdc.Unmarshal(bz, &info); err != nil {
		return info, fmt.Errorf("failed to unmarshal validator info: %w", err)
	}
	return info, nil
}

func (k *Keeper) SetValidatorInfo(ctx sdk.Context, addr sdk.ValAddress, info revenuetypes.ValidatorInfo) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&info)
	if err != nil {
		return fmt.Errorf("failed to marshal validator info: %w", err)
	}

	if err := store.Set(revenuetypes.GetValidatorInfoKey(addr), bz); err != nil {
		return fmt.Errorf("failed to write validator info to the store: %w", err)
	}
	return nil
}

// RecordValidatorsParticipation checks whether validators have voted for the block and oracle
// prices and updates their info in accordance with the results.
func (k *Keeper) RecordValidatorsParticipation(ctx sdk.Context, votes []revenuetypes.ValidatorParticipation) error {
	for _, vote := range votes {
		valInfo, err := k.getOrCreateValidatorInfo(ctx, vote.ValOperAddress)
		if err != nil {
			return err
		}

		valInfo.InActiveValsetForBlocksInPeriod++
		// BlockIDFlagAbsent means that no block vote has been received from the validator
		if vote.BlockVote == tmtypes.BlockIDFlagAbsent {
			k.Logger(ctx).Debug("missing validator's block signature",
				"validator", vote.ValOperAddress.String(),
				"height", ctx.BlockHeight(),
			)
		} else {
			valInfo.CommitedBlocksInPeriod++
		}
		// empty oracle prices means that no prices vote has been received from the validator
		if len(vote.OracleVoteExtension.Prices) == 0 {
			k.Logger(ctx).Debug("missing validator's oracle prices",
				"validator", vote.ValOperAddress.String(),
				"height", ctx.BlockHeight(),
			)
		} else {
			valInfo.CommitedOracleVotesInPeriod++
		}

		if err := k.SetValidatorInfo(ctx, vote.ValOperAddress, valInfo); err != nil {
			return err
		}
		k.Logger(ctx).Debug("validator participation recorded",
			"validator", vote.ValOperAddress.String(),
			"in_active_valset_for_blocks_in_period", valInfo.InActiveValsetForBlocksInPeriod,
			"committed_blocks_in_period", valInfo.CommitedBlocksInPeriod,
			"committed_oracle_votes_in_period", valInfo.CommitedOracleVotesInPeriod,
		)
	}
	return nil
}

// ProcessRevenue calculates and distributes revenue to validators based on their performance during
// the current period. It determines each validator's compensation, transfers the appropriate amount
// of revenue from the module's treasury pool to the validator's account, and resets the validator's
// performance stats in preparation for the next period.
func (k *Keeper) ProcessRevenue(ctx sdk.Context, params revenuetypes.Params, ps revenuetypes.PaymentScheduleI) error {
	epp := ps.EffectivePeriodProgress(ctx)
	if epp.LT(math.LegacyZeroDec()) || epp.GT(math.LegacyOneDec()) {
		return fmt.Errorf("invalid effective period progress %s: expected to be between 0.0 and 1.0", epp.String())
	}

	infos, err := k.GetAllValidatorInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all validator info: %w", err)
	}
	baseRevenueAmount, err := k.CalcBaseRevenueAmount(ctx)
	if err != nil {
		return fmt.Errorf("failed to calculate base revenue amount: %w", err)
	}
	periodRevenueAmount := k.CalcPeriodRevenueAmount(baseRevenueAmount, epp)
	if periodRevenueAmount.IsZero() {
		ctx.EventManager().EmitEvent(sdk.NewEvent(revenuetypes.EventTypeRevenueDistributionNone,
			sdk.NewAttribute(revenuetypes.EventAttributeRevenueAmount, periodRevenueAmount.String()),
			sdk.NewAttribute(revenuetypes.EventAttributeEffectivePeriodProgress, epp.String()),
		))
		return nil // nothing to distribute
	}
	blocksInPeriod := ps.TotalBlocksInPeriod(ctx)

	for _, info := range infos {
		pr, valCompensation := evaluateValCommitment(
			params,
			periodRevenueAmount,
			info,
			blocksInPeriod,
		)

		_, valOperAddrBytes, err := bech32types.DecodeAndConvert(info.ValOperAddress)
		if err != nil {
			return fmt.Errorf("failed to convert valoper address %s to bytes: %w", info.ValOperAddress, err)
		}

		if !valCompensation.IsPositive() {
			emitDistributeRevenueEvent(ctx, info, sdk.NewCoin(params.RewardAsset, math.ZeroInt()), pr, blocksInPeriod, epp)
			continue
		}

		revenueAmt := sdk.NewCoin(params.RewardAsset, valCompensation)
		err = k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			revenuetypes.RevenueTreasuryPoolName,
			valOperAddrBytes,
			sdk.NewCoins(revenueAmt),
		)
		if err != nil {
			ctx.EventManager().EmitEvent(sdk.NewEvent(revenuetypes.EventTypeRevenueDistributionError,
				sdk.NewAttribute(revenuetypes.EventAttributeValidator, info.ValOperAddress),
				sdk.NewAttribute(revenuetypes.EventAttributePaymentFailure, err.Error()),
				sdk.NewAttribute(revenuetypes.EventAttributeRevenueAmount, revenueAmt.String()),
			))
			k.Logger(ctx).Debug("failed to send revenue to validator", "validator", info.ValOperAddress, "err", err)
		} else {
			emitDistributeRevenueEvent(ctx, info, revenueAmt, pr, blocksInPeriod, epp)
			k.Logger(ctx).Debug("revenue sent to validator", "validator", info.ValOperAddress, "revenue", revenueAmt.String())
		}
	}
	return nil
}

// ResetValidatorsInfo resets the validators' performance info in the module state.
func (k *Keeper) ResetValidatorsInfo(ctx sdk.Context) error {
	infos, err := k.GetAllValidatorInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all validator info: %w", err)
	}
	store := k.storeService.OpenKVStore(ctx)

	for _, info := range infos {
		valOperAddr, err := sdk.ValAddressFromBech32(info.ValOperAddress)
		if err != nil {
			return fmt.Errorf("failed to create valoper addr from bech32 %s: %w", info.ValOperAddress, err)
		}
		if err := store.Delete(revenuetypes.GetValidatorInfoKey(valOperAddr)); err != nil {
			return fmt.Errorf("failed to remove validator info from the store: %w", err)
		}
	}
	k.Logger(ctx).Debug("all validators info has been reset")
	return nil
}

// CalcBaseRevenueAmount calculates the base reward for validators measured in reward asset based
// on the current reward asset TWAP and reward quote amount param.
func (k *Keeper) CalcBaseRevenueAmount(ctx sdk.Context) (math.Int, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to get module params: %w", err)
	}

	assetPrice, err := k.GetTWAP(ctx)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to get TWAP: %w", err)
	}
	if assetPrice.Equal(math.LegacyZeroDec()) {
		return math.ZeroInt(), fmt.Errorf("invalid TWAP: price must be greater than zero")
	}

	exp, err := k.getRewardAssetExponent(ctx)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to get reward asset exponent: %w", err)
	}

	return math.LegacyNewDecFromInt(math.NewIntFromUint64(params.RewardQuote.Amount)).
		Quo(assetPrice).
		MulInt64(int64(stdmath.Pow10(int(exp)))).
		TruncateInt(), nil
}

// CalcPeriodRevenueAmount calculates the compensation amount for validators based on the current
// base revenue amount and effective payment period progress.
func (k *Keeper) CalcPeriodRevenueAmount(
	baseRevenueAmount math.Int,
	effectivePeriodProgress math.LegacyDec,
) math.Int {
	return effectivePeriodProgress.MulInt(baseRevenueAmount).TruncateInt()
}

func (k *Keeper) getOrCreateValidatorInfo(
	ctx sdk.Context,
	addr sdk.ValAddress,
) (info revenuetypes.ValidatorInfo, err error) {
	info, err = k.GetValidatorInfo(ctx, addr)
	if err != nil && !errors.Is(err, revenuetypes.ErrNoValidatorInfoFound) {
		return info, fmt.Errorf("failed to read validator info from the store: %w", err)
	}

	// means there is a validator info entry in the store. otherwise fallback to creation
	if err == nil {
		return info, nil
	}

	info = revenuetypes.ValidatorInfo{
		ValOperAddress: addr.String(),
	}
	if err := k.SetValidatorInfo(ctx, addr, info); err != nil {
		return info, fmt.Errorf("failed to write validator info to the store: %w", err)
	}
	k.Logger(ctx).Debug("new validator info created", "info", info)
	return info, nil
}

// getPaymentSchedule gets the current payment schedule without any transformations.
func (k *Keeper) getPaymentSchedule(ctx sdk.Context) (*revenuetypes.PaymentSchedule, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(revenuetypes.PaymentScheduleKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read payment schedule from the store: %w", err)
	}
	if bz == nil {
		return nil, fmt.Errorf("no payment schedule found in the module store")
	}

	var ps revenuetypes.PaymentSchedule
	if err = k.cdc.Unmarshal(bz, &ps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment schedule: %w", err)
	}
	return &ps, nil
}

// emitDistributeRevenueEvent emits an event that contains information about the revenue distribution.
func emitDistributeRevenueEvent(
	ctx sdk.Context,
	info revenuetypes.ValidatorInfo,
	revenueAmt sdk.Coin,
	rating math.LegacyDec,
	blocksInPeriod uint64,
	effectivePeriodProgress math.LegacyDec,
) {
	ctx.EventManager().EmitEvent(sdk.NewEvent(revenuetypes.EventTypeRevenueDistribution,
		sdk.NewAttribute(revenuetypes.EventAttributeValidator, info.ValOperAddress),
		sdk.NewAttribute(revenuetypes.EventAttributeRevenueAmount, revenueAmt.String()),
		sdk.NewAttribute(revenuetypes.EventAttributePerformanceRating, rating.String()),
		sdk.NewAttribute(revenuetypes.EventAttributeInActiveValsetForBlocksInPeriod, fmt.Sprintf("%d", info.InActiveValsetForBlocksInPeriod)),
		sdk.NewAttribute(revenuetypes.EventAttributeCommittedBlocksInPeriod, fmt.Sprintf("%d", info.CommitedBlocksInPeriod)),
		sdk.NewAttribute(revenuetypes.EventAttributeCommittedOracleVotesInPeriod, fmt.Sprintf("%d", info.CommitedOracleVotesInPeriod)),
		sdk.NewAttribute(revenuetypes.EventAttributeTotalBlockInPeriod, fmt.Sprintf("%d", blocksInPeriod)),
		sdk.NewAttribute(revenuetypes.EventAttributeEffectivePeriodProgress, effectivePeriodProgress.String()),
	))
}
