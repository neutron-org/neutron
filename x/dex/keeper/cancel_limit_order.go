package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/types"
)

// CancelLimitOrderCore handles the logic for MsgCancelLimitOrder including bank operations and event emissions.
func (k Keeper) CancelLimitOrderCore(
	goCtx context.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	makerCoinOut, takerCoinOut, tradePairID, err := k.ExecuteCancelLimitOrder(ctx, trancheKey, callerAddr)
	if err != nil {
		return err
	}

	coinsOut := sdk.NewCoins(makerCoinOut, takerCoinOut)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		callerAddr,
		coinsOut,
	)
	if err != nil {
		return err
	}

	// This will never panic since TradePairID has already been successfully constructed by ExecuteCancelLimitOrder
	pairID := tradePairID.MustPairID()
	ctx.EventManager().EmitEvent(types.CancelLimitOrderEvent(
		callerAddr,
		pairID.Token0,
		pairID.Token1,
		tradePairID.MakerDenom,
		tradePairID.TakerDenom,
		makerCoinOut.Amount,
		takerCoinOut.Amount,
		trancheKey,
	))

	return nil
}

// ExecuteCancelLimitOrder handles the core logic for CancelLimitOrder -- removing remaining TokenIn from the
// LimitOrderTranche and returning it to the user, updating the number of canceled shares on the LimitOrderTrancheUser.
// IT DOES NOT PERFORM ANY BANKING OPERATIONS
func (k Keeper) ExecuteCancelLimitOrder(
	ctx sdk.Context,
	trancheKey string,
	callerAddr sdk.AccAddress,
) (makerCoinOut, takerCoinOut sdk.Coin, tradePairID *types.TradePairID, error error) {
	trancheUser, found := k.GetLimitOrderTrancheUser(ctx, callerAddr.String(), trancheKey)
	if !found {
		return sdk.Coin{}, sdk.Coin{}, nil, types.ErrActiveLimitOrderNotFound
	}

	tradePairID, tickIndex := trancheUser.TradePairId, trancheUser.TickIndexTakerToMaker
	tranche := k.GetLimitOrderTranche(
		ctx,
		&types.LimitOrderTrancheKey{
			TradePairId:           tradePairID,
			TickIndexTakerToMaker: tickIndex,
			TrancheKey:            trancheKey,
		},
	)
	if tranche == nil {
		return sdk.Coin{}, sdk.Coin{}, nil, types.ErrActiveLimitOrderNotFound
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
		return sdk.Coin{}, sdk.Coin{}, nil, sdkerrors.Wrapf(types.ErrCancelEmptyLimitOrder, "%s", tranche.Key.TrancheKey)
	}

	k.SaveTrancheUser(ctx, trancheUser)
	k.SaveTranche(ctx, tranche)

	if trancheUser.OrderType.HasExpiration() {
		k.RemoveLimitOrderExpiration(ctx, *tranche.ExpirationTime, tranche.Key.KeyMarshal())
	}

	makerCoinOut = sdk.NewCoin(tradePairID.MakerDenom, makerAmountToReturn)
	takerCoinOut = sdk.NewCoin(tradePairID.TakerDenom, takerAmountOut)

	return makerCoinOut, takerCoinOut, tradePairID, nil
}
