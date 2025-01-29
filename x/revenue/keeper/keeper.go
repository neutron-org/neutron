package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/core/comet"
	coretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bech32types "github.com/cosmos/cosmos-sdk/types/bech32"

	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
)

type Keeper struct {
	cdc            codec.BinaryCodec
	storeService   coretypes.KVStoreService
	voteAggregator revenuetypes.VoteAggregator
	stakingKeeper  revenuetypes.StakingKeeper
	bankKeeper     revenuetypes.BankKeeper
	oracleKeeper   revenuetypes.OracleKeeper
	authority      string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService coretypes.KVStoreService,
	voteAggregator revenuetypes.VoteAggregator,
	stakingKeeper revenuetypes.StakingKeeper,
	bankKeeper revenuetypes.BankKeeper,
	oracleKeeper revenuetypes.OracleKeeper,
	authority string,
) *Keeper {
	// ensure bonded and not bonded module accounts are set
	// if addr := ak.GetModuleAddress(types.BondedPoolName); addr == nil {
	//	panic(fmt.Sprintf("%s module account has not been set", types.BondedPoolName))
	// }
	//
	// if addr := ak.GetModuleAddress(types.NotBondedPoolName); addr == nil {
	//	panic(fmt.Sprintf("%s module account has not been set", types.NotBondedPoolName))
	// }
	//
	// ensure that authority is a valid AccAddress
	// if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
	//	panic("authority is not a valid acc address")
	// }
	return &Keeper{
		cdc:            cdc,
		storeService:   storeService,
		voteAggregator: voteAggregator,
		stakingKeeper:  stakingKeeper,
		bankKeeper:     bankKeeper,
		oracleKeeper:   oracleKeeper,
		authority:      authority,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", revenuetypes.ModuleName))
}

// EndBlock records validators' participation in block creation and oracle price provisioning,
// ensuring the module's state remains up to date. At the start of each month, it calculates and
// distributes rewards to all validators based on their performance during the previous period.
func (k *Keeper) EndBlock(ctx sdk.Context) error {
	state, err := k.GetState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module state: %w", err)
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module params: %w", err)
	}

	pscv := state.PaymentSchedule.GetCachedValue()
	ps, ok := pscv.(revenuetypes.PaymentSchedule)
	if !ok {
		return fmt.Errorf("expected state.PaymentSchedule to be of type PaymentSchedule, but got %T", pscv)
	}

	var stateRequiresUpdate bool
	switch {
	// payment schedule either haven't been set or has been changed in the current block
	// in this case, we need to reflect the change in the currently used payment schedule
	case !revenuetypes.PaymentScheduleMatchesType(ps, params.PaymentScheduleType):
		ps = revenuetypes.PaymentScheduleByType(params.PaymentScheduleType)
		ps.StartNewPeriod(ctx)
		stateRequiresUpdate = true

	// if the period has ended, revenue needs to be processed and module's state set to the next period
	case ps.PeriodEnded(ctx):
		if err := k.ProcessRevenue(ctx, params, ps.TotalBlocksInPeriod(ctx)); err != nil {
			return fmt.Errorf("failed to process revenue: %w", err)
		}
		ps.StartNewPeriod(ctx)
		stateRequiresUpdate = true
	}
	if stateRequiresUpdate {
		packedPs, err := codectypes.NewAnyWithValue(ps)
		if err != nil {
			return fmt.Errorf("failed to pack new payment schedule %+v: %w", ps, err)
		}
		state.PaymentSchedule = packedPs
		if err := k.SetState(ctx, state); err != nil {
			return fmt.Errorf("failed to set module state after changing payment schedule: %w", err)
		}
	}

	if err := k.RecordValidatorsParticipation(ctx); err != nil {
		return fmt.Errorf("failed to record validators participation for current block: %w", err)
	}

	return nil
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

func (k *Keeper) GetState(ctx sdk.Context) (state revenuetypes.State, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(revenuetypes.StateKey)
	if err != nil {
		return state, fmt.Errorf("failed to read state from the store: %w", err)
	}
	if bz == nil {
		return state, nil
	}

	if err = k.cdc.Unmarshal(bz, &state); err != nil {
		return state, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	return state, nil
}

func (k *Keeper) SetState(ctx sdk.Context, state revenuetypes.State) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err = store.Set(revenuetypes.StateKey, bz); err != nil {
		return fmt.Errorf("failed to write state to the store: %w", err)
	}
	return nil
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

func (k *Keeper) GetValidatorInfo(ctx sdk.Context, addr sdk.ConsAddress) (info revenuetypes.ValidatorInfo, err error) {
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

func (k *Keeper) SetValidatorInfo(ctx sdk.Context, addr sdk.ConsAddress, info revenuetypes.ValidatorInfo) error {
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
func (k *Keeper) RecordValidatorsParticipation(ctx sdk.Context) error {
	for _, info := range ctx.VoteInfos() {
		var blockVote bool  // whether the validator has voted for the block
		var oracleVote bool // whether the validator has voted for the oracle prices

		// BlockIDFlagAbsent means that no block vote has been received from the validator
		if comet.BlockIDFlag(info.BlockIdFlag) == comet.BlockIDFlagAbsent {
			k.Logger(ctx).Debug("missing validator's block signature",
				"validator", info.Validator.Address,
				"height", ctx.BlockHeight(),
			)
		} else {
			blockVote = true
		}
		// empty oracle prices means that no prices vote has been received from the validator
		if len(k.voteAggregator.GetPriceForValidator(sdk.ConsAddress(info.Validator.Address))) == 0 {
			k.Logger(ctx).Debug("missing validator's oracle prices",
				"validator", info.Validator.Address,
				"height", ctx.BlockHeight(),
			)
		} else {
			oracleVote = true
		}
		if !oracleVote && !blockVote {
			continue // nothing to update for the validator
		}

		// update validator's info in the module state
		valInfo, err := k.getOrCreateValidatorInfo(ctx, info.Validator.Address)
		if err != nil {
			return err
		}
		if oracleVote {
			valInfo.CommitedOracleVotesInPeriod++
		}
		if blockVote {
			valInfo.CommitedBlocksInPeriod++
		}
		if err := k.SetValidatorInfo(ctx, info.Validator.Address, valInfo); err != nil {
			return err
		}
	}
	return nil
}

// ProcessRevenue calculates and distributes revenue to validators based on their performance during
// the current period. It determines each validator's compensation, transfers the appropriate amount
// of revenue from the module's treasury pool to the validator's account, and resets the validator's
// performance stats in preparation for the next period.
func (k *Keeper) ProcessRevenue(ctx sdk.Context, params revenuetypes.Params, blocksPerPeriod uint64) error {
	infos, err := k.GetAllValidatorInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all validator info: %w", err)
	}
	baseCompensation := k.CalcBaseRevenueAmount(ctx)

	for _, info := range infos {
		valConsAddr, err := sdk.ConsAddressFromBech32(info.ConsensusAddress)
		if err != nil {
			return fmt.Errorf("failed to create valcons addr from bech32 %s: %w", info.ConsensusAddress, err)
		}

		rating := PerformanceRating(
			params.BlocksPerformanceRequirement,
			params.OracleVotesPerformanceRequirement,
			int64(blocksPerPeriod-info.GetCommitedBlocksInPeriod()),
			int64(blocksPerPeriod-info.GetCommitedOracleVotesInPeriod()),
			int64(blocksPerPeriod),
		)
		valCompensation := rating.MulInt64(baseCompensation).TruncateInt()

		if valCompensation.IsPositive() {
			validator, err := k.stakingKeeper.GetValidatorByConsAddr(ctx, valConsAddr)
			if err != nil {
				return fmt.Errorf("failed to get validator by cons addr %s from staking keeper: %w", valConsAddr, err)
			}

			_, valOperAddr, err := bech32types.DecodeAndConvert(validator.OperatorAddress)
			if err != nil {
				return fmt.Errorf("failed to convert valoper address %s to bytes: %w", validator.OperatorAddress, err)
			}

			err = k.bankKeeper.SendCoinsFromModuleToAccount(
				ctx,
				revenuetypes.RevenueTreasuryPoolName,
				valOperAddr,
				sdk.NewCoins(sdk.NewCoin(
					params.DenomCompensation, valCompensation,
				)),
			)
			if err != nil {
				return fmt.Errorf("failed to send revenue to validator %s: %w", validator.OperatorAddress, err)
			}
		}

		info.CommitedBlocksInPeriod = 0
		info.CommitedOracleVotesInPeriod = 0
		if err := k.SetValidatorInfo(ctx, valConsAddr, info); err != nil {
			return fmt.Errorf("failed to reset a validator info: %w", err)
		}
	}
	return nil
}

// CalcBaseRevenueAmount calculates the base compensation amount for validators based on the current
// price of the compensation denomination. The final compensation amount for a validator is
// determined by multiplying the base revenue amount by the validator's performance rating.
func (k *Keeper) CalcBaseRevenueAmount(_ sdk.Context) int64 {
	// TODO: implement calculation of base compensation
	// TODO: think about price obsolescence case (if the price is too old, should we use it for
	// payments?)
	return 10_000_000
}

func (k *Keeper) getOrCreateValidatorInfo(
	ctx sdk.Context,
	addr sdk.ConsAddress,
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
		ConsensusAddress: addr.String(),
	}
	if err := k.SetValidatorInfo(ctx, addr, info); err != nil {
		return info, fmt.Errorf("failed to write validator info to the store: %w", err)
	}
	k.Logger(ctx).Debug("new validator info created", "info", info)
	return info, nil
}
