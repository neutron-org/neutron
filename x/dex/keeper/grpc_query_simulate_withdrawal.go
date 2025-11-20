package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v8/x/dex/types"
)

func (k Keeper) SimulateWithdrawal(
	goCtx context.Context,
	req *types.QuerySimulateWithdrawalRequest,
) (*types.QuerySimulateWithdrawalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	msg := req.Msg

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(msg.Receiver)
	pairID, err := types.NewPairID(msg.TokenA, msg.TokenB)
	if err != nil {
		return nil, err
	}

	tickIndexes := NormalizeAllTickIndexes(msg.TokenA, pairID.Token0, msg.TickIndexesAToB)

	poolsToRemoveFrom, err := k.PoolDataToPools(ctx, pairID, tickIndexes, msg.Fees)
	if err != nil {
		return nil, err
	}

	reserve0Withdrawn, reserve1Withdrawn, sharesBurned, _, err := k.ExecuteWithdraw(
		cacheCtx,
		pairID,
		callerAddr,
		receiverAddr,
		poolsToRemoveFrom,
		msg.SharesToRemove,
	)
	if err != nil {
		return nil, err
	}

	return &types.QuerySimulateWithdrawalResponse{
		Resp: &types.MsgWithdrawalResponse{
			Reserve0Withdrawn:    reserve0Withdrawn.TruncateInt(),
			Reserve1Withdrawn:    reserve1Withdrawn.TruncateInt(),
			DecReserve0Withdrawn: reserve0Withdrawn,
			DecReserve1Withdrawn: reserve1Withdrawn,
			SharesBurned:         sharesBurned,
		},
	}, nil
}

func (k Keeper) SimulateWithdrawalWithShares(
	goCtx context.Context,
	req *types.QuerySimulateWithdrawalWithSharesRequest,
) (*types.QuerySimulateWithdrawalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	msg := req.Msg

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(msg.Receiver)
	poolsToRemoveFrom, shareAmountsToRemove, err := k.SharesToPools(ctx, msg.SharesToRemove)
	if err != nil {
		return nil, err
	}

	// This is
	pairID := poolsToRemoveFrom[0].MustPairID()

	reserve0Withdrawn, reserve1Withdrawn, sharesBurned, _, err := k.ExecuteWithdraw(
		cacheCtx,
		pairID,
		callerAddr,
		receiverAddr,
		poolsToRemoveFrom,
		shareAmountsToRemove,
	)
	if err != nil {
		return nil, err
	}

	return &types.QuerySimulateWithdrawalResponse{
		Resp: &types.MsgWithdrawalResponse{
			Reserve0Withdrawn:    reserve0Withdrawn.TruncateInt(),
			Reserve1Withdrawn:    reserve1Withdrawn.TruncateInt(),
			DecReserve0Withdrawn: reserve0Withdrawn,
			DecReserve1Withdrawn: reserve1Withdrawn,
			SharesBurned:         sharesBurned,
		},
	}, nil
}
