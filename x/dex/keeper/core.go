package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	math_utils "github.com/neutron-org/neutron/utils/math"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/dex/utils"
)

// NOTE: Currently we are using TruncateInt in multiple places for converting Decs back into sdk.Ints.
// This may create some accounting anomalies but seems preferable to other alternatives.
// See full ADR here: https://www.notion.so/dualityxyz/A-Modest-Proposal-For-Truncating-696a919d59254876a617f82fb9567895

// Handles core logic for MsgDeposit, checking and initializing data structures (tick, pair), calculating
// shares based on amount deposited, and sending funds to moduleAddress.
func (k Keeper) DepositCore(
	goCtx context.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	amounts0 []sdk.Int,
	amounts1 []sdk.Int,
	tickIndices []int64,
	fees []uint64,
	options []*types.DepositOptions,
) (amounts0Deposit, amounts1Deposit []sdk.Int, sharesIssued sdk.Coins, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	totalAmountReserve0 := sdk.ZeroInt()
	totalAmountReserve1 := sdk.ZeroInt()
	amounts0Deposited := make([]sdk.Int, len(amounts0))
	amounts1Deposited := make([]sdk.Int, len(amounts1))
	sharesIssued = sdk.Coins{}

	for i := 0; i < len(amounts0); i++ {
		amounts0Deposited[i] = sdk.ZeroInt()
		amounts1Deposited[i] = sdk.ZeroInt()
	}

	for i, amount0 := range amounts0 {
		amount1 := amounts1[i]
		tickIndex := tickIndices[i]
		fee := fees[i]

		if err := k.ValidateFee(ctx, fee); err != nil {
			return nil, nil, nil, err
		}
		autoswap := !options[i].DisableAutoswap

		pool, err := k.GetOrInitPool(
			ctx,
			pairID,
			tickIndex,
			fee,
		)
		if err != nil {
			return nil, nil, nil, err
		}

		existingShares := k.bankKeeper.GetSupply(ctx, pool.GetPoolDenom()).Amount

		inAmount0, inAmount1, outShares := pool.Deposit(amount0, amount1, existingShares, autoswap)

		k.SetPool(ctx, pool)

		if inAmount0.IsZero() && inAmount1.IsZero() {
			return nil, nil, nil, types.ErrZeroTrueDeposit
		}

		if outShares.IsZero() {
			return nil, nil, nil, types.ErrDepositShareUnderflow
		}

		if err := k.MintShares(ctx, receiverAddr, outShares); err != nil {
			return nil, nil, nil, err
		}
		sharesIssued = sharesIssued.Add(outShares)

		amounts0Deposited[i] = inAmount0
		amounts1Deposited[i] = inAmount1
		totalAmountReserve0 = totalAmountReserve0.Add(inAmount0)
		totalAmountReserve1 = totalAmountReserve1.Add(inAmount1)

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

	if totalAmountReserve0.IsPositive() {
		coin0 := sdk.NewCoin(pairID.Token0, totalAmountReserve0)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, callerAddr, types.ModuleName, sdk.Coins{coin0}); err != nil {
			return nil, nil, nil, err
		}
	}

	if totalAmountReserve1.IsPositive() {
		coin1 := sdk.NewCoin(pairID.Token1, totalAmountReserve1)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, callerAddr, types.ModuleName, sdk.Coins{coin1}); err != nil {
			return nil, nil, nil, err
		}
	}

	return amounts0Deposited, amounts1Deposited, sharesIssued, nil
}

