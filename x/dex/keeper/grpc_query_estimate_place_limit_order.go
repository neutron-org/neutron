package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v2/x/dex/types"
)

func (k Keeper) EstimatePlaceLimitOrder(
	goCtx context.Context,
	req *types.QueryEstimatePlaceLimitOrderRequest,
) (*types.QueryEstimatePlaceLimitOrderResponse, error) {
	msg := types.MsgPlaceLimitOrder{
		Creator:          req.Creator,
		Receiver:         req.Receiver,
		TokenIn:          req.TokenIn,
		TokenOut:         req.TokenOut,
		TickIndexInToOut: req.TickIndexInToOut,
		AmountIn:         req.AmountIn,
		OrderType:        req.OrderType,
		ExpirationTime:   req.ExpirationTime,
		MaxAmountOut:     req.MaxAmountOut,
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()
	cacheGoCtx := sdk.WrapSDKContext(cacheCtx)

	callerAddr := sdk.MustAccAddressFromBech32(req.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(req.Receiver)

	err := msg.ValidateGoodTilExpiration(cacheCtx.BlockTime().Unix())
	if err != nil {
		return nil, err
	}

	_, totalInCoin, swapInCoin, swapOutCoin, err := k.PlaceLimitOrderCore(
		cacheGoCtx,
		req.TokenIn,
		req.TokenOut,
		req.AmountIn,
		req.TickIndexInToOut,
		req.OrderType,
		req.ExpirationTime,
		req.MaxAmountOut,
		callerAddr,
		receiverAddr,
	)
	if err != nil {
		return nil, err
	}

	// NB: We're only using a cache context so we don't expect any writes to happen.

	return &types.QueryEstimatePlaceLimitOrderResponse{
		TotalInCoin: totalInCoin,
		SwapInCoin:  swapInCoin,
		SwapOutCoin: swapOutCoin,
	}, nil
}
