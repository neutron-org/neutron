package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (k Keeper) Swap(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
	maxAmountTakerDenom math.Int,
	maxAmountMakerDenom *math.Int,
	limitPrice *math_utils.PrecDec,
) (totalTakerCoin, totalMakerCoin sdk.Coin, orderFilled bool, err error) {
	gasBefore := ctx.GasMeter().GasConsumed()
	useMaxOut := maxAmountMakerDenom != nil
	var remainingMakerDenom *math.Int
	if useMaxOut {
		temp := *maxAmountMakerDenom
		remainingMakerDenom = &temp
	}

	remainingTakerDenom := maxAmountTakerDenom
	totalMakerDenom := math.ZeroInt()
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
		// this avoids unnecessary iteration since outAmount will always be 0 going forward
		// this also catches the normal exit case where remainingTakerDenom == 0

		// This also allows us to handle a corner case where totalTakerCoin < maxAmountAmountTakerDenom
		// and there is still valid tradeable liquidity but the order cannot be filled any further due to monotonic rounding.
		if math_utils.NewPrecDecFromInt(remainingTakerDenom).Quo(liq.Price()).LT(math_utils.OnePrecDec()) {
			orderFilled = true
			break
		}

		if useMaxOut {
			temp := remainingMakerDenom.Sub(outAmount)
			remainingMakerDenom = &temp

			// if maxAmountOut has been used up then exit
			if remainingMakerDenom.LTE(math.ZeroInt()) {
				orderFilled = true
				break
			}
		}
	}
	totalTakerDenom := maxAmountTakerDenom.Sub(remainingTakerDenom)

	gasAfter := ctx.GasMeter().GasConsumed()
	ctx.EventManager().EmitEvents(types.GetEventsGasConsumed(gasBefore, gasAfter))

	return sdk.NewCoin(
			tradePairID.TakerDenom,
			totalTakerDenom,
		), sdk.NewCoin(
			tradePairID.MakerDenom,
			totalMakerDenom,
		), orderFilled, nil
}

func (k Keeper) SwapWithCache(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
	maxAmountIn math.Int,
	maxAmountOut *math.Int,
	limitPrice *math_utils.PrecDec,
) (totalIn, totalOut sdk.Coin, orderFilled bool, err error) {
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
	amountIn math.Int,
	maxAmountOut *math.Int,
	limitPrice math_utils.PrecDec,
	minAvgSellPrice math_utils.PrecDec,
	orderType types.LimitOrderType,
) (totalInCoin, totalOutCoin sdk.Coin, err error) {
	totalInCoin, totalOutCoin, orderFilled, err := k.SwapWithCache(
		ctx,
		&tradePairID,
		amountIn,
		maxAmountOut,
		&limitPrice,
	)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	if orderType.IsFoK() && !orderFilled {
		return sdk.Coin{}, sdk.Coin{}, types.ErrFoKLimitOrderNotFilled
	}

	if totalInCoin.Amount.IsZero() {
		return sdk.Coin{}, sdk.Coin{}, types.ErrNoLiquidity
	}

	truePrice := math_utils.NewPrecDecFromInt(totalOutCoin.Amount).QuoInt(totalInCoin.Amount)

	if truePrice.LT(minAvgSellPrice) {
		return sdk.Coin{}, sdk.Coin{}, types.ErrLimitPriceNotSatisfied
	}

	return totalInCoin, totalOutCoin, nil
}

// Wrapper for maker LimitOrders
// Ensures the swap portion + maker portion of the limit order will have an output >= the limit price output
func (k Keeper) MakerLimitOrderSwap(
	ctx sdk.Context,
	tradePairID types.TradePairID,
	amountIn math.Int,
	limitPrice math_utils.PrecDec,
	minAvgSellPrice math_utils.PrecDec,
) (totalInCoin, totalOutCoin sdk.Coin, filled bool, err error) {
	totalInCoin, totalOutCoin, filled, err = k.SwapWithCache(
		ctx,
		&tradePairID,
		amountIn,
		nil,
		&limitPrice,
	)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, filled, err
	}

	if totalInCoin.Amount.IsPositive() {
		remainingIn := amountIn.Sub(totalInCoin.Amount)
		expectedOutMakerPortion := math_utils.NewPrecDecFromInt(remainingIn).Quo(limitPrice)
		totalExpectedOut := expectedOutMakerPortion.Add(math_utils.NewPrecDecFromInt(totalOutCoin.Amount))
		truePrice := totalExpectedOut.QuoInt(amountIn)

		if truePrice.LT(minAvgSellPrice) {
			return sdk.Coin{}, sdk.Coin{}, false, types.ErrLimitPriceNotSatisfied
		}
	}

	return totalInCoin, totalOutCoin, filled, nil
}
