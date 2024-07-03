package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/utils"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/types"
)

// NOTE: Currently we are using TruncateInt in multiple places for converting Decs back into math.Ints.
// This may create some accounting anomalies but seems preferable to other alternatives.
// See full ADR here: https://www.notion.so/dualityxyz/A-Modest-Proposal-For-Truncating-696a919d59254876a617f82fb9567895

// Handles core logic for MsgDeposit, checking and initializing data structures (tick, pair), calculating
// shares based on amount deposited, and sending funds to moduleAddress.
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
		failedDeposits,
		err := k.CalculateDeposit(ctx, pairID, callerAddr, receiverAddr, amounts0, amounts1, tickIndices, fees, options)
	if err != nil {
		return nil, nil, nil, failedDeposits, err
	}

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

func (k Keeper) CalculateDeposit(
	ctx sdk.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	amounts0 []math.Int,
	amounts1 []math.Int,
	tickIndices []int64,
	fees []uint64,
	options []*types.DepositOptions) (amounts0Deposit, amounts1Deposit []math.Int, totalAmountReserve0, totalAmountReserve1 math.Int, sharesIssued sdk.Coins, failedDeposits []*types.FailedDeposit, err error) {
	totalAmountReserve0 = math.ZeroInt()
	totalAmountReserve1 = math.ZeroInt()
	amounts0Deposited := make([]math.Int, len(amounts0))
	amounts1Deposited := make([]math.Int, len(amounts1))
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
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, err
		}

		if k.IsPoolBehindEnemyLines(ctx, pairID, tickIndex, fee, amount0, amount1) {
			err = sdkerrors.Wrapf(types.ErrDepositBehindEnemyLines,
				"deposit failed at tick %d fee %d", tickIndex, fee)
			if option.FailTxOnBel {
				return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, err
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
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, err
		}

		existingShares := k.bankKeeper.GetSupply(ctx, pool.GetPoolDenom()).Amount

		inAmount0, inAmount1, outShares := pool.Deposit(amount0, amount1, existingShares, autoswap)

		k.SetPool(ctx, pool)

		if inAmount0.IsZero() && inAmount1.IsZero() {
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, types.ErrZeroTrueDeposit
		}

		if outShares.IsZero() {

			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, types.ErrDepositShareUnderflow
		}

		sharesIssued = append(sharesIssued, outShares)

		amounts0Deposited[i] = inAmount0
		amounts1Deposited[i] = inAmount1
		totalAmountReserve0 = totalAmountReserve0.Add(inAmount0)
		totalAmountReserve1 = totalAmountReserve1.Add(inAmount1)

		//TODO: probably don't emit events here
		ctx.EventManager().EmitEvent(types.CreateDepositEvent(
			callerAddr,
			receiverAddr,
			pairID.Token0,
			pairID.Token1,
			tickIndex,
			fee,
			inAmount0,
			inAmount1,
			outShares.Amount,
		))
	}

	// At this point shares issued is not sorted and may have duplicates
	// we must sanitize to convert it to a valid set of coins
	sharesIssued = utils.SanitizeCoins(sharesIssued)
	return amounts0Deposit, amounts1Deposit, totalAmountReserve0, totalAmountReserve1, sharesIssued, failedDeposits, nil
}

