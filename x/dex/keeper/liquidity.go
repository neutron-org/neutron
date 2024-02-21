package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v2/utils/math"
	"github.com/neutron-org/neutron/v2/x/dex/types"
)

func (k Keeper) Swap(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
	maxAmountTakerDenom math.Int,
	maxAmountMakerDenom *math.Int,
	limitPrice *math_utils.PrecDec,
) (totalTakerCoin, totalMakerCoin sdk.Coin, orderFilled bool, err error) {
	params := k.GetParams(ctx)
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

		// If (due to rounding) the actual price given to the maker is demonstrably unfair
		// we do not save the results of the swap and we exit.
		// While the decrease in price quality for the maker is semi-linear with the amount
		// being swapped, it is possible that the next swap could yield a "fair" price.
		// Nonethless, once the remainingTakerDenom gets small enough to start causing unfair swaps
		// it is much simpler to just abort.
		if inAmount.IsZero() || isUnfairTruePrice(params.MaxTrueTakerSpread, inAmount, outAmount, liq) {
			// If they've already swapped just end the swap
			if remainingTakerDenom.LT(maxAmountTakerDenom) {
				break
			}
			// If they have not swapped anything return informative error
			return sdk.Coin{}, sdk.Coin{}, false, types.ErrSwapAmountTooSmall
		}
		k.SaveLiquidity(ctx, liq)

		remainingTakerDenom = remainingTakerDenom.Sub(inAmount)
		totalMakerDenom = totalMakerDenom.Add(outAmount)

		// break if remainingTakerDenom will yield less than 1 tokenOut at current price
		// this avoids unnecessary iteration since outAmount will always be 0 going forward
		// this also catches the normal exit case where remainingTakerDenom == 0

		// NOTE: In theory this check should be: price * remainingTakerDenom < 1
		// but due to rounding and inaccuracy of fixed decimal math, it is possible
		// for liq.swap to use the full the amount of taker liquidity and have a leftover
		// amount of the taker Denom > than 1 token worth of maker denom
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

	// TODO: should we check for err != nil before writeCache() call?
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

func isUnfairTruePrice(
	maxTrueTakerSpread math_utils.PrecDec,
	inAmount, outAmount math.Int,
	liq types.Liquidity,
) bool {
	bookPrice := liq.Price()
	truePrice := math_utils.NewPrecDecFromInt(outAmount).QuoInt(inAmount)
	priceDiffFromExpected := truePrice.Sub(bookPrice)
	pctDiff := priceDiffFromExpected.Quo(bookPrice)

	return pctDiff.GT(maxTrueTakerSpread)
}
