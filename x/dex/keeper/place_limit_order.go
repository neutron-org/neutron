package keeper

import (
	"context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/types"
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
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
) (trancheKey string, totalInCoin, swapInCoin, swapOutCoin sdk.Coin, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	takerTradePairID, err := types.NewTradePairID(tokenIn, tokenOut)
	if err != nil {
		return trancheKey, totalInCoin, swapInCoin, swapOutCoin, err
	}
	trancheKey, totalIn, swapInCoin, swapOutCoin, sharesIssued, err := k.ExecutePlaceLimitOrder(
		ctx,
		takerTradePairID,
		amountIn,
		tickIndexInToOut,
		orderType,
		goodTil,
		maxAmountOut,
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
	types.EmitEventWithTimestamp(
		ctx,
		types.CreatePlaceLimitOrderEvent(
			callerAddr,
			receiverAddr,
			pairID.Token0,
			pairID.Token1,
			tokenIn,
			tokenOut,
			totalIn,
			tickIndexInToOut,
			orderType.String(),
			sharesIssued,
			trancheKey,
			swapInCoin.Amount,
			swapOutCoin.Amount,
		),
	)

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
	receiverAddr sdk.AccAddress,
) (trancheKey string, totalIn math.Int, swapInCoin, swapOutCoin sdk.Coin, sharesIssued math.Int, err error) {
	amountLeft := amountIn

	var limitPrice math_utils.PrecDec
	limitPrice, err = types.CalcPrice(tickIndexInToOut)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
	}

	// Ensure that after rounding user will get at least 1 token out.
	err = types.ValidateFairOutput(amountIn, limitPrice)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
	}

	var orderFilled bool
	if orderType.IsTakerOnly() {
		swapInCoin, swapOutCoin, err = k.TakerLimitOrderSwap(ctx, *takerTradePairID, amountIn, maxAmountOut, limitPrice, orderType)
	} else {
		swapInCoin, swapOutCoin, orderFilled, err = k.MakerLimitOrderSwap(ctx, *takerTradePairID, amountIn, limitPrice)
	}
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
	}

	totalIn = swapInCoin.Amount
	amountLeft = amountLeft.Sub(swapInCoin.Amount)

	makerTradePairID := takerTradePairID.Reversed()
	makerTickIndexTakerToMaker := tickIndexInToOut * -1
	var placeTranche *types.LimitOrderTranche
	placeTranche, err = k.GetOrInitPlaceTranche(
		ctx,
		makerTradePairID,
		makerTickIndexTakerToMaker,
		goodTil,
		orderType,
	)
	if err != nil {
		return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
	}

	trancheKey = placeTranche.Key.TrancheKey
	trancheUser := k.GetOrInitLimitOrderTrancheUser(
		ctx,
		makerTradePairID,
		makerTickIndexTakerToMaker,
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
		err = types.ValidateFairOutput(amountLeft, limitPrice)
		if err != nil {
			return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
		}
		placeTranche.PlaceMakerLimitOrder(amountLeft)
		trancheUser.SharesOwned = trancheUser.SharesOwned.Add(amountLeft)

		if orderType.HasExpiration() {
			goodTilRecord := NewLimitOrderExpiration(placeTranche)
			k.SetLimitOrderExpiration(ctx, goodTilRecord)
			ctx.GasMeter().ConsumeGas(types.ExpiringLimitOrderGas, "Expiring LimitOrder Fee")
		}

		k.SaveTranche(ctx, placeTranche)

		totalIn = totalIn.Add(amountLeft)
		sharesIssued = amountLeft
	}

	k.SaveTrancheUser(ctx, trancheUser)

	if orderType.IsJIT() {
		err = k.AssertCanPlaceJIT(ctx)
		if err != nil {
			return trancheKey, totalIn, swapInCoin, swapOutCoin, math.ZeroInt(), err
		}
		k.IncrementJITsInBlock(ctx)
	}

	return trancheKey, totalIn, swapInCoin, swapOutCoin, sharesIssued, nil
}