// Handles core logic for MsgWithdrawal; calculating and withdrawing reserve0,reserve1 from a specified tick
// given a specified number of shares to remove.
// Calculates the amount of reserve0, reserve1 to withdraw based on the percentage of the desired
// number of shares to remove compared to the total number of shares at the given tick.
func (k Keeper) WithdrawCore(
	goCtx context.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	sharesToRemoveList []math.Int,
	tickIndicesNormalized []int64,
	fees []uint64,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	totalReserve0ToRemove, totalReserve1ToRemove, events, err := k.CalculateWithdraw(ctx, pairID, callerAddr, receiverAddr, sharesToRemoveList, tickIndicesNormalized, fees)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvents(events)

	if totalReserve0ToRemove.IsPositive() {
		coin0 := sdk.NewCoin(pairID.Token0, totalReserve0ToRemove)

		err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			receiverAddr,
			sdk.Coins{coin0},
		)
		ctx.EventManager().EmitEvents(types.GetEventsWithdrawnAmount(sdk.Coins{coin0}))
		if err != nil {
			return err
		}
	}

	// sends totalReserve1ToRemove to receiverAddr
	if totalReserve1ToRemove.IsPositive() {
		coin1 := sdk.NewCoin(pairID.Token1, totalReserve1ToRemove)
		err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			receiverAddr,
			sdk.Coins{coin1},
		)
		ctx.EventManager().EmitEvents(types.GetEventsWithdrawnAmount(sdk.Coins{coin1}))
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) CalculateWithdraw(
	ctx sdk.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	sharesToRemoveList []math.Int,
	tickIndicesNormalized []int64,
	fees []uint64,
) (totalReserves0ToRemove, totalReserves1ToRemove math.Int, events sdk.Events, err error) {
	totalReserve0ToRemove := math.ZeroInt()
	totalReserve1ToRemove := math.ZeroInt()

	for i, fee := range fees {
		sharesToRemove := sharesToRemoveList[i]
		tickIndex := tickIndicesNormalized[i]

		pool, err := k.GetOrInitPool(ctx, pairID, tickIndex, fee)
		if err != nil {
			return math.ZeroInt(), math.ZeroInt(), nil, err
		}

		poolDenom := pool.GetPoolDenom()

		totalShares := k.bankKeeper.GetSupply(ctx, poolDenom).Amount
		if totalShares.LT(sharesToRemove) {
			return math.ZeroInt(), math.ZeroInt(), nil, sdkerrors.Wrapf(
				types.ErrInsufficientShares,
				"%s does not have %s shares of type %s",
				callerAddr,
				sharesToRemove,
				poolDenom,
			)
		}

		outAmount0, outAmount1 := pool.Withdraw(sharesToRemove, totalShares)
		k.SetPool(ctx, pool)

		if sharesToRemove.IsPositive() {
			if err := k.BurnShares(ctx, callerAddr, sharesToRemove, poolDenom); err != nil {
				return math.ZeroInt(), math.ZeroInt(), nil, err
			}
		}

		totalReserve0ToRemove = totalReserve0ToRemove.Add(outAmount0)
		totalReserve1ToRemove = totalReserve1ToRemove.Add(outAmount1)

		event := types.CreateWithdrawEvent(
			callerAddr,
			receiverAddr,
			pairID.Token0,
			pairID.Token1,
			tickIndex,
			fee,
			outAmount0,
			outAmount1,
			sharesToRemove,
		)
		events = append(events, event)
	}
	return totalReserve0ToRemove, totalReserve1ToRemove, events, nil
}

func (k Keeper) MultiHopSwapCore(
	goCtx context.Context,
	amountIn math.Int,
	routes []*types.MultiHopRoute,
	exitLimitPrice math_utils.PrecDec,
	pickBestRoute bool,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
) (coinOut sdk.Coin, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	bestRoute, initialInCoin, err := k.CalculateMultiHopSwap(ctx, amountIn, routes, exitLimitPrice, pickBestRoute, callerAddr, receiverAddr)
	if err != nil {
		return sdk.Coin{}, err
	}

	bestRoute.write()
	err = k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		callerAddr,
		types.ModuleName,
		sdk.Coins{initialInCoin},
	)
	if err != nil {
		return sdk.Coin{}, err
	}

	// send both dust and coinOut to receiver
	// note that dust can be multiple coins collected from multiple hops.
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		receiverAddr,
		bestRoute.dust.Add(bestRoute.coinOut),
	)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to send out coin and dust to the receiver: %w", err)
	}

	ctx.EventManager().EmitEvent(types.CreateMultihopSwapEvent(
		callerAddr,
		receiverAddr,
		initialInCoin.Denom,
		bestRoute.coinOut.Denom,
		initialInCoin.Amount,
		bestRoute.coinOut.Amount,
		bestRoute.route,
		bestRoute.dust,
	))

	return bestRoute.coinOut, nil
}

