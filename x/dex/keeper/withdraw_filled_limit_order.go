package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// WithdrawFilledLimitOrderCore handles MsgWithdrawFilledLimitOrder including bank operations and event emissions.
func (k Keeper) WithdrawFilledLimitOrderCore(
	goCtx context.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) (takerCoinOut, makerCoinOut sdk.Coin, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	takerCoinOut, makerCoinOut, err = k.ExecuteWithdrawFilledLimitOrder(ctx, trancheKey, callerAddr)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// NOTE: it is possible for coinTakerDenomOut xor coinMakerDenomOut to be zero. These are removed by the sanitize call in sdk.NewCoins
	// ExecuteWithdrawFilledLimitOrder ensures that at least one of these has am amount > 0.
	coins := sdk.NewCoins(takerCoinOut, makerCoinOut)
	ctx.EventManager().EmitEvents(types.GetEventsWithdrawnAmount(sdk.NewCoins(takerCoinOut)))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, callerAddr, coins); err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	makerDenom := makerCoinOut.Denom
	takerDenom := takerCoinOut.Denom
	// This will never panic since TradePairID has already been successfully constructed by ExecuteWithdrawFilledLimitOrder
	pairID := types.MustNewPairID(makerDenom, takerDenom)
	ctx.EventManager().EmitEvent(types.WithdrawFilledLimitOrderEvent(
		callerAddr,
		pairID.Token0,
		pairID.Token1,
		makerDenom,
		takerDenom,
		takerCoinOut.Amount,
		makerCoinOut.Amount,
		trancheKey,
	))

	return takerCoinOut, makerCoinOut, nil
}

// ExecuteWithdrawFilledLimitOrder handles the for logic for WithdrawFilledLimitOrder -- calculates and sends filled liquidity from module to user,
// returns any remaining TokenIn from inactive limit orders, and updates the LimitOrderTranche and LimitOrderTrancheUser.
// IT DOES NOT PERFORM ANY BANKING OPERATIONS
func (k Keeper) ExecuteWithdrawFilledLimitOrder(
	ctx sdk.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) (takerCoinOut, makerCoinOut sdk.Coin, err error) {
	trancheUser, found := k.GetLimitOrderTrancheUser(
		ctx,
		callerAddr.String(),
		trancheKey,
	)
	if !found {
		return makerCoinOut, takerCoinOut, sdkerrors.Wrapf(types.ErrValidLimitOrderTrancheNotFound, "%s", trancheKey)
	}

	tradePairID, tickIndex := trancheUser.TradePairId, trancheUser.TickIndexTakerToMaker

	tranche, wasFilled, found := k.FindLimitOrderTranche(
		ctx,
		&types.LimitOrderTrancheKey{
			TradePairId:           tradePairID,
			TickIndexTakerToMaker: tickIndex,
			TrancheKey:            trancheKey,
		},
	)

	amountOutTokenOut := math.ZeroInt()
	remainingTokenIn := math.ZeroInt()
	// It's possible that a TrancheUser exists but tranche does not if LO was filled entirely through a swap
	if found {
		var amountOutTokenIn math.Int
		amountOutTokenIn, amountOutTokenOut = tranche.Withdraw(trancheUser)

		if wasFilled {
			// This is only relevant for inactive JIT and GoodTil limit orders
			remainingTokenIn = tranche.RemoveTokenIn(trancheUser)
			k.UpdateInactiveTranche(ctx, tranche)

			// Since the order has already been filled we treat this as a complete withdrawal
			trancheUser.SharesWithdrawn = trancheUser.SharesOwned

		} else {
			// This was an active tranche (still has MakerReserves) and we have only removed TakerReserves; we will save it as an active tranche
			k.UpdateTranche(ctx, tranche)
			trancheUser.SharesWithdrawn = trancheUser.SharesWithdrawn.Add(amountOutTokenIn)
		}

	}
	// Save the tranche user
	k.UpdateTrancheUser(ctx, trancheUser)

	if !amountOutTokenOut.IsPositive() && !remainingTokenIn.IsPositive() {
		return takerCoinOut, makerCoinOut, types.ErrWithdrawEmptyLimitOrder
	}

	takerCoinOut = sdk.NewCoin(tradePairID.TakerDenom, amountOutTokenOut)
	makerCoinOut = sdk.NewCoin(tradePairID.MakerDenom, remainingTokenIn)

	return takerCoinOut, makerCoinOut, nil
}
