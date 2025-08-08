package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	"github.com/neutron-org/neutron/v7/x/dex/types"
)

func (k Keeper) Swap(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
	maxAmountTakerDenom math_utils.PrecDec,
	maxAmountMakerDenom *math.Int,
	limitPrice *math_utils.PrecDec,
) (totalTakerCoin, totalMakerCoin types.PrecDecCoin, orderFilled bool, err error) {
	gasBefore := ctx.GasMeter().GasConsumed()
	useMaxOut := maxAmountMakerDenom != nil
	var remainingMakerDenom *math_utils.PrecDec
	if useMaxOut {
		temp := math_utils.NewPrecDecFromInt(*maxAmountMakerDenom)
		remainingMakerDenom = &temp
	}

	remainingTakerDenom := maxAmountTakerDenom
	totalMakerDenom := math_utils.ZeroPrecDec()
	orderFilled = false

	// verify that amount left is not zero and that there are additional valid ticks to check
	liqIter := k.NewLiquidityIterator(ctx, tradePairID)
	defer liqIter.Close()
	for {
		liq := liqIter.Next()
		if liq == nil {
			break
		}

		// break as soon as we iterated past limitPrice
		if limitPrice != nil && liq.Price().GT(*limitPrice) {
			break
		}

		inAmount, outAmount := liq.Swap(remainingTakerDenom, remainingMakerDenom)

		swapMetadata := types.SwapMetadata{
			AmountIn:  inAmount,
			AmountOut: outAmount,
			TokenIn:   tradePairID.TakerDenom,
		}
		k.SaveLiquidity(ctx, liq, swapMetadata)

		remainingTakerDenom = remainingTakerDenom.Sub(inAmount)
		totalMakerDenom = totalMakerDenom.Add(outAmount)

		// break if remainingTakerDenom will yield less than 1 tokenOut at current price
		// this avoids unnecessary iteration for marginal return.
		// this also catches the normal exit case where remainingTakerDenom == 0

		if remainingTakerDenom.Quo(liq.Price()).LT(math_utils.OnePrecDec()) {
			orderFilled = true
			break
		}

		if useMaxOut {
			temp := remainingMakerDenom.Sub(outAmount)
			remainingMakerDenom = &temp

			// if maxAmountOut has been used up then exit
			if !remainingMakerDenom.IsPositive() {
				orderFilled = true
				break
			}
		}
	}
	totalTakerDenom := maxAmountTakerDenom.Sub(remainingTakerDenom)

	gasAfter := ctx.GasMeter().GasConsumed()
	ctx.EventManager().EmitEvents(types.GetEventsGasConsumed(gasBefore, gasAfter))

	return types.NewPrecDecCoin(
			tradePairID.TakerDenom,
			totalTakerDenom,
		), types.NewPrecDecCoin(
			tradePairID.MakerDenom,
			totalMakerDenom,
		), orderFilled, nil
}

func (k Keeper) SwapWithCache(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
	maxAmountIn math_utils.PrecDec,
	maxAmountOut *math.Int,
	limitPrice *math_utils.PrecDec,
) (totalIn, totalOut types.PrecDecCoin, orderFilled bool, err error) {
	cacheCtx, writeCache := ctx.CacheContext()
	totalIn, totalOut, orderFilled, err = k.Swap(
		cacheCtx,
		tradePairID,
		maxAmountIn,
		maxAmountOut,
		limitPrice,
	)

	writeCache()

	return totalIn, totalOut, orderFilled, err
}

func (k Keeper) SaveLiquidity(sdkCtx sdk.Context, liquidityI types.Liquidity, swapMetadata ...types.SwapMetadata) {
	switch liquidity := liquidityI.(type) {
	case *types.LimitOrderTranche:
		// If there is still makerReserves we will save the tranche as active, if not, we will move it to inactive
		k.UpdateTranche(sdkCtx, liquidity, swapMetadata...)
	case *types.PoolLiquidity:
		// Save updated to both sides of the pool. If one of the sides is empty it will be deleted
		k.UpdatePool(sdkCtx, liquidity.Pool, swapMetadata...)
	default:
		panic("Invalid liquidity type")
	}
}

// Wrapper for taker LimitOrders
// Ensures Fok behavior is correct and that the output >= limit price output
func (k Keeper) TakerLimitOrderSwap(
	ctx sdk.Context,
	tradePairID types.TradePairID,
	amountIn math_utils.PrecDec,
	maxAmountOut *math.Int,
	limitPrice math_utils.PrecDec,
	minAvgSellPrice math_utils.PrecDec,
	orderType types.LimitOrderType,
) (totalInCoin, totalOutCoin types.PrecDecCoin, err error) {
	totalInCoin, totalOutCoin, orderFilled, err := k.SwapWithCache(
		ctx,
		&tradePairID,
		amountIn,
		maxAmountOut,
		&limitPrice,
	)
	if err != nil {
		return types.PrecDecCoin{}, types.PrecDecCoin{}, err
	}

	if orderType.IsFoK() && !orderFilled {
		return types.PrecDecCoin{}, types.PrecDecCoin{}, types.ErrFoKLimitOrderNotFilled
	}

	if totalInCoin.Amount.IsZero() {
		return types.PrecDecCoin{}, types.PrecDecCoin{}, types.ErrNoLiquidity
	}

	truePrice := totalOutCoin.Amount.Quo(totalInCoin.Amount)

	if truePrice.LT(minAvgSellPrice) {
		return types.PrecDecCoin{}, types.PrecDecCoin{}, types.ErrLimitPriceNotSatisfied
	}

	if totalOutCoin.Amount.IsZero() {
		return types.PrecDecCoin{}, types.PrecDecCoin{}, types.ErrTradeTooSmall
	}

	return totalInCoin, totalOutCoin, nil
}

// Wrapper for maker LimitOrders
// Ensures the swap portion + maker portion of the limit order will have an output >= the limit price output
func (k Keeper) MakerLimitOrderSwap(
	ctx sdk.Context,
	tradePairID types.TradePairID,
	amountIn math_utils.PrecDec,
	limitPrice math_utils.PrecDec,
	minAvgSellPrice math_utils.PrecDec,
) (totalInCoin, totalOutCoin types.PrecDecCoin, filled bool, err error) {
	totalInCoin, totalOutCoin, filled, err = k.SwapWithCache(
		ctx,
		&tradePairID,
		amountIn,
		nil,
		&limitPrice,
	)
	if err != nil {
		return types.PrecDecCoin{}, types.PrecDecCoin{}, filled, err
	}

	if totalInCoin.Amount.IsPositive() {
		remainingIn := amountIn.Sub(totalInCoin.Amount)
		expectedOutMakerPortion := remainingIn.Quo(limitPrice)
		totalExpectedOut := expectedOutMakerPortion.Add(totalOutCoin.Amount)
		truePrice := totalExpectedOut.Quo(amountIn)

		if truePrice.LT(minAvgSellPrice) {
			return types.PrecDecCoin{}, types.PrecDecCoin{}, false, types.ErrLimitPriceNotSatisfied
		}
	}

	return totalInCoin, totalOutCoin, filled, nil
}
