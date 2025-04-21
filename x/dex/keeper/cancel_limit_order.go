package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// CancelLimitOrderCore handles the logic for MsgCancelLimitOrder including bank operations and event emissions.
func (k Keeper) CancelLimitOrderCore(
	goCtx context.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) (makerCoinOut, takerCoinOut sdk.Coin, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	makerCoinOut, takerCoinOut, err = k.ExecuteCancelLimitOrder(ctx, trancheKey, callerAddr)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	coinsOut := sdk.NewCoins(makerCoinOut, takerCoinOut)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		callerAddr,
		coinsOut,
	)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	makerDenom := makerCoinOut.Denom
	takerDenom := takerCoinOut.Denom
	// This will never panic since PairID has already been successfully constructed during tranche creation
	pairID := types.MustNewPairID(makerDenom, takerDenom)
	ctx.EventManager().EmitEvent(types.CancelLimitOrderEvent(
		callerAddr,
		pairID.Token0,
		pairID.Token1,
		makerDenom,
		takerDenom,
		takerCoinOut.Amount,
		makerCoinOut.Amount,
		trancheKey,
	))

	return makerCoinOut, takerCoinOut, nil
}

// ExecuteCancelLimitOrder handles the core logic for CancelLimitOrder -- removing remaining TokenIn from the
// LimitOrderTranche and returning it to the user, updating the number of canceled shares on the LimitOrderTrancheUser.
// IT DOES NOT PERFORM ANY BANKING OPERATIONS
func (k Keeper) ExecuteCancelLimitOrder(
	ctx sdk.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) (makerCoinOut, takerCoinOut sdk.Coin, err error) {
	trancheUser, found := k.GetLimitOrderTrancheUser(ctx, callerAddr.String(), trancheKey)
	if !found {
		return sdk.Coin{}, sdk.Coin{}, sdkerrors.Wrapf(types.ErrValidLimitOrderTrancheNotFound, "%s", trancheKey)
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
	if !found {
		return sdk.Coin{}, sdk.Coin{}, sdkerrors.Wrapf(types.ErrValidLimitOrderTrancheNotFound, "%s", trancheKey)
	}

	makerAmountToReturn := tranche.RemoveTokenIn(trancheUser)
	_, takerAmountOut := tranche.Withdraw(trancheUser)

	// Remove the canceled shares from the maker side of the limitOrder
	tranche.TotalMakerDenom = tranche.TotalMakerDenom.Sub(trancheUser.SharesOwned)

	// Calculate total number of shares removed previously withdrawn by the user (denominated in takerDenom)
	sharesWithdrawnTakerDenom := math_utils.NewPrecDecFromInt(trancheUser.SharesWithdrawn).
		Quo(tranche.PriceTakerToMaker).
		TruncateInt()

	// Calculate the total amount removed including prior withdrawals (denominated in takerDenom)
	totalAmountOutTakerDenom := sharesWithdrawnTakerDenom.Add(takerAmountOut)

	// Decrease the tranche TotalTakerDenom by the amount being removed
	tranche.TotalTakerDenom = tranche.TotalTakerDenom.Sub(totalAmountOutTakerDenom)

	// Set TrancheUser to 100% shares withdrawn
	trancheUser.SharesWithdrawn = trancheUser.SharesOwned

	if !makerAmountToReturn.IsPositive() && !takerAmountOut.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, sdkerrors.Wrapf(types.ErrCancelEmptyLimitOrder, "%s", tranche.Key.TrancheKey)
	}

	// This will ALWAYS result in a deletion of the TrancheUser, but we still use UpdateTranche user so that the relevant events will be emitted
	k.UpdateTrancheUser(ctx, trancheUser)

	// If there is still liquidity from other shareholders we will either save the tranche as active/inactive or delete it entirely
	if wasFilled {
		k.UpdateInactiveTranche(ctx, tranche)
	} else {
		k.UpdateTranche(ctx, tranche)
	}

	// If the tranche still is being moved to inactive we can safely remove the LimitOrderExpiration.
	// If there is still remaining dust after the cancel we leave it in on the orderbook
	// It will still be purged according to the expiration logic
	if !tranche.HasTokenIn() && tranche.HasExpiration() {
		k.RemoveLimitOrderExpiration(ctx, *tranche.ExpirationTime, tranche.Key.KeyMarshal())
	}

	makerCoinOut = sdk.NewCoin(tradePairID.MakerDenom, makerAmountToReturn)
	takerCoinOut = sdk.NewCoin(tradePairID.TakerDenom, takerAmountOut)

	return makerCoinOut, takerCoinOut, nil
}
