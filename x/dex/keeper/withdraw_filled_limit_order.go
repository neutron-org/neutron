package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dex/types"
)

// WithdrawFilledLimitOrderCore handles MsgWithdrawFilledLimitOrder including bank operations and event emissions.
func (k Keeper) WithdrawFilledLimitOrderCore(
	goCtx context.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amountOutTokenOut, remainingTokenIn, tradePairID, err := k.ExecuteWithdrawFilledLimitOrder(ctx, trancheKey, callerAddr)
	if err != nil {
		return err
	}

	coinTakerDenomOut := sdk.NewCoin(tradePairID.TakerDenom, amountOutTokenOut)
	coinMakerDenomRefund := sdk.NewCoin(tradePairID.MakerDenom, remainingTokenIn)
	// NOTE: it is possible for coinTakerDenomOut xor coinMakerDenomOut to be zero. These are removed by the sanitize call in sdk.NewCoins
	// ExecuteWithdrawFilledLimitOrder ensures that at least one of these has am amount > 0.
	coins := sdk.NewCoins(coinTakerDenomOut, coinMakerDenomRefund)
	ctx.EventManager().EmitEvents(types.GetEventsWithdrawnAmount(sdk.NewCoins(coinTakerDenomOut)))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, callerAddr, coins); err != nil {
		return err
	}

	// This will never panic since TradePairID has already been successfully constructed by ExecuteWithdrawFilledLimitOrder
	pairID := tradePairID.MustPairID()
	ctx.EventManager().EmitEvent(types.WithdrawFilledLimitOrderEvent(
		callerAddr,
		pairID.Token0,
		pairID.Token1,
		tradePairID.MakerDenom,
		tradePairID.TakerDenom,
		amountOutTokenOut,
		trancheKey,
	))

	return nil
}

// ExecuteWithdrawFilledLimitOrder handles the for logic for WithdrawFilledLimitOrder -- calculates and sends filled liquidity from module to user,
// returns any remaining TokenIn from inactive limit orders, and updates the LimitOrderTranche and LimitOrderTrancheUser.
// IT DOES NOT PERFORM ANY BANKING OPERATIONS
func (k Keeper) ExecuteWithdrawFilledLimitOrder(
	ctx sdk.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) (amountOutTokenOut, remainingTokenIn math.Int, tradePairID *types.TradePairID, err error) {
	trancheUser, found := k.GetLimitOrderTrancheUser(
		ctx,
		callerAddr.String(),
		trancheKey,
	)
	if !found {
		return math.ZeroInt(), math.ZeroInt(), nil, sdkerrors.Wrapf(types.ErrValidLimitOrderTrancheNotFound, "%s", trancheKey)
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

	amountOutTokenOut = math.ZeroInt()
	remainingTokenIn = math.ZeroInt()
	// It's possible that a TrancheUser exists but tranche does not if LO was filled entirely through a swap
	if found {
		var amountOutTokenIn math.Int
		amountOutTokenIn, amountOutTokenOut = tranche.Withdraw(trancheUser)

		if wasFilled {
			// This is only relevant for inactive JIT and GoodTil limit orders
			remainingTokenIn = tranche.RemoveTokenIn(trancheUser)
			k.SaveOrRemoveInactiveTranche(ctx, tranche)

			// Since the order has already been filled we treat this as a complete withdrawal
			trancheUser.SharesWithdrawn = trancheUser.SharesOwned

		} else {
			k.SetLimitOrderTranche(ctx, tranche)
			trancheUser.SharesWithdrawn = trancheUser.SharesWithdrawn.Add(amountOutTokenIn)
		}

	}

	k.SaveOrRemoveTrancheUser(ctx, trancheUser)

	if !amountOutTokenOut.IsPositive() && !remainingTokenIn.IsPositive() {
		return math.ZeroInt(), math.ZeroInt(), tradePairID, types.ErrWithdrawEmptyLimitOrder
	}

	return amountOutTokenOut, remainingTokenIn, tradePairID, nil
}
