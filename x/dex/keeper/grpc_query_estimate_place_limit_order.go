package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dex/types"
)

func (k Keeper) EstimatePlaceLimitOrder(
	goCtx context.Context,
	req *types.QueryEstimatePlaceLimitOrderRequest,
) (*types.QueryEstimatePlaceLimitOrderResponse, error) {
	msg := types.MsgPlaceLimitOrder{
		// Add a random address so that Validate passes. This address is not used for anything within the query
		Creator:          "neutron1dft8nwxzr0u27wvr2cknpermjkreqvp9fdy0uz",
		Receiver:         "neutron1dft8nwxzr0u27wvr2cknpermjkreqvp9fdy0uz",
		TokenIn:          req.TokenIn,
		TokenOut:         req.TokenOut,
		TickIndexInToOut: req.TickIndexInToOut,
		AmountIn:         req.AmountIn,
		OrderType:        req.OrderType,
		ExpirationTime:   req.ExpirationTime,
		MaxAmountOut:     req.MaxAmountOut,
	}
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	err := msg.ValidateGoodTilExpiration(ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	oldBk := k.bankKeeper
	k.bankKeeper = NewSimulationBankKeeper(k.bankKeeper)
	_, totalInCoin, swapInCoin, swapOutCoin, err := k.PlaceLimitOrderCore(
		cacheCtx,
		req.TokenIn,
		req.TokenOut,
		req.AmountIn,
		req.TickIndexInToOut,
		req.OrderType,
		req.ExpirationTime,
		req.MaxAmountOut,
		[]byte("caller"),
		[]byte("receiver"),
	)
	if err != nil {
		return nil, err
	}

	//nolint:staticcheck // Should be unnecessary but out of an abundance of caution we restore the old bankkeeper
	k.bankKeeper = oldBk

	return &types.QueryEstimatePlaceLimitOrderResponse{
		TotalInCoin: totalInCoin,
		SwapInCoin:  swapInCoin,
		SwapOutCoin: swapOutCoin,
	}, nil
}
