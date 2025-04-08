package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (k Keeper) SimulateCancelLimitOrder(
	goCtx context.Context,
	req *types.QuerySimulateCancelLimitOrderRequest,
) (*types.QuerySimulateCancelLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	msg := req.Msg

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)
	makerCoinOut, takerCoinOut, err := k.ExecuteCancelLimitOrder(
		cacheCtx,
		msg.TrancheKey,
		callerAddr,
	)
	if err != nil {
		return nil, err
	}

	return &types.QuerySimulateCancelLimitOrderResponse{
		Resp: &types.MsgCancelLimitOrderResponse{
			TakerCoinOut: takerCoinOut,
			MakerCoinOut: makerCoinOut,
		},
	}, nil
}
