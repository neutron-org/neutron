package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/types"
)

// TODO: This doesn't run ValidateBasic() checks.
func (k Keeper) EstimateMultiHopSwap(
	goCtx context.Context,
	req *types.QueryEstimateMultiHopSwapRequest,
) (*types.QueryEstimateMultiHopSwapResponse, error) {
	msg := types.MsgMultiHopSwap{
		Creator:        req.Creator,
		Receiver:       req.Receiver,
		Routes:         req.Routes,
		AmountIn:       req.AmountIn,
		ExitLimitPrice: req.ExitLimitPrice,
		PickBestRoute:  req.PickBestRoute,
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()
	cacheGoCtx := sdk.WrapSDKContext(cacheCtx)

	callerAddr := sdk.MustAccAddressFromBech32(req.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(req.Receiver)

	coinOut, err := k.MultiHopSwapCore(
		cacheGoCtx,
		req.AmountIn,
		req.Routes,
		req.ExitLimitPrice,
		req.PickBestRoute,
		callerAddr,
		receiverAddr,
	)
	if err != nil {
		return nil, err
	}

	// NB: Critically, we do not write the best route's buffered state context since this is only an estimate.

	return &types.QueryEstimateMultiHopSwapResponse{CoinOut: coinOut}, nil
}
