package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/utils/math"
	"github.com/neutron-org/neutron/x/dex/types"
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
		copy := *maxAmountMakerDenom
		remainingMakerDenom = &copy
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
		if liq.Price().MulInt(remainingTakerDenom).LT(math_utils.OnePrecDec()) {
			orderFilled = true
			break
		}

		if useMaxOut {
			copy := remainingMakerDenom.Sub(outAmount)
			remainingMakerDenom = &copy

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
