package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v5/utils"
	"github.com/neutron-org/neutron/v5/x/dex/types"
)

// DepositCore handles core logic for MsgDeposit including bank operations and event emissions
func (k Keeper) DepositCore(
	goCtx context.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	amounts0 []math.Int,
	amounts1 []math.Int,
	tickIndices []int64,
	fees []uint64,
	options []*types.DepositOptions,
) (amounts0Deposit, amounts1Deposit []math.Int, sharesIssued sdk.Coins, failedDeposits []*types.FailedDeposit, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amounts0Deposited,
		amounts1Deposited,
		totalAmountReserve0,
		totalAmountReserve1,
		sharesIssued,
		events,
		failedDeposits,
		err := k.ExecuteDeposit(ctx, pairID, callerAddr, receiverAddr, amounts0, amounts1, tickIndices, fees, options)
	if err != nil {
		return nil, nil, nil, failedDeposits, err
	}

	ctx.EventManager().EmitEvents(events)

	if totalAmountReserve0.IsPositive() {
		coin0 := sdk.NewCoin(pairID.Token0, totalAmountReserve0)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, callerAddr, types.ModuleName, sdk.Coins{coin0}); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	if totalAmountReserve1.IsPositive() {
		coin1 := sdk.NewCoin(pairID.Token1, totalAmountReserve1)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, callerAddr, types.ModuleName, sdk.Coins{coin1}); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	if err := k.MintShares(ctx, receiverAddr, sharesIssued); err != nil {
		return nil, nil, nil, nil, err
	}

	return amounts0Deposited, amounts1Deposited, sharesIssued, failedDeposits, nil
}

// ExecuteDeposit handles core logic for deposits -- checking and initializing data structures (tick, pair), calculating
// shares based on amount deposited. IT DOES NOT PERFORM ANY BANKING OPERATIONS.
func (k Keeper) ExecuteDeposit(
	ctx sdk.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	amounts0 []math.Int,
	amounts1 []math.Int,
	tickIndices []int64,
	fees []uint64,
	options []*types.DepositOptions) (
	amounts0Deposited, amounts1Deposited []math.Int,
	totalAmountReserve0, totalAmountReserve1 math.Int,
	sharesIssued sdk.Coins,
	events sdk.Events,
	failedDeposits []*types.FailedDeposit,
	err error,
) {
	totalAmountReserve0 = math.ZeroInt()
	totalAmountReserve1 = math.ZeroInt()
	amounts0Deposited = make([]math.Int, len(amounts0))
	amounts1Deposited = make([]math.Int, len(amounts1))
	sharesIssued = sdk.Coins{}

	for i := 0; i < len(amounts0); i++ {
		amounts0Deposited[i] = math.ZeroInt()
		amounts1Deposited[i] = math.ZeroInt()
	}

	for i, amount0 := range amounts0 {
		amount1 := amounts1[i]
		tickIndex := tickIndices[i]
		fee := fees[i]
		option := options[i]
		if option == nil {
			option = &types.DepositOptions{}
		}
		autoswap := !option.DisableAutoswap

		if err := k.ValidateFee(ctx, fee); err != nil {
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, err
		}

		if k.IsPoolBehindEnemyLines(ctx, pairID, tickIndex, fee, amount0, amount1) {
			err = sdkerrors.Wrapf(types.ErrDepositBehindEnemyLines,
				"deposit failed at tick %d fee %d", tickIndex, fee)
			if option.FailTxOnBel {
				return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, err
			}
			failedDeposits = append(failedDeposits, &types.FailedDeposit{DepositIdx: uint64(i), Error: err.Error()})
			continue
		}

		pool, err := k.GetOrInitPool(
			ctx,
			pairID,
			tickIndex,
			fee,
		)
		if err != nil {
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, err
		}

		existingShares := k.bankKeeper.GetSupply(ctx, pool.GetPoolDenom()).Amount

		inAmount0, inAmount1, outShares := pool.Deposit(amount0, amount1, existingShares, autoswap)

		k.SetPool(ctx, pool)

		if inAmount0.IsZero() && inAmount1.IsZero() {
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, types.ErrZeroTrueDeposit
		}

		if outShares.IsZero() {
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, types.ErrDepositShareUnderflow
		}

		sharesIssued = append(sharesIssued, outShares)

		amounts0Deposited[i] = inAmount0
		amounts1Deposited[i] = inAmount1
		totalAmountReserve0 = totalAmountReserve0.Add(inAmount0)
		totalAmountReserve1 = totalAmountReserve1.Add(inAmount1)

		depositEvent := types.CreateDepositEvent(
			callerAddr,
			receiverAddr,
			pairID.Token0,
			pairID.Token1,
			tickIndex,
			fee,
			inAmount0,
			inAmount1,
			outShares.Amount,
		)
		events = append(events, depositEvent)
	}

	// At this point shares issued is not sorted and may have duplicates
	// we must sanitize to convert it to a valid set of coins
	sharesIssued = utils.SanitizeCoins(sharesIssued)
	return amounts0Deposited, amounts1Deposited, totalAmountReserve0, totalAmountReserve1, sharesIssued, events, failedDeposits, nil
}
