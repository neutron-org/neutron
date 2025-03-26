package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (k Keeper) SimulateMultiHopSwap(
	goCtx context.Context,
	req *types.QuerySimulateMultiHopSwapRequest,
) (*types.QuerySimulateMultiHopSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	msg := req.Msg
	msg.Creator = types.DummyAddress
	msg.Receiver = types.DummyAddress

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	bestRoute, _, err := k.CalulateMultiHopSwap(
		cacheCtx,
		msg.AmountIn,
		msg.Routes,
		msg.ExitLimitPrice,
		msg.PickBestRoute,
	)
	if err != nil {
		return nil, err
	}

	return &types.QuerySimulateMultiHopSwapResponse{
		Resp: &types.MsgMultiHopSwapResponse{
			CoinOut: bestRoute.coinOut,
			Dust:    bestRoute.dust,
			Route:   &types.MultiHopRoute{Hops: bestRoute.route},
		},
	}, nil
}