// Handles core logic for MsgWithdrawal; calculating and withdrawing reserve0,reserve1 from a specified tick
// given a specfied number of shares to remove.
// Calculates the amount of reserve0, reserve1 to withdraw based on the percentage of the desired
// number of shares to remove compared to the total number of shares at the given tick.
func (k Keeper) WithdrawCore(
	goCtx context.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	sharesToRemoveList []sdk.Int,
	tickIndicesNormalized []int64,
	fees []uint64,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	totalReserve0ToRemove := sdk.ZeroInt()
	totalReserve1ToRemove := sdk.ZeroInt()

	for i, fee := range fees {
		sharesToRemove := sharesToRemoveList[i]
		tickIndex := tickIndicesNormalized[i]

		pool, err := k.GetOrInitPool(ctx, pairID, tickIndex, fee)
		if err != nil {
			return err
		}

		poolDenom := pool.GetPoolDenom()

		totalShares := k.bankKeeper.GetSupply(ctx, poolDenom).Amount
		if totalShares.LT(sharesToRemove) {
			return sdkerrors.Wrapf(
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
				return err
			}
		}

		totalReserve0ToRemove = totalReserve0ToRemove.Add(outAmount0)
		totalReserve1ToRemove = totalReserve1ToRemove.Add(outAmount1)

		ctx.EventManager().EmitEvent(types.CreateWithdrawEvent(
			callerAddr,
			receiverAddr,
			pairID.Token0,
			pairID.Token1,
			tickIndex,
			fee,
			outAmount0,
			outAmount1,
			sharesToRemove,
		))
	}

	if totalReserve0ToRemove.IsPositive() {
		coin0 := sdk.NewCoin(pairID.Token0, totalReserve0ToRemove)

		err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			receiverAddr,
			sdk.Coins{coin0},
		)
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
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) MultiHopSwapCore(
	goCtx context.Context,
	amountIn sdk.Int,
	routes []*types.MultiHopRoute,
	exitLimitPrice math_utils.PrecDec,
	pickBestRoute bool,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
) (coinOut sdk.Coin, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var routeErrors []error
	initialInCoin := sdk.NewCoin(routes[0].Hops[0], amountIn)
	stepCache := make(map[multihopCacheKey]StepResult)
	var bestRoute struct {
		write   func()
		coinOut sdk.Coin
		route   []string
	}
	bestRoute.coinOut = sdk.Coin{Amount: sdk.ZeroInt()}

	for _, route := range routes {
		routeCoinOut, writeRoute, err := k.RunMultihopRoute(
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
		}
		if !pickBestRoute {
			break
		}
	}

	if len(routeErrors) == len(routes) {
		// All routes have failed

		allErr := utils.JoinErrors(types.ErrAllMultiHopRoutesFailed, routeErrors...)

		return sdk.Coin{}, allErr
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

	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		receiverAddr,
		sdk.Coins{bestRoute.coinOut},
	)
	if err != nil {
		return sdk.Coin{}, err
	}
	ctx.EventManager().EmitEvent(types.CreateMultihopSwapEvent(
		callerAddr,
		receiverAddr,
		initialInCoin.Denom,
		bestRoute.coinOut.Denom,
		initialInCoin.Amount,
		bestRoute.coinOut.Amount,
		bestRoute.route,
	))

	return bestRoute.coinOut, nil
}

// Handles MsgPlaceLimitOrder, initializing (tick, pair) data structures if needed, calculating and
// storing information for a new limit order at a specific tick.
func (k Keeper) PlaceLimitOrderCore(
	goCtx context.Context,
	tokenIn string,
	tokenOut string,
	amountIn sdk.Int,
	tickIndexInToOut int64,
	orderType types.LimitOrderType,
	goodTil *time.Time,
	maxAmountOut *sdk.Int,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
) (trancheKey string, totalInCoin sdk.Coin, swapInCoin sdk.Coin, swapOutCoin sdk.Coin, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var pairID *types.PairID
	pairID, err = types.NewPairIDFromUnsorted(tokenIn, tokenOut)
	if err != nil {
		return
	}

	amountLeft, totalIn := amountIn, sdk.ZeroInt()

	// For everything except just-in-time (JIT) orders try to execute as a swap first
	if !orderType.IsJIT() {
		// This is ok because tokenOut is provided to the constructor of PairID above
		takerTradePairID := pairID.MustTradePairIDFromMaker(tokenOut)
		var limitPrice math_utils.PrecDec
		limitPrice, err = types.CalcPrice(tickIndexInToOut)
		err = err
		if err != nil {
			return
		}

		var orderFilled bool
		swapInCoin, swapOutCoin, orderFilled, err = k.SwapWithCache(
			ctx,
			takerTradePairID,
			amountIn,
			maxAmountOut,
			&limitPrice,
		)
		if err != nil {
			return
		}

		if orderType.IsFoK() && !orderFilled {
			err = types.ErrFoKLimitOrderNotFilled
			return
		}

		totalIn = swapInCoin.Amount
		amountLeft = amountLeft.Sub(swapInCoin.Amount)

		if swapOutCoin.IsPositive() {
			err = k.bankKeeper.SendCoinsFromModuleToAccount(
				ctx,
				types.ModuleName,
				receiverAddr,
				sdk.Coins{swapOutCoin},
			)
			if err != nil {
				return
			}
		}
	}

	// This is ok because pairID was constructed from tokenIn above.
	makerTradePairID := pairID.MustTradePairIDFromMaker(tokenIn)
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
		return
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

	sharesIssued := sdk.ZeroInt()
	// FOR GTC, JIT & GoodTil try to place a maker limitOrder with remaining Amount
	if amountLeft.IsPositive() &&
		(orderType.IsGTC() || orderType.IsJIT() || orderType.IsGoodTil()) {
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

	if totalIn.IsPositive() {
		totalInCoin = sdk.NewCoin(tokenIn, totalIn)

		err = k.bankKeeper.SendCoinsFromAccountToModule(
			ctx,
			callerAddr,
			types.ModuleName,
			sdk.Coins{totalInCoin},
		)
		if err != nil {
			return
		}
	}

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

	return
}

// Handles MsgCancelLimitOrder, removing a specified number of shares from a limit order
// and returning the respective amount in terms of the reserve to the user.
func (k Keeper) CancelLimitOrderCore(
	goCtx context.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	trancheUser, found := k.GetLimitOrderTrancheUser(ctx, callerAddr.String(), trancheKey)
	if !found {
		return types.ErrActiveLimitOrderNotFound
	}

	tradePairID, tickIndex := trancheUser.TradePairID, trancheUser.TickIndexTakerToMaker
	tranche := k.GetLimitOrderTranche(
		ctx,
		&types.LimitOrderTrancheKey{
			TradePairID:           tradePairID,
			TickIndexTakerToMaker: tickIndex,
			TrancheKey:            trancheKey,
		},
	)
	if tranche == nil {
		return types.ErrActiveLimitOrderNotFound
	}

	amountToCancel := tranche.RemoveTokenIn(trancheUser)
	trancheUser.SharesCancelled = trancheUser.SharesCancelled.Add(amountToCancel)

	if amountToCancel.IsPositive() {
		coinOut := sdk.NewCoin(tradePairID.MakerDenom, amountToCancel)

		err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			callerAddr,
			sdk.Coins{coinOut},
		)
		if err != nil {
			return err
		}

		k.SaveTrancheUser(ctx, trancheUser)
		k.SaveTranche(ctx, tranche)

		if trancheUser.OrderType.HasExpiration() {
			k.RemoveLimitOrderExpiration(ctx, *tranche.ExpirationTime, tranche.Key.KeyMarshal())
		}
	} else {
		return sdkerrors.Wrapf(types.ErrCancelEmptyLimitOrder, "%s", tranche.Key.TrancheKey)
	}

	pairID := tradePairID.MustPairID()
	ctx.EventManager().EmitEvent(types.CancelLimitOrderEvent(
		callerAddr,
		pairID.Token0,
		pairID.Token1,
		tradePairID.MakerDenom,
		tradePairID.TakerDenom,
		amountToCancel,
		trancheKey,
	))

	return nil
}