func (k Keeper) CalculateMultiHopSwap(
	ctx sdk.Context,
	amountIn math.Int,
	routes []*types.MultiHopRoute,
	exitLimitPrice math_utils.PrecDec,
	pickBestRoute bool,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
) (bestRoute routeOutput, initialInCoin sdk.Coin, err error) {
	var routeErrors []error
	initialInCoin = sdk.NewCoin(routes[0].Hops[0], amountIn)
	stepCache := make(map[multihopCacheKey]StepResult)

	bestRoute.coinOut = sdk.Coin{Amount: math.ZeroInt()}

	for _, route := range routes {
		routeDust, routeCoinOut, writeRoute, err := k.RunMultihopRoute(
			ctx,
			*route,
			initialInCoin,
			exitLimitPrice,
			stepCache,
		)
		if err != nil {
			routeErrors = append(routeErrors, err)
			continue
		}

		if !pickBestRoute || bestRoute.coinOut.Amount.LT(routeCoinOut.Amount) {
			bestRoute.coinOut = routeCoinOut
			bestRoute.write = writeRoute
			bestRoute.route = route.Hops
			bestRoute.dust = routeDust
		}
		if !pickBestRoute {
			break
		}
	}

	if len(routeErrors) == len(routes) {
		// All routes have failed

		allErr := errors.Join(append([]error{types.ErrAllMultiHopRoutesFailed}, routeErrors...)...)

		return routeOutput{}, sdk.Coin{}, allErr
	}

	return bestRoute, initialInCoin, nil
}

// PlaceLimitOrderCore handles MsgPlaceLimitOrder, initializing (tick, pair) data structures if needed, calculating and
// storing information for a new limit order at a specific tick.
func (k Keeper) PlaceLimitOrderCore(
	goCtx context.Context,
	tokenIn string,
	tokenOut string,
	amountIn math.Int,
	tickIndexInToOut int64,
	orderType types.LimitOrderType,
	goodTil *time.Time,
	maxAmountOut *math.Int,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
) (trancheKey string, totalInCoin, swapInCoin, swapOutCoin sdk.Coin, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	takerTradePairID, err := types.NewTradePairID(tokenIn, tokenOut)
	if err != nil {
		return trancheKey, totalInCoin, swapInCoin, swapOutCoin, err
	}
	trancheKey, totalIn, swapInCoin, swapOutCoin, sharesIssued, err := k.CalculatePlaceLimitOrder(ctx, takerTradePairID, amountIn, tickIndexInToOut, orderType, goodTil, maxAmountOut, receiverAddr)
	if err != nil {
		return trancheKey, totalInCoin, swapInCoin, swapOutCoin, err
	}

	if swapOutCoin.IsPositive() {
		err = k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			receiverAddr,
			sdk.Coins{swapOutCoin},
		)
		if err != nil {
			return trancheKey, totalInCoin, swapInCoin, swapOutCoin, err
		}
	}

	if totalIn.IsPositive() {
		totalInCoin = sdk.NewCoin(tokenIn, totalIn)

		err = k.bankKeeper.SendCoinsFromAccountToModule(
			ctx,
			callerAddr,
			types.ModuleName,
			sdk.Coins{totalInCoin},
		)
		if err != nil {
			return trancheKey, totalInCoin, swapInCoin, swapOutCoin, err
		}
	}

	// This is ok because we've already successfully constructed a TradePairID above
	pairID := takerTradePairID.MustPairID()
	ctx.EventManager().EmitEvent(types.CreatePlaceLimitOrderEvent(
		callerAddr,
		receiverAddr,
		pairID.Token0,
		pairID.Token1,
		tokenIn,
		tokenOut,
		totalIn,
		tickIndexInToOut,
		orderType.String(),
		sharesIssued,
		trancheKey,
	))

	return trancheKey, totalInCoin, swapInCoin, swapOutCoin, nil
}

