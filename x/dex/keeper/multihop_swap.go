package keeper

import (
	"context"
	"errors"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

type MultihopStep struct {
	RemainingBestPrice math_utils.PrecDec
	tradePairID        *types.TradePairID
}

type MultiHopRouteOutput struct {
	write   func()
	coinOut sdk.Coin
	route   []string
	dust    sdk.Coins
}

// MultiHopSwapCore handles logic for MsgMultihopSwap including bank operations and event emissions.
func (k Keeper) MultiHopSwapCore(
	goCtx context.Context,
	amountIn math.Int,
	routes []*types.MultiHopRoute,
	exitLimitPrice math_utils.PrecDec,
	pickBestRoute bool,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
) (coinOut sdk.Coin, route []string, dust sdk.Coins, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	bestRoute, initialInCoin, err := k.CalulateMultiHopSwap(ctx, amountIn, routes, exitLimitPrice, pickBestRoute)
	if err != nil {
		return sdk.Coin{}, []string{}, sdk.Coins{}, err
	}

	bestRoute.write()
	err = k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		callerAddr,
		types.ModuleName,
		sdk.Coins{initialInCoin},
	)
	if err != nil {
		return sdk.Coin{}, []string{}, sdk.Coins{}, err
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
		return sdk.Coin{}, []string{}, sdk.Coins{}, fmt.Errorf("failed to send out coin and dust to the receiver: %w", err)
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

	return bestRoute.coinOut, bestRoute.route, bestRoute.dust, nil
}

// CalulateMultiHopSwap handles the core logic for MultiHopSwap -- simulating swap operations across all routes (when applicable)
// and picking the best route to execute. It uses a cache and does not modify state.
func (k Keeper) CalulateMultiHopSwap(
	ctx sdk.Context,
	amountIn math.Int,
	routes []*types.MultiHopRoute,
	exitLimitPrice math_utils.PrecDec,
	pickBestRoute bool,
) (bestRoute MultiHopRouteOutput, initialInCoin sdk.Coin, err error) {
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

		return MultiHopRouteOutput{}, sdk.Coin{}, allErr
	}

	return bestRoute, initialInCoin, nil
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
			return routeArr, types.ErrLimitPriceNotSatisfied
		}
		priceAcc = priceAcc.Quo(price)
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
	Dust    sdk.Coin
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
		return val.Dust, val.CoinOut, ctxBranchCopy, val.Err
	}

	// Due to rounding on swap it is possible to leak tokens at each hop.
	// As an intermediary fix, we credit the unswapped coins back to the user's account.
	// To solve this without sending user dust we would have to pre-calculate the route such that
	// the amount in will be used completely at each step.

	dust, coinOut, err := k.SwapFullAmountIn(bCtx.Ctx, step.tradePairID, inCoin.Amount)
	ctxBranch := bCtx.Branch()
	stepCache[cacheKey] = StepResult{Ctx: bCtx, CoinOut: coinOut, Dust: dust, Err: err}
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, bCtx, err
	}

	return dust, coinOut, ctxBranch, nil
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

	var stepOutCoin sdk.Coin
	var stepDust sdk.Coin
	inCoin := initialInCoin
	bCacheCtx := types.NewBranchableCache(ctx)

	var dustAcc sdk.Coins

	for _, step := range routeData {
		// If we can't hit the best possible price we can greedily abort
		priceUpperbound := currentPrice.Mul(step.RemainingBestPrice)
		if exitLimitPrice.GT(priceUpperbound) {
			return sdk.Coins{}, sdk.Coin{}, bCacheCtx.WriteCache, types.ErrLimitPriceNotSatisfied
		}

		stepDust, stepOutCoin, bCacheCtx, err = k.MultihopStep(
			bCacheCtx,
			step,
			inCoin,
			stepCache,
		)
		inCoin = stepOutCoin
		if err != nil {
			return sdk.Coins{}, sdk.Coin{}, nil, sdkerrors.Wrapf(
				err,
				"Failed at pair: %s",
				step.tradePairID.MustPairID().CanonicalString(),
			)
		}

		// Add what hasn't been swapped to dustAcc
		dustAcc = dustAcc.Add(stepDust)

		currentPrice = math_utils.NewPrecDecFromInt(stepOutCoin.Amount).
			Quo(math_utils.NewPrecDecFromInt(initialInCoin.Amount))
	}

	if exitLimitPrice.GT(currentPrice) {
		return sdk.Coins{}, sdk.Coin{}, nil, types.ErrLimitPriceNotSatisfied
	}

	return dustAcc, stepOutCoin, bCacheCtx.WriteCache, nil
}

// SwapFullAmountIn swaps full amount of given `amountIn` to the `tradePairID` taker denom.
// NOTE: SwapFullAmountIn does not ensure that 100% of amountIn is used. Due to rounding it is possible that
// a dust amount of AmountIn remains unswapped. It is the caller's responsibility to handle this appropriately.
// It returns remaining dust as a first argument.
func (k Keeper) SwapFullAmountIn(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
	amountIn math.Int,
) (dust, totalOut sdk.Coin, err error) {
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
		return sdk.Coin{}, sdk.Coin{}, types.ErrNoLiquidity
	}

	dust = sdk.Coin.Sub(sdk.NewCoin(swapAmountTakerDenom.Denom, amountIn), swapAmountTakerDenom)
	if dust.IsNegative() {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("dust coins are negative")
	}

	return dust, swapAmountMakerDenom, err
}
