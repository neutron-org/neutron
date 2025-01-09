package keeper

import (
	"fmt"

	"cosmossdk.io/core/comet"
	coretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
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
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService coretypes.KVStoreService,
	voteAggregator revenuetypes.VoteAggregator,
	stakingKeeper revenuetypes.StakingKeeper,
	bankKeeper revenuetypes.BankKeeper,
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
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", revenuetypes.ModuleName))
}

func (k *Keeper) GetState(ctx sdk.Context) (state revenuetypes.State, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(revenuetypes.StateKey)
	if err != nil {
		return state, err
	}
	if bz == nil {
		return state, nil
	}

	err = k.cdc.Unmarshal(bz, &state)
	return state, err
}

func (k *Keeper) SetState(ctx sdk.Context, state revenuetypes.State) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&state)
	if err != nil {
		return err
	}
	err = store.Set(revenuetypes.StateKey, bz)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keeper) GetAllValidatorInfo(ctx sdk.Context) (infos []revenuetypes.ValidatorInfo, err error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(revenuetypes.PrefixValidatorInfoKey, storetypes.PrefixEndBytes(revenuetypes.PrefixValidatorInfoKey))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	infos = make([]revenuetypes.ValidatorInfo, 0)
	for ; iter.Valid(); iter.Next() {
		var info revenuetypes.ValidatorInfo
		err = k.cdc.Unmarshal(iter.Value(), &info)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (k *Keeper) getOrCreateValidatorInfo(
	ctx sdk.Context,
	addr sdk.ConsAddress,
) (info revenuetypes.ValidatorInfo, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(revenuetypes.GetValidatorInfoKey(addr))
	if err != nil {
		return info, err
	}
	if bz == nil {
		stakingVal, err := k.stakingKeeper.GetValidatorByConsAddr(ctx, addr)
		if err != nil {
			// TODO: should actually never happen, but try fill OperatorAddress later if stakingVal not found
			k.Logger(ctx).Error(err.Error())
			// TODO: handle error
		}

		info = revenuetypes.ValidatorInfo{
			// GetOperator might return empty string if validator in staking module not found by ConsAddress
			OperatorAddress:  stakingVal.GetOperator(),
			ConsensusAddress: addr.String(),
		}
		infoBz, err := k.cdc.Marshal(&info)
		if err != nil {
			return info, err
		}
		err = store.Set(revenuetypes.GetValidatorInfoKey(addr), infoBz)
		if err != nil {
			return info, err
		}
		k.Logger(ctx).Debug("new validator info created", "info", info)
		return info, nil
	}
	err = k.cdc.Unmarshal(bz, &info)
	return info, err
}

func (k *Keeper) SetValidatorInfo(ctx sdk.Context, addr sdk.ConsAddress, info revenuetypes.ValidatorInfo) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&info)
	if err != nil {
		return nil
	}
	err = store.Set(revenuetypes.GetValidatorInfoKey(addr), bz)
	return err
}

// TODO: cognizant actions on EACH possible error case in EndBlock
func (k *Keeper) EndBlock(ctx sdk.Context) error {
	state, err := k.GetState(ctx)
	if err != nil {
		return err
	}

	// if the current month has changed
	if state.CurrentMonth != int32(ctx.BlockTime().Month()) {
		// TODO: pause revenue processing in case if any error during EndBlock ???
		k.ProcessRevenue(ctx)
		k.ResetValidators(ctx)
		state.CurrentMonth = int32(ctx.BlockTime().Month())
		state.BlockCounter = 0

	}

	if err := k.RecordValidatorsParticipation(ctx); err != nil {
		return err
	}

	state.BlockCounter++
	if err := k.SetState(ctx, state); err != nil {
		return err
	}
	return nil
}

func (k *Keeper) ResetValidators(ctx sdk.Context) {
	// TODO:
}

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

		// record votes in the state
		valInfo, err := k.getOrCreateValidatorInfo(ctx, info.Validator.Address)
		if err != nil {
			return err
		}
		if oracleVote {
			valInfo.CommitedOracleVotesInMonth++
		}
		if blockVote {
			valInfo.CommitedBlocksInMonth++
		}
		if err := k.SetValidatorInfo(ctx, info.Validator.Address, valInfo); err != nil {
			return err
		}
	}
	return nil
}

func (k *Keeper) ProcessRevenue(ctx sdk.Context) error {
	infos, err := k.GetAllValidatorInfo(ctx)
	if err != nil {
		// TODO: better errors msgs
		return err
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	state, err := k.GetState(ctx)
	if err != nil {
		return err
	}
	baseCompensation := k.GetBaseNTRNAmount(ctx)
	for _, info := range infos {
		rating := PerformanceRating(
			params,
			state.BlockCounter-info.GetCommitedBlocksInMonth(),
			state.BlockCounter-info.GetCommitedOracleVotesInMonth(),
			state.BlockCounter,
		)
		valCompensation := rating.MulInt64(baseCompensation).TruncateInt()
		_, addr, err := bech32types.DecodeAndConvert(info.OperatorAddress)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
			// TODO: handle error
		}

		if valCompensation.IsPositive() {
			err = k.bankKeeper.SendCoinsFromModuleToAccount(
				ctx,
				revenuetypes.RevenueTreasuryPoolName,
				addr,
				sdk.NewCoins(sdk.NewCoin(
					params.DenomCompensation, valCompensation,
				)),
			)
			if err != nil {
				k.Logger(ctx).Error(err.Error())
				// TODO: handle error
			}
		}
	}
	return nil
}

func (k *Keeper) GetBaseNTRNAmount(ctx sdk.Context) int64 {
	// TODO: implement calculation of base compensation
	// TODO: think about price obsolescence case (if the price is too old, should we use it for
	// payments?)
	return 10_000_000
}
