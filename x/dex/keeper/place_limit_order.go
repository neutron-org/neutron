package keeper

import (
	"context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// PlaceLimitOrderCore handles the logic for MsgPlaceLimitOrder including bank operations and event emissions.
func (k Keeper) PlaceLimitOrderCore(
	goCtx context.Context,
	tokenIn string,
	tokenOut string,
	amountIn math.Int,
	tickIndexInToOut int64,
	orderType types.LimitOrderType,
	goodTil *time.Time,
	maxAmountOut *math.Int,
	minAvgSellPriceP *math_utils.PrecDec,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
) (trancheKey string, totalInCoin, swapInCoin, swapOutCoin sdk.Coin, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	takerTradePairID, err := types.NewTradePairID(tokenIn, tokenOut)
	if err != nil {
		return trancheKey, totalInCoin, swapInCoin, swapOutCoin, err
	}
	trancheKey, totalIn, swapInCoin, swapOutCoin, sharesIssued, minAvgSellPrice, err := k.ExecutePlaceLimitOrder(
		ctx,
		takerTradePairID,
		amountIn,
		tickIndexInToOut,
		orderType,
		goodTil,
		maxAmountOut,
		minAvgSellPriceP,
		receiverAddr,
	)
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

	// This will never panic because we've already successfully constructed a TradePairID above
	pairID := takerTradePairID.MustPairID()
	ctx.EventManager().EmitEvent(types.CreatePlaceLimitOrderEvent(
		callerAddr,
		receiverAddr,
		pairID.Token0,
		pairID.Token1,
		tokenIn,
		tokenOut,
		totalIn,
		amountIn,
		tickIndexInToOut,
		orderType.String(),
		maxAmountOut,
		minAvgSellPrice,
		sharesIssued,
		trancheKey,
		swapInCoin.Amount,
		swapOutCoin.Amount,
	))

	return trancheKey, totalInCoin, swapInCoin, swapOutCoin, nil
}

// ExecutePlaceLimitOrder handles the core logic for PlaceLimitOrder -- performing taker a swap
// and (when applicable) adding a maker limit order to the orderbook.
// IT DOES NOT PERFORM ANY BANKING OPERATIONS
func (k Keeper) ExecutePlaceLimitOrder(
	ctx sdk.Context,
	takerTradePairID *types.TradePairID,
	amountIn math.Int,
	tickIndexInToOut int64,
	orderType types.LimitOrderType,
	goodTil *time.Time,
	maxAmountOut *math.Int,
	minAvgSellPriceP *math_utils.PrecDec,
	receiverAddr sdk.AccAddress,
) (
	trancheKey string,
	totalIn math.Int,
	swapInCoin, swapOutCoin sdk.Coin,
	sharesIssued math.Int,
	minAvgSellPrice math_utils.PrecDec,
	err error,
) {
	amountLeft := amountIn

	limitBuyPrice, err := types.CalcPrice(tickIndexInToOut)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), math_utils.ZeroPrecDec(), err
	}

	// Use limitPrice for minAvgSellPrice if it has not been specified
	minAvgSellPrice = math_utils.OnePrecDec().Quo(limitBuyPrice)

	if minAvgSellPriceP != nil {
		minAvgSellPrice = *minAvgSellPriceP
	}

	// Ensure that after rounding user will get at least 1 token out.
	err = types.ValidateFairOutput(amountIn, limitBuyPrice)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), minAvgSellPrice, err
	}

	var orderFilled bool
	if orderType.IsTakerOnly() {
		swapInCoin, swapOutCoin, err = k.TakerLimitOrderSwap(ctx, *takerTradePairID, amountIn, maxAmountOut, limitBuyPrice, minAvgSellPrice, orderType)
	} else {
		swapInCoin, swapOutCoin, orderFilled, err = k.MakerLimitOrderSwap(ctx, *takerTradePairID, amountIn, limitBuyPrice, minAvgSellPrice)
	}
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), minAvgSellPrice, err
	}

	totalIn = swapInCoin.Amount
	amountLeft = amountLeft.Sub(swapInCoin.Amount)

	makerTradePairID := takerTradePairID.Reversed()
	tickIndexTakerToMaker := tickIndexInToOut * -1
	var placeTranche *types.LimitOrderTranche
	placeTranche, err = k.GetOrInitPlaceTranche(
		ctx,
		makerTradePairID,
		tickIndexTakerToMaker,
		goodTil,
		orderType,
	)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), minAvgSellPrice, err
	}

	trancheKey = placeTranche.Key.TrancheKey
	trancheUser := k.GetOrInitLimitOrderTrancheUser(
		ctx,
		makerTradePairID,
		tickIndexTakerToMaker,
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
		err = types.ValidateFairOutput(amountLeft, limitBuyPrice)
		if err != nil {
			return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), minAvgSellPrice, err
		}
		placeTranche.PlaceMakerLimitOrder(amountLeft)
		trancheUser.SharesOwned = trancheUser.SharesOwned.Add(amountLeft)

		if orderType.HasExpiration() {
			goodTilRecord := NewLimitOrderExpiration(placeTranche)
			k.SetLimitOrderExpiration(ctx, goodTilRecord)
			ctx.GasMeter().ConsumeGas(types.ExpiringLimitOrderGas, "Expiring LimitOrder Fee")
		}

		// This update will ALWAYS save the tranche as active.
		// But we use the general updateTranche function so the correct events are emitted
		k.UpdateTranche(ctx, placeTranche)

		totalIn = totalIn.Add(amountLeft)
		sharesIssued = amountLeft
	}

	// This update will ALWAYS save the trancheUser as active.
	// But we use the general updateTranche function so the correct events are emitted
	k.UpdateTrancheUser(ctx, trancheUser)

	if orderType.IsJIT() {
		err = k.AssertCanPlaceJIT(ctx)
		if err != nil {
			return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), minAvgSellPrice, err
		}
		k.IncrementJITsInBlock(ctx)
	}

	return trancheKey, totalIn, swapInCoin, swapOutCoin, sharesIssued, minAvgSellPrice, nil
}
