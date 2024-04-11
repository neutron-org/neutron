package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v3/utils/math"
	"github.com/neutron-org/neutron/v3/x/dex/types"
)

type MultihopStep struct {
	RemainingBestPrice math_utils.PrecDec
	tradePairID        *types.TradePairID
}

func (k Keeper) HopsToRouteData(
	ctx sdk.Context,
	hops []string,
) ([]MultihopStep, error) {
	nPairs := len(hops) - 1
	routeArr := make([]MultihopStep, nPairs)
	priceAcc := math_utils.OnePrecDec()
	for i := range routeArr {
		index := len(routeArr) - 1 - i
		tokenIn := hops[index]
		tokenOut := hops[index+1]
		tradePairID, err := types.NewTradePairID(tokenIn, tokenOut)
		if err != nil {
			return routeArr, err
		}
		price, found := k.GetCurrPrice(ctx, tradePairID)
		if !found {
			return routeArr, types.ErrInsufficientLiquidity
		}
		priceAcc = priceAcc.Mul(price)
		routeArr[index] = MultihopStep{
			tradePairID:        tradePairID,
			RemainingBestPrice: priceAcc,
		}
	}

	return routeArr, nil
}

type StepResult struct {
	Ctx     *types.BranchableCache
	CoinOut sdk.Coin
	Err     error
}

type multihopCacheKey struct {
	TokenIn  string
	TokenOut string
	InAmount math.Int
}

func newCacheKey(tokenIn, tokenOut string, inAmount math.Int) multihopCacheKey {
	return multihopCacheKey{
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
		InAmount: inAmount,
	}
}

func (k Keeper) MultihopStep(
	bctx *types.BranchableCache,
	step MultihopStep,
	inCoin sdk.Coin,
	stepCache map[multihopCacheKey]StepResult,
) (sdk.Coin, *types.BranchableCache, error) {
	cacheKey := newCacheKey(step.tradePairID.TakerDenom, step.tradePairID.MakerDenom, inCoin.Amount)
	val, ok := stepCache[cacheKey]
	if ok {
		ctxBranchCopy := val.Ctx.Branch()
		return val.CoinOut, ctxBranchCopy, val.Err
	}

	// TODO: Due to rounding on swap it is possible to leak tokens at each hop.
	// In these cases the user will lose trace amounts of tokens from intermediary steps.
	// To fix this we would have to pre-calculate the route such that the amount
	// in will be used completely at each step.
	// As an intermediary fix, we should credit the unswapped coins back to the user's account.

	coinOut, err := k.SwapFullAmountIn(bctx.Ctx, step.tradePairID, inCoin.Amount)
	ctxBranch := bctx.Branch()
	stepCache[cacheKey] = StepResult{Ctx: bctx, CoinOut: coinOut, Err: err}
	if err != nil {
		return sdk.Coin{}, bctx, err
	}

	return coinOut, ctxBranch, nil
}

func (k Keeper) RunMultihopRoute(
	ctx sdk.Context,
	route types.MultiHopRoute,
	initialInCoin sdk.Coin,
	exitLimitPrice math_utils.PrecDec,
	stepCache map[multihopCacheKey]StepResult,
) (sdk.Coin, func(), error) {
	routeData, err := k.HopsToRouteData(ctx, route.Hops)
	if err != nil {
		return sdk.Coin{}, nil, err
	}
	currentPrice := math_utils.OnePrecDec()

	var currentOutCoin sdk.Coin
	inCoin := initialInCoin
	bCacheCtx := types.NewBranchableCache(ctx)

	for _, step := range routeData {
		// If we can't hit the best possible price we can greedily abort
		priceUpperbound := currentPrice.Mul(step.RemainingBestPrice)
		if exitLimitPrice.GT(priceUpperbound) {
			return sdk.Coin{}, bCacheCtx.WriteCache, types.ErrExitLimitPriceHit
		}

		currentOutCoin, bCacheCtx, err = k.MultihopStep(
			bCacheCtx,
			step,
			inCoin,
			stepCache,
		)
		inCoin = currentOutCoin
		if err != nil {
			return sdk.Coin{}, nil, sdkerrors.Wrapf(
				err,
				"Failed at pair: %s",
				step.tradePairID.MustPairID().CanonicalString(),
			)
		}

		currentPrice = math_utils.NewPrecDecFromInt(currentOutCoin.Amount).
			Quo(math_utils.NewPrecDecFromInt(initialInCoin.Amount))
	}

	if exitLimitPrice.GT(currentPrice) {
		return sdk.Coin{}, nil, types.ErrExitLimitPriceHit
	}

	return currentOutCoin, bCacheCtx.WriteCache, nil
}

// NOTE: SwapFullAmountIn does not ensure that 100% of amountIn is used. Due to rounding it is possible that
// a dust amount of AmountIn remains unswapped. It is the caller's responsibility to handle this appropriately.
func (k Keeper) SwapFullAmountIn(ctx sdk.Context,
	tradePairID *types.TradePairID,
	amountIn math.Int,
) (totalOut sdk.Coin, err error) {
	_, swapAmountMakerDenom, orderFilled, err := k.Swap(
		ctx,
		tradePairID,
		amountIn,
		nil,
		nil,
	)
	if err != nil {
		return sdk.Coin{}, err
	}
	if !orderFilled {
		return sdk.Coin{}, types.ErrInsufficientLiquidity
	}

	return swapAmountMakerDenom, err
}
