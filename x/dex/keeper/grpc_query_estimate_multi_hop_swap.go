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

	oldBk := k.bankKeeper
	k.bankKeeper = NewSimulationBankKeeper(k.bankKeeper)
	coinOut, err := k.MultiHopSwapCore(
		cacheCtx,
		req.AmountIn,
		req.Routes,
		req.ExitLimitPrice,
		req.PickBestRoute,
		[]byte("caller"),
		[]byte("receiver"),
	)
	if err != nil {
		return nil, err
	}

	//nolint:staticcheck // Should be unnecessary but out of an abundance of caution we restore the old bankkeeper
	k.bankKeeper = oldBk

	return &types.QueryEstimateMultiHopSwapResponse{CoinOut: coinOut}, nil
}
