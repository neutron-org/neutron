package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dex/types"
)

func (k Keeper) EstimateMultiHopSwap(
	goCtx context.Context,
	req *types.QueryEstimateMultiHopSwapRequest,
) (*types.QueryEstimateMultiHopSwapResponse, error) {
	msg := types.MsgMultiHopSwap{
		// Add a random address so that Validate passes. This address is not used for anything within the query
		Creator:        "neutron1dft8nwxzr0u27wvr2cknpermjkreqvp9fdy0uz",
		Receiver:       "neutron1dft8nwxzr0u27wvr2cknpermjkreqvp9fdy0uz",
		Routes:         req.Routes,
		AmountIn:       req.AmountIn,
		ExitLimitPrice: req.ExitLimitPrice,
		PickBestRoute:  req.PickBestRoute,
	}
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	bestRoute, _, err := k.CalulateMultiHopSwap(
		cacheCtx,
		req.AmountIn,
		req.Routes,
		req.ExitLimitPrice,
		req.PickBestRoute,
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryEstimateMultiHopSwapResponse{CoinOut: bestRoute.coinOut}, nil
}
