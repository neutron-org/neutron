package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// TODO: This doesn't run ValidateBasic() checks.
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
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	callerAddr := sdk.MustAccAddressFromBech32(req.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(req.Receiver)

	blockTime := cacheCtx.BlockTime()
	if req.OrderType.IsGoodTil() && !req.ExpirationTime.After(blockTime) {
		return nil, sdkerrors.Wrapf(types.ErrExpirationTimeInPast,
			"Current BlockTime: %s; Provided ExpirationTime: %s",
			blockTime.String(),
			req.ExpirationTime.String(),
		)
	}

	_, totalInCoin, swapInCoin, swapOutCoin, err := k.PlaceLimitOrderCore(
		cacheCtx,
		req.TokenIn,
		req.TokenOut,
		req.AmountIn,
		req.TickIndexInToOut,
		req.OrderType,
		req.ExpirationTime,
		req.MaxAmountOut,
		nil,
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
