package keeper

import (
	"context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v8/utils/math"
	"github.com/neutron-org/neutron/v8/x/dex/types"
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
) (trancheKey string, totalInCoin, swapInCoin, swapOutCoin types.PrecDecCoin, err error) {
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
		err = k.FractionalBanker.SendFractionalCoinsFromDexToAccount(
			ctx,
			receiverAddr,
			[]types.PrecDecCoin{swapOutCoin},
		)
		if err != nil {
			return trancheKey, totalInCoin, swapInCoin, swapOutCoin, err
		}
	}

	if totalIn.IsPositive() {
		totalInCoin = types.NewPrecDecCoin(tokenIn, totalIn)

		err = k.FractionalBanker.SendFractionalCoinsFromAccountToDex(
			ctx,
			callerAddr,
			[]types.PrecDecCoin{totalInCoin},
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
	totalIn math_utils.PrecDec,
	swapInCoin, swapOutCoin types.PrecDecCoin,
	sharesIssued math.Int,
	minAvgSellPrice math_utils.PrecDec,
	err error,
) {
	amountLeft := math_utils.NewPrecDecFromInt(amountIn)

	limitBuyPrice, err := types.CalcPrice(tickIndexInToOut)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), math_utils.ZeroPrecDec(), err
	}

	// Use limitPrice for minAvgSellPrice if it has not been specified
	minAvgSellPrice = math_utils.OnePrecDec().Quo(limitBuyPrice)

	if minAvgSellPriceP != nil {
		minAvgSellPrice = *minAvgSellPriceP
	}

	var orderFilled bool
	if orderType.IsTakerOnly() {
		swapInCoin, swapOutCoin, err = k.TakerLimitOrderSwap(ctx, *takerTradePairID, amountLeft, maxAmountOut, limitBuyPrice, minAvgSellPrice, orderType)
	} else {
		swapInCoin, swapOutCoin, orderFilled, err = k.MakerLimitOrderSwap(ctx, *takerTradePairID, amountLeft, limitBuyPrice, minAvgSellPrice)
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
	trancheUser, err := k.GetOrInitLimitOrderTrancheUser(
		ctx,
		makerTradePairID,
		tickIndexTakerToMaker,
		trancheKey,
		orderType,
		receiverAddr.String(),
	)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), minAvgSellPrice, err
	}

	// FOR GTC, JIT & GoodTil try to place a maker limitOrder with remaining Amount
	if amountLeft.TruncateInt().IsPositive() && !orderFilled &&
		(orderType.IsGTC() || orderType.IsJIT() || orderType.IsGoodTil()) {

		amountToPlace := amountLeft.TruncateInt()
		placeTranche.PlaceMakerLimitOrder(amountToPlace)
		trancheUser.SharesOwned = trancheUser.SharesOwned.Add(amountToPlace)

		if orderType.HasExpiration() {
			goodTilRecord := NewLimitOrderExpiration(placeTranche)
			k.SetLimitOrderExpiration(ctx, goodTilRecord)
			ctx.GasMeter().ConsumeGas(types.ExpiringLimitOrderGas, "Expiring LimitOrder Fee")
		}

		// This update will ALWAYS save the tranche as active.
		// But we use the general updateTranche function so the correct events are emitted
		k.UpdateTranche(ctx, placeTranche)

		totalIn = totalIn.Add(math_utils.NewPrecDecFromInt(amountToPlace))
		sharesIssued = amountToPlace
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
