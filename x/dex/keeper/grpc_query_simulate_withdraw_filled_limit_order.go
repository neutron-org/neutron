package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (k Keeper) SimulateWithdrawFilledLimitOrder(
	goCtx context.Context,
	req *types.QuerySimulateWithdrawFilledLimitOrderRequest,
) (*types.QuerySimulateWithdrawFilledLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	msg := req.Msg

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)
	takerCoinOut, makerCoinOut, err := k.ExecuteWithdrawFilledLimitOrder(
		cacheCtx,
		msg.TrancheKey,
		callerAddr,
	)
	if err != nil {
		return nil, err
	}

	return &types.QuerySimulateWithdrawFilledLimitOrderResponse{
		Resp: &types.MsgWithdrawFilledLimitOrderResponse{
			TakerCoinOut: takerCoinOut,
			MakerCoinOut: makerCoinOut,
		},
	}, nil
}