func (k Keeper) CalculatePlaceLimitOrder(
	ctx sdk.Context,
	takerTradePairID *types.TradePairID,
	amountIn math.Int,
	tickIndexInToOut int64,
	orderType types.LimitOrderType,
	goodTil *time.Time,
	maxAmountOut *math.Int,
	receiverAddr sdk.AccAddress,
) (trancheKey string, totalIn math.Int, swapInCoin, swapOutCoin sdk.Coin, sharesIssued math.Int, err error) {

	amountLeft := amountIn

	var limitPrice math_utils.PrecDec
	limitPrice, err = types.CalcPrice(tickIndexInToOut)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
	}

	// Ensure that after rounding user will get at least 1 token out.
	err = types.ValidateFairOutput(amountIn, limitPrice)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
	}

	var orderFilled bool
	if orderType.IsTakerOnly() {
		swapInCoin, swapOutCoin, err = k.TakerLimitOrderSwap(ctx, *takerTradePairID, amountIn, maxAmountOut, limitPrice, orderType)
	} else {
		swapInCoin, swapOutCoin, orderFilled, err = k.MakerLimitOrderSwap(ctx, *takerTradePairID, amountIn, limitPrice)
	}
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
	}

	totalIn = swapInCoin.Amount
	amountLeft = amountLeft.Sub(swapInCoin.Amount)

	makerTradePairID := takerTradePairID.Reversed()
	makerTickIndexTakerToMaker := tickIndexInToOut * -1
	var placeTranche *types.LimitOrderTranche
	placeTranche, err = k.GetOrInitPlaceTranche(
		ctx,
		makerTradePairID,
		makerTickIndexTakerToMaker,
		goodTil,
		orderType,
	)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
	}

	trancheKey = placeTranche.Key.TrancheKey
	trancheUser := k.GetOrInitLimitOrderTrancheUser(
		ctx,
		makerTradePairID,
		makerTickIndexTakerToMaker,
		trancheKey,
		orderType,
		receiverAddr.String(),
	)

	// FOR GTC, JIT & GoodTil try to place a maker limitOrder with remaining Amount
	if amountLeft.IsPositive() && !orderFilled &&
		(orderType.IsGTC() || orderType.IsJIT() || orderType.IsGoodTil()) {

		// Ensure that the maker portion will generate at least 1 token of output
		// NOTE: This does mean that a successful taker leg of the trade will be thrown away since the entire tx will fail.
		// In most circumstances this seems preferable to executing the taker leg and exiting early before placing a maker
		// order with the remaining liquidity.
		err = types.ValidateFairOutput(amountLeft, limitPrice)
		if err != nil {
			return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
		}
		placeTranche.PlaceMakerLimitOrder(amountLeft)
		trancheUser.SharesOwned = trancheUser.SharesOwned.Add(amountLeft)

		if orderType.HasExpiration() {
			goodTilRecord := NewLimitOrderExpiration(placeTranche)
			k.SetLimitOrderExpiration(ctx, goodTilRecord)
			ctx.GasMeter().ConsumeGas(types.ExpiringLimitOrderGas, "Expiring LimitOrder Fee")
		}

		k.SaveTranche(ctx, placeTranche)

		totalIn = totalIn.Add(amountLeft)
		sharesIssued = amountLeft
	}

	k.SaveTrancheUser(ctx, trancheUser)

	if orderType.IsJIT() {
		err = k.AssertCanPlaceJIT(ctx)
		if err != nil {
			return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
		}
		k.IncrementJITsInBlock(ctx)
	}

	return trancheKey, totalIn, swapInCoin, swapOutCoin, sharesIssued, nil
}

// CancelLimitOrderCore handles MsgCancelLimitOrder, removing a specified number of shares from a limit order
// and returning the respective amount in terms of the reserve to the user.
func (k Keeper) CancelLimitOrderCore(
	goCtx context.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coinOut, tradePairID, err := k.CalculateCancelLimitOrder(ctx, trancheKey, callerAddr)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		callerAddr,
		sdk.Coins{coinOut},
	)
	if err != nil {
		return err
	}

	pairID := tradePairID.MustPairID()
	ctx.EventManager().EmitEvent(types.CancelLimitOrderEvent(
		callerAddr,
		pairID.Token0,
		pairID.Token1,
		tradePairID.MakerDenom,
		tradePairID.TakerDenom,
		coinOut.Amount,
		trancheKey,
	))

	return nil
}

func (k Keeper) CalculateCancelLimitOrder(
	ctx sdk.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) (coinOut sdk.Coin, tradePairID *types.TradePairID, error error) {
	trancheUser, found := k.GetLimitOrderTrancheUser(ctx, callerAddr.String(), trancheKey)
	if !found {
		return sdk.Coin{}, nil, types.ErrActiveLimitOrderNotFound
	}

	tradePairID, tickIndex := trancheUser.TradePairId, trancheUser.TickIndexTakerToMaker
	tranche := k.GetLimitOrderTranche(
		ctx,
		&types.LimitOrderTrancheKey{
			TradePairId:           tradePairID,
			TickIndexTakerToMaker: tickIndex,
			TrancheKey:            trancheKey,
		},
	)
	if tranche == nil {
		return sdk.Coin{}, nil, types.ErrActiveLimitOrderNotFound
	}

	amountToCancel := tranche.RemoveTokenIn(trancheUser)
	trancheUser.SharesCancelled = trancheUser.SharesCancelled.Add(amountToCancel)

	if !amountToCancel.IsPositive() {
		return sdk.Coin{}, nil, sdkerrors.Wrapf(types.ErrCancelEmptyLimitOrder, "%s", tranche.Key.TrancheKey)
	}

	k.SaveTrancheUser(ctx, trancheUser)
	k.SaveTranche(ctx, tranche)

	if trancheUser.OrderType.HasExpiration() {
		k.RemoveLimitOrderExpiration(ctx, *tranche.ExpirationTime, tranche.Key.KeyMarshal())
	}
	coinOut = sdk.NewCoin(tradePairID.MakerDenom, amountToCancel)

	return coinOut, tradePairID, nil

}

