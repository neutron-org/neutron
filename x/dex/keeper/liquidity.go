package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v3/utils/math"
	"github.com/neutron-org/neutron/v3/x/dex/types"
)

func (k Keeper) Swap(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
	maxAmountTakerDenom math.Int,
	maxAmountMakerDenom *math.Int,
	limitPrice *math_utils.PrecDec,
) (totalTakerCoin, totalMakerCoin sdk.Coin, orderFilled bool, err error) {
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
		if limitPrice != nil && liq.Price().LT(*limitPrice) {
			break
		}

		inAmount, outAmount := liq.Swap(remainingTakerDenom, remainingMakerDenom)

		k.SaveLiquidity(ctx, liq)

		remainingTakerDenom = remainingTakerDenom.Sub(inAmount)
		totalMakerDenom = totalMakerDenom.Add(outAmount)

		// break if remainingTakerDenom will yield less than 1 tokenOut at current price
		// this avoids unnecessary iteration since outAmount will always be 0 going forward
		// this also catches the normal exit case where remainingTakerDenom == 0

		// NOTE: In theory this check should be: price * remainingTakerDenom < 1
		// but due to rounding and inaccuracy of fixed decimal math, it is possible
		// for liq.swap to use the full the amount of taker liquidity and have a leftover
		// amount amount of the taker Denom > than 1 token worth of maker denom
		if liq.Price().MulInt(remainingTakerDenom).LT(math_utils.NewPrecDec(2)) {
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

func (k Keeper) SaveLiquidity(sdkCtx sdk.Context, liquidityI types.Liquidity) {
	switch liquidity := liquidityI.(type) {
	case *types.LimitOrderTranche:
		k.SaveTranche(sdkCtx, liquidity)

	case *types.PoolLiquidity:
		k.SetPool(sdkCtx, liquidity.Pool)
	default:
		panic("Invalid liquidity type")
	}
}

func (k Keeper) TakerLimitOrderSwap(
	ctx sdk.Context,
	tradePairID types.TradePairID,
	amountIn math.Int,
	maxAmountOut *math.Int,
	limitPrice math_utils.PrecDec,
	orderType types.LimitOrderType) (totalInCoin, totalOutCoin sdk.Coin, err error) {

	totalInCoin, totalOutCoin, orderFilled, err := k.SwapWithCache(
		ctx,
		&tradePairID,
		amountIn,
		maxAmountOut,
		&limitPrice,
	)

	if orderType.IsFoK() && !orderFilled {
		return sdk.Coin{}, sdk.Coin{}, types.ErrFoKLimitOrderNotFilled
	}

	if totalInCoin.Amount.IsZero() {
		return sdk.Coin{}, sdk.Coin{}, types.ErrInsufficientLiquidity
	}

	truePrice := math_utils.NewPrecDecFromInt(totalOutCoin.Amount).QuoInt(totalInCoin.Amount)

	if truePrice.LT(limitPrice) {
		return sdk.Coin{}, sdk.Coin{}, types.ErrLimitPriceNotSatisfied
	}

	return totalInCoin, totalOutCoin, nil

}

func (k Keeper) MakerLimitOrderSwap(
	ctx sdk.Context,
	tradePairID types.TradePairID,
	amountIn math.Int,
	limitPrice math_utils.PrecDec) (totalInCoin, totalOutCoin sdk.Coin, err error) {

	totalInCoin, totalOutCoin, _, err = k.SwapWithCache(
		ctx,
		&tradePairID,
		amountIn,
		nil,
		&limitPrice,
	)

	if totalInCoin.Amount.IsPositive() {
		remainingIn := amountIn.Sub(totalInCoin.Amount)
		expectedOutMakerPortion := limitPrice.MulInt(remainingIn).Ceil()
		totalExpectedOut := expectedOutMakerPortion.Add(math_utils.NewPrecDecFromInt(totalOutCoin.Amount))
		truePrice := totalExpectedOut.QuoInt(amountIn)

		if truePrice.LT(limitPrice) {
			return sdk.Coin{}, sdk.Coin{}, types.ErrLimitPriceNotSatisfied
		}
	}

	return totalInCoin, totalOutCoin, nil

}