// Handles MsgWithdrawFilledLimitOrder, calculates and sends filled liqudity from module to user
// for a limit order based on amount wished to receive.
func (k Keeper) WithdrawFilledLimitOrderCore(
	goCtx context.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	trancheUser, found := k.GetLimitOrderTrancheUser(
		ctx,
		callerAddr.String(),
		trancheKey,
	)
	if !found {
		return sdkerrors.Wrapf(types.ErrValidLimitOrderTrancheNotFound, "%s", trancheKey)
	}

	tradePairID, tickIndex := trancheUser.TradePairID, trancheUser.TickIndexTakerToMaker
	pairID := tradePairID.MustPairID()

	tranche, wasFilled, found := k.FindLimitOrderTranche(
		ctx,
		&types.LimitOrderTrancheKey{
			TradePairID:           tradePairID,
			TickIndexTakerToMaker: tickIndex,
			TrancheKey:            trancheKey,
		},
	)

	amountOutTokenOut := math_utils.ZeroPrecDec()
	remainingTokenIn := sdk.ZeroInt()
	// It's possible that a TrancheUser exists but tranche does not if LO was filled entirely through a swap
	if found {
		var amountOutTokenIn sdk.Int
		amountOutTokenIn, amountOutTokenOut = tranche.Withdraw(trancheUser)

		if wasFilled {
			// This is only relevant for inactive JIT and GoodTil limit orders
			remainingTokenIn = tranche.RemoveTokenIn(trancheUser)
			k.SaveInactiveTranche(ctx, tranche)
		} else {
			k.SetLimitOrderTranche(ctx, tranche)
		}

		trancheUser.SharesWithdrawn = trancheUser.SharesWithdrawn.Add(amountOutTokenIn)
	}

	k.SaveTrancheUser(ctx, trancheUser)

	if amountOutTokenOut.IsPositive() || remainingTokenIn.IsPositive() {
		coinTakerDenomOut := sdk.NewCoin(tradePairID.TakerDenom, amountOutTokenOut.TruncateInt())
		coinMakerDenomRefund := sdk.NewCoin(tradePairID.MakerDenom, remainingTokenIn)
		coins := sdk.NewCoins(coinTakerDenomOut, coinMakerDenomRefund)
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, callerAddr, coins); err != nil {
			return err
		}
	} else {
		return types.ErrWithdrawEmptyLimitOrder
	}

	ctx.EventManager().EmitEvent(types.WithdrawFilledLimitOrderEvent(
		callerAddr,
		pairID.Token0,
		pairID.Token1,
		tradePairID.MakerDenom,
		tradePairID.TakerDenom,
		amountOutTokenOut.TruncateInt(),
		trancheKey,
	))

	return nil
}
