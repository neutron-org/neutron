package keeper

import (
	"cosmossdk.io/core/comet"
	coretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/skip-mev/slinky/abci/strategies/aggregator"
)

type Keeper struct {
	cdc           codec.BinaryCodec
	storeService  coretypes.KVStoreService
	va            aggregator.VoteAggregator
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    bankkeeper.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService coretypes.KVStoreService,
	va aggregator.VoteAggregator,
	stakingKeeper *stakingkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) *Keeper {
	// ensure bonded and not bonded module accounts are set
	//if addr := ak.GetModuleAddress(types.BondedPoolName); addr == nil {
	//	panic(fmt.Sprintf("%s module account has not been set", types.BondedPoolName))
	//}
	//
	//if addr := ak.GetModuleAddress(types.NotBondedPoolName); addr == nil {
	//	panic(fmt.Sprintf("%s module account has not been set", types.NotBondedPoolName))
	//}
	//
	//// ensure that authority is a valid AccAddress
	//if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
	//	panic("authority is not a valid acc address")
	//}
	return &Keeper{
		cdc:           cdc,
		storeService:  storeService,
		va:            va,
		stakingKeeper: stakingKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) GetState(ctx sdk.Context) (state types.State, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.StateKey)
	if err != nil {
		return state, err
	}
	if bz == nil {
		return state, nil
	}

	err = k.cdc.Unmarshal(bz, &state)
	return state, err
}

func (k *Keeper) SetState(ctx sdk.Context, state types.State) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&state)
	if err != nil {
		return err
	}
	err = store.Set(types.StateKey, bz)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keeper) GetAllValidatorInfo(ctx sdk.Context) (infos []types.ValidatorInfo, err error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.PrefixValidatorInfoKey, storetypes.PrefixEndBytes(types.PrefixValidatorInfoKey))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	infos = make([]types.ValidatorInfo, 0)
	for ; iter.Valid(); iter.Next() {
		var info types.ValidatorInfo
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
	blocksPassed uint64,
) (info types.ValidatorInfo, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetValidatorInfoKey(addr))
	if err != nil {
		return info, err
	}
	if bz == nil {
		stakingVal, err := k.stakingKeeper.GetValidatorByConsAddr(ctx, addr)
		if err != nil {
			// TODO: should actually never happen, but try fill OperatorAddress later if stakingVal not found
			k.Logger(ctx).Error(err.Error())
		}

		info = types.ValidatorInfo{
			// GetOperator might return empty string if validator in staking module not found by ConsAddress
			OperatorAddress:          stakingVal.GetOperator(),
			ConsensusAddress:         addr.String(),
			MissedBlocksInMonth:      blocksPassed,
			MissedOracleVotesInMonth: blocksPassed,
		}
		infoBz, err := k.cdc.Marshal(&info)
		if err != nil {
			return info, err
		}
		err = store.Set(types.GetValidatorInfoKey(addr), infoBz)
		if err != nil {
			return info, err
		}
		k.Logger(ctx).Debug("new validator info created", "info", info)
		return info, nil
	}
	err = k.cdc.Unmarshal(bz, &info)
	return info, err
}

func (k *Keeper) SetValidatorInfo(ctx sdk.Context, addr sdk.ConsAddress, info types.ValidatorInfo) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&info)
	if err != nil {
		return nil
	}
	err = store.Set(types.GetValidatorInfoKey(addr), bz)
	return err
}

func (k *Keeper) EndBlocker(ctx sdk.Context) error {
	state, err := k.GetState(ctx)
	if err != nil {
		return err
	}
	// if the first block of the next month
	if state.WorkingMonth != int32(ctx.BlockTime().Month()) {
		// TODO: pause revenue processing in case if any error during endblocker ???
		k.ProcessRevenue(ctx)
		k.ResetValidators(ctx)
		state.WorkingMonth = int32(ctx.BlockTime().Month())
		state.BlockCounter = 0

	}

	err = k.ProcessSignatures(ctx, state.BlockCounter)
	if err != nil {
		return err
	}

	err = k.ProcessOracleVotes(ctx, state.BlockCounter)
	if err != nil {
		return err
	}

	state.BlockCounter++

	err = k.SetState(ctx, state)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keeper) ResetValidators(ctx sdk.Context) {
	// TODO:
}

func (k *Keeper) ProcessSignatures(ctx sdk.Context, blocksProgress uint64) error {
	for _, info := range ctx.VoteInfos() {
		if comet.BlockIDFlag(info.BlockIdFlag) == comet.BlockIDFlagAbsent {
			// missed
			k.Logger(ctx).Debug("missed signature", "validator", info.Validator.Address)

			valInfo, err := k.getOrCreateValidatorInfo(ctx, info.Validator.Address, blocksProgress)
			if err != nil {
				return err
			}

			valInfo.MissedBlocksInMonth++

			err = k.SetValidatorInfo(ctx, info.Validator.Address, valInfo)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (k *Keeper) ProcessOracleVotes(ctx sdk.Context, blocksProgress uint64) error {
	for _, info := range ctx.VoteInfos() {
		addr := sdk.ConsAddress(info.Validator.Address)
		prices := k.va.GetPriceForValidator(addr)
		if len(prices) == 0 {
			//missed oracle
			k.Logger(ctx).Debug("missed oracle vote", "validator", info.Validator.Address)

			valInfo, err := k.getOrCreateValidatorInfo(ctx, info.Validator.Address, blocksProgress)
			if err != nil {
				return err
			}

			valInfo.MissedOracleVotesInMonth++

			err = k.SetValidatorInfo(ctx, info.Validator.Address, valInfo)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (k *Keeper) ProcessRevenue(ctx sdk.Context) error {
	infos, err := k.GetAllValidatorInfo(ctx)
	if err != nil {
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
			info.GetMissedBlocksInMonth(),
			info.GetMissedOracleVotesInMonth(),
			state.BlockCounter,
		)
		valCompensation := rating.MulInt64(baseCompensation).TruncateInt()
		_, addr, err := bech32.DecodeAndConvert(info.OperatorAddress)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
		}
		err = k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.RevenueTreasuryPoolName,
			addr,
			sdk.NewCoins(sdk.NewCoin(
				params.DenomCompensation, valCompensation,
			)),
		)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
		}
	}
	return nil
}

func (k *Keeper) GetBaseNTRNAmount(ctx sdk.Context) int64 {
	// TODO: implement calculation of base compensation
	return 10_000_000
}