// WithdrawFilledLimitOrderCore handles MsgWithdrawFilledLimitOrder, calculates and sends filled liquidity from module to user
// for a limit order based on amount wished to receive.
func (k Keeper) WithdrawFilledLimitOrderCore(
	goCtx context.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amountOutTokenOut, remainingTokenIn, tradePairID, err := k.CalculateFilledLimitOrderCore(ctx, trancheKey, callerAddr)
	if err != nil {
		return err
	}

	coinTakerDenomOut := sdk.NewCoin(tradePairID.TakerDenom, amountOutTokenOut)
	coinMakerDenomRefund := sdk.NewCoin(tradePairID.MakerDenom, remainingTokenIn)
	coins := sdk.NewCoins(coinTakerDenomOut, coinMakerDenomRefund)
	ctx.EventManager().EmitEvents(types.GetEventsWithdrawnAmount(sdk.NewCoins(coinTakerDenomOut)))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, callerAddr, coins); err != nil {
		return err
	}

	// tradePairID has already been constructed so this will not error
	pairID := tradePairID.MustPairID()
	ctx.EventManager().EmitEvent(types.WithdrawFilledLimitOrderEvent(
		callerAddr,
		pairID.Token0,
		pairID.Token1,
		tradePairID.MakerDenom,
		tradePairID.TakerDenom,
		amountOutTokenOut,
		trancheKey,
	))

	return nil
}

func (k Keeper) CalculateFilledLimitOrderCore(
	ctx sdk.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) (amountOutTokenOut, remainingTokenIn math.Int, tradePairID *types.TradePairID, err error) {
	trancheUser, found := k.GetLimitOrderTrancheUser(
		ctx,
		callerAddr.String(),
		trancheKey,
	)
	if !found {
		return math.ZeroInt(), math.ZeroInt(), nil, sdkerrors.Wrapf(types.ErrValidLimitOrderTrancheNotFound, "%s", trancheKey)
	}

	tradePairID, tickIndex := trancheUser.TradePairId, trancheUser.TickIndexTakerToMaker

	tranche, wasFilled, found := k.FindLimitOrderTranche(
		ctx,
		&types.LimitOrderTrancheKey{
			TradePairId:           tradePairID,
			TickIndexTakerToMaker: tickIndex,
			TrancheKey:            trancheKey,
		},
	)

	amountOutTokenOut = math.ZeroInt()
	remainingTokenIn = math.ZeroInt()
	// It's possible that a TrancheUser exists but tranche does not if LO was filled entirely through a swap
	if found {
		var amountOutTokenIn math.Int
		amountOutTokenIn, amountOutTokenOut = tranche.Withdraw(trancheUser)

		if wasFilled {
			// This is only relevant for inactive JIT and GoodTil limit orders
			remainingTokenIn = tranche.RemoveTokenIn(trancheUser)
			k.SaveInactiveTranche(ctx, tranche)

			// Treat the removed tokenIn as cancelled shares
			trancheUser.SharesCancelled = trancheUser.SharesCancelled.Add(remainingTokenIn)

		} else {
			k.SetLimitOrderTranche(ctx, tranche)
		}

		trancheUser.SharesWithdrawn = trancheUser.SharesWithdrawn.Add(amountOutTokenIn)
	}

	k.SaveTrancheUser(ctx, trancheUser)

	if !amountOutTokenOut.IsPositive() && !remainingTokenIn.IsPositive() {

		return math.ZeroInt(), math.ZeroInt(), tradePairID, types.ErrWithdrawEmptyLimitOrder
	}

	return amountOutTokenOut, remainingTokenIn, tradePairID, nil
}
