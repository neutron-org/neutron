package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v9/utils/math"
	"github.com/neutron-org/neutron/v9/x/dex/types"
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

	poolsToRemoveFrom, err := k.PoolDataToPools(ctx, pairID, tickIndicesNormalized, fees)
	if err != nil {
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, err
	}

	return k.WithdrawHandler(ctx, callerAddr, receiverAddr, pairID, poolsToRemoveFrom, sharesToRemoveList)
}

// WithdrawWithSharesCore handles logic for MsgWithdrawalWithShares including bank operations and event emissions.
func (k Keeper) WithdrawWithSharesCore(
	goCtx context.Context,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	sharesToRemove sdk.Coins,
) (reserves0ToRemoved, reserves1ToRemoved math_utils.PrecDec, sharesBurned sdk.Coins, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	poolsToRemoveFrom, shareAmountsToRemove, err := k.SharesToPools(ctx, sharesToRemove)
	if err != nil {
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, err
	}

	return k.WithdrawHandler(ctx, callerAddr, receiverAddr, poolsToRemoveFrom[0].MustPairID(), poolsToRemoveFrom, shareAmountsToRemove)
}

// WithdrawHandler handles logic for both MsgWithdrawal and MsgWithdrawalWithShares including bank operations and event emissions.
func (k Keeper) WithdrawHandler(
	ctx sdk.Context,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	pairID *types.PairID,
	poolsToRemoveFrom []*types.Pool,
	sharesAmountsToRemove []math.Int,
) (reserves0ToRemoved, reserves1ToRemoved math_utils.PrecDec, sharesBurned sdk.Coins, err error) {
	totalReserve0ToRemove, totalReserve1ToRemove, coinsToBurn, events, err := k.ExecuteWithdraw(
		ctx,
		pairID,
		callerAddr,
		receiverAddr,
		poolsToRemoveFrom,
		sharesAmountsToRemove,
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
	poolsToRemoveFrom []*types.Pool,
	sharesAmountsToRemove []math.Int,
) (totalReserves0ToRemove, totalReserves1ToRemove math_utils.PrecDec, coinsToBurn sdk.Coins, events sdk.Events, err error) {
	totalReserve0ToRemove := math_utils.ZeroPrecDec()
	totalReserve1ToRemove := math_utils.ZeroPrecDec()

	for i, pool := range poolsToRemoveFrom {

		sharesToRemove := sharesAmountsToRemove[i]

		poolDenom := pool.GetPoolDenom()

		sharesOwned := k.bankKeeper.GetBalance(ctx, callerAddr, poolDenom).Amount

		if sharesOwned.LT(sharesToRemove) {
			return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, nil, sdkerrors.Wrapf(
				types.ErrInsufficientShares,
				"%s does not have %s shares of type %s",
				callerAddr,
				sharesToRemove,
				poolDenom,
			)
		}

		totalShares := k.bankKeeper.GetSupply(ctx, poolDenom).Amount
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
			pool.CenterTickIndexToken1(),
			pool.Fee(),
			outAmount0,
			outAmount1,
			pool.Id,
			sharesToRemove,
		)
		events = append(events, withdrawEvent)
	}
	return totalReserve0ToRemove, totalReserve1ToRemove, coinsToBurn, events, nil
}

func (k Keeper) PoolDataToPools(
	ctx sdk.Context,
	pairID *types.PairID,
	tickIndicesNormalized []int64,
	fees []uint64,
) ([]*types.Pool, error) {
	poolsToRemoveFrom := make([]*types.Pool, len(tickIndicesNormalized))
	for i, tickIndex := range tickIndicesNormalized {
		pool, found := k.GetPool(ctx, pairID, tickIndex, fees[i])
		if !found {
			return nil, sdkerrors.Wrapf(
				types.ErrPoolNotFound,
				"pool for pair %s, tick index %d, fee %d not found",
				pairID.CanonicalString(),
				tickIndex,
				fees[i],
			)
		}
		poolsToRemoveFrom[i] = pool
	}
	return poolsToRemoveFrom, nil
}

func (k Keeper) SharesToPools(
	ctx sdk.Context,
	sharesToRemove sdk.Coins,
) ([]*types.Pool, []math.Int, error) {
	shareAmountsToRemove := make([]math.Int, len(sharesToRemove))
	poolsToRemoveFrom := make([]*types.Pool, len(sharesToRemove))

	for i, share := range sharesToRemove {
		poolID, err := types.ParsePoolIDFromDenom(share.Denom)
		if err != nil {
			return nil, nil, sdkerrors.Wrapf(
				types.ErrInvalidPoolDenom,
				"invalid pool denom (%s)",
				share.Denom,
			)
		}
		pool, found := k.GetPoolByID(ctx, poolID)
		if !found {
			return nil, nil, sdkerrors.Wrapf(
				types.ErrPoolNotFound,
				"pool %d not found",
				poolID,
			)
		}

		// Safety check to ensure that all pools are from the same pair
		// this cannot be validated for MsgWithdrawalWithShares so it must be done here
		if i > 0 && !pool.MustPairID().Equal(poolsToRemoveFrom[0].MustPairID()) {
			return nil, nil, sdkerrors.Wrapf(
				types.ErrCanOnlyWithdrawFromSamePair,
				"pool %d is not part of pair %s",
				pool.Id,
				poolsToRemoveFrom[0].MustPairID().CanonicalString(),
			)
		}
		poolsToRemoveFrom[i] = pool
		shareAmountsToRemove[i] = share.Amount
	}
	return poolsToRemoveFrom, shareAmountsToRemove, nil
}
