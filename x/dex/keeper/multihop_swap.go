package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v2/utils/math"
	"github.com/neutron-org/neutron/v2/x/dex/types"
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
	LeftIn  sdk.Coin
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
	bCtx *types.BranchableCache,
	step MultihopStep,
	inCoin sdk.Coin,
	stepCache map[multihopCacheKey]StepResult,
) (sdk.Coin, sdk.Coin, *types.BranchableCache, error) {
	cacheKey := newCacheKey(step.tradePairID.TakerDenom, step.tradePairID.MakerDenom, inCoin.Amount)
	val, ok := stepCache[cacheKey]
	if ok {
		ctxBranchCopy := val.Ctx.Branch()
		return val.LeftIn, val.CoinOut, ctxBranchCopy, val.Err
	}

	// Due to rounding on swap it is possible to leak tokens at each hop.
	// As an intermediary fix, we credit the unswapped coins back to the user's account.
	// To solve this without sending user dust we would have to pre-calculate the route such that
	// the amount in will be used completely at each step.

	leftIn, coinOut, err := k.SwapFullAmountIn(bCtx.Ctx, step.tradePairID, inCoin.Amount)
	ctxBranch := bCtx.Branch()
	stepCache[cacheKey] = StepResult{Ctx: bCtx, CoinOut: coinOut, LeftIn: leftIn, Err: err}
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, bCtx, err
	}

	return leftIn, coinOut, ctxBranch, nil
}

func (k Keeper) RunMultihopRoute(
	ctx sdk.Context,
	route types.MultiHopRoute,
	initialInCoin sdk.Coin,
	exitLimitPrice math_utils.PrecDec,
	stepCache map[multihopCacheKey]StepResult,
) (sdk.Coins, sdk.Coin, func(), error) {
	routeData, err := k.HopsToRouteData(ctx, route.Hops)
	if err != nil {
		return sdk.Coins{}, sdk.Coin{}, nil, err
	}
	currentPrice := math_utils.OnePrecDec()

	var currentOutCoin sdk.Coin
	var leftIn sdk.Coin
	inCoin := initialInCoin
	bCacheCtx := types.NewBranchableCache(ctx)

	var dust sdk.Coins

	for _, step := range routeData {
		// If we can't hit the best possible price we can greedily abort
		priceUpperbound := currentPrice.Mul(step.RemainingBestPrice)
		if exitLimitPrice.GT(priceUpperbound) {
			return sdk.Coins{}, sdk.Coin{}, bCacheCtx.WriteCache, types.ErrExitLimitPriceHit
		}

		leftIn, currentOutCoin, bCacheCtx, err = k.MultihopStep(
			bCacheCtx,
			step,
			inCoin,
			stepCache,
		)
		inCoin = currentOutCoin
		if err != nil {
			return sdk.Coins{}, sdk.Coin{}, nil, sdkerrors.Wrapf(
				err,
				"Failed at pair: %s",
				step.tradePairID.MustPairID().CanonicalString(),
			)
		}

		// Add what hasn't been swapped to dust
		dust = dust.Add(leftIn)

		currentPrice = math_utils.NewPrecDecFromInt(currentOutCoin.Amount).
			Quo(math_utils.NewPrecDecFromInt(initialInCoin.Amount))
	}

	if exitLimitPrice.GT(currentPrice) {
		return sdk.Coins{}, sdk.Coin{}, nil, types.ErrExitLimitPriceHit
	}

	return dust, currentOutCoin, bCacheCtx.WriteCache, nil
}

// SwapFullAmountIn swaps full amount of given `amountIn` to the `tradePairID` taker denom.
// NOTE: SwapFullAmountIn does not ensure that 100% of amountIn is used. Due to rounding it is possible that
// a dust amount of AmountIn remains unswapped. It is the caller's responsibility to handle this appropriately.
// It returns remaining dust as a first argument.
func (k Keeper) SwapFullAmountIn(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
	amountIn math.Int,
) (leftIn sdk.Coin, totalOut sdk.Coin, err error) {
	swapAmountTakerDenom, swapAmountMakerDenom, orderFilled, err := k.Swap(
		ctx,
		tradePairID,
		amountIn,
		nil,
		nil,
	)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	if !orderFilled {
		return sdk.Coin{}, sdk.Coin{}, types.ErrInsufficientLiquidity
	}

	leftIn = sdk.Coin.Sub(sdk.NewCoin(swapAmountTakerDenom.Denom, amountIn), swapAmountTakerDenom)
	return leftIn, swapAmountMakerDenom, err
}
