package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	"github.com/neutron-org/neutron/v7/x/dex/types"
)

// WithdrawCore handles logic for MsgWithdrawal including bank operations and event emissions.
func (k Keeper) WithdrawCore(
	goCtx context.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	sharesToRemoveList []math.Int,
	tickIndicesNormalized []int64,
	fees []uint64,
) (reserves0ToRemoved, reserves1ToRemoved math_utils.PrecDec, sharesBurned sdk.Coins, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	totalReserve0ToRemove, totalReserve1ToRemove, coinsToBurn, events, err := k.ExecuteWithdraw(
		ctx,
		pairID,
		callerAddr,
		receiverAddr,
		sharesToRemoveList,
		tickIndicesNormalized,
		fees,
	)
	if err != nil {
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, err
	}

	ctx.EventManager().EmitEvents(events)

	if err := k.BurnShares(ctx, callerAddr, coinsToBurn); err != nil {
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, err
	}

	coin0 := types.NewPrecDecCoin(pairID.Token0, totalReserve0ToRemove)
	if totalReserve0ToRemove.IsPositive() {
		ctx.EventManager().EmitEvents(types.GetEventsWithdrawnAmount(sdk.Coins{coin0.TruncateToCoin()}))
	}

	coin1 := types.NewPrecDecCoin(pairID.Token1, totalReserve1ToRemove)
	if totalReserve1ToRemove.IsPositive() {
		ctx.EventManager().EmitEvents(types.GetEventsWithdrawnAmount(sdk.Coins{coin1.TruncateToCoin()}))
	}

	// NewPrecDecCoins will remove zero amounts
	coinsToRemove := types.NewPrecDecCoins(coin0, coin1)

	err = k.FractionalBanker.SendFractionalCoinsFromDexToAccount(
		ctx,
		receiverAddr,
		coinsToRemove,
	)
	if err != nil {
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, err
	}

	return totalReserve0ToRemove, totalReserve1ToRemove, coinsToBurn, nil
}

// ExecuteWithdraw handles the core Withdraw logic including calculating and withdrawing reserve0,reserve1 from a specified tick
// given a specified number of shares to remove.
// Calculates the amount of reserve0, reserve1 to withdraw based on the percentage of the desired
// number of shares to remove compared to the total number of shares at the given tick.
// IT DOES NOT PERFORM ANY BANKING OPERATIONS.
func (k Keeper) ExecuteWithdraw(
	ctx sdk.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	sharesToRemoveList []math.Int,
	tickIndicesNormalized []int64,
	fees []uint64,
) (totalReserves0ToRemove, totalReserves1ToRemove math_utils.PrecDec, coinsToBurn sdk.Coins, events sdk.Events, err error) {
	totalReserve0ToRemove := math_utils.ZeroPrecDec()
	totalReserve1ToRemove := math_utils.ZeroPrecDec()

	for i, fee := range fees {
		sharesToRemove := sharesToRemoveList[i]
		tickIndex := tickIndicesNormalized[i]

		pool, err := k.GetOrInitPool(ctx, pairID, tickIndex, fee)
		if err != nil {
			return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, nil, err
		}

		poolDenom := pool.GetPoolDenom()

		// TODO: this is a bit hacky. Since it is possible to have multiple withdrawals from the same pool we have to artificially update the bank balance
		// In the future we should enforce only one withdraw operation per pool in the message validation
		alreadyWithdrawnOfDenom := coinsToBurn.AmountOf(poolDenom)
		sharesOwned := k.bankKeeper.GetBalance(ctx, callerAddr, poolDenom).Amount.Sub(alreadyWithdrawnOfDenom)
		if sharesOwned.LT(sharesToRemove) {
			return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, nil, sdkerrors.Wrapf(
				types.ErrInsufficientShares,
				"%s does not have %s shares of type %s",
				callerAddr,
				sharesToRemove,
				poolDenom,
			)
		}

		totalShares := k.bankKeeper.GetSupply(ctx, poolDenom).Amount.Sub(alreadyWithdrawnOfDenom)
		outAmount0, outAmount1 := pool.Withdraw(sharesToRemove, totalShares)

		// Save both sides of the pool. If one or both sides are empty they will be deleted.
		k.UpdatePool(ctx, pool)

		totalReserve0ToRemove = totalReserve0ToRemove.Add(outAmount0)
		totalReserve1ToRemove = totalReserve1ToRemove.Add(outAmount1)

		coinsToBurn = coinsToBurn.Add(sdk.NewCoin(poolDenom, sharesToRemove))

		withdrawEvent := types.CreateWithdrawEvent(
			callerAddr,
			receiverAddr,
			pairID.Token0,
			pairID.Token1,
			tickIndex,
			fee,
			outAmount0,
			outAmount1,
			pool.Id,
			sharesToRemove,
		)
		events = append(events, withdrawEvent)
	}
	return totalReserve0ToRemove, totalReserve1ToRemove, coinsToBurn, events, nil
}
