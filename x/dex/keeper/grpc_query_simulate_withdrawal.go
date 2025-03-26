package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
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

	reserve0Withdrawn, reserve1Withdrawn, sharesBurned, _, err := k.ExecuteWithdraw(
		cacheCtx,
		pairID,
		callerAddr,
		receiverAddr,
		msg.SharesToRemove,
		tickIndexes,
		msg.Fees,
	)
	if err != nil {
		return nil, err
	}

	return &types.QuerySimulateWithdrawalResponse{
		Resp: &types.MsgWithdrawalResponse{
			Reserve0Withdrawn: reserve0Withdrawn,
			Reserve1Withdrawn: reserve1Withdrawn,
			SharesBurned:      sharesBurned,
		},
	}, nil
}
