package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (k Keeper) SimulateDeposit(
	goCtx context.Context,
	req *types.QuerySimulateDepositRequest,
) (*types.QuerySimulateDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	msg := req.Msg
	msg.Creator = types.DummyAddress
	msg.Receiver = types.DummyAddress

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(msg.Receiver)
	pairID, err := types.NewPairID(msg.TokenA, msg.TokenB)
	if err != nil {
		return nil, err
	}

	// sort amounts
	amounts0, amounts1 := SortAmounts(msg.TokenA, pairID.Token0, msg.AmountsA, msg.AmountsB)

	tickIndexes := NormalizeAllTickIndexes(msg.TokenA, pairID.Token0, msg.TickIndexesAToB)

	//nolint:dogsled
	reserve0Deposited, reserve1Deposited, _, _, sharesIssued, _, failedDeposits, err := k.ExecuteDeposit(
		cacheCtx,
		pairID,
		callerAddr,
		receiverAddr,
		amounts0,
		amounts1,
		tickIndexes,
		msg.Fees,
		msg.Options,
	)
	if err != nil {
		return nil, err
	}

	return &types.QuerySimulateDepositResponse{
		Resp: &types.MsgDepositResponse{
			Reserve0Deposited: reserve0Deposited,
			Reserve1Deposited: reserve1Deposited,
			FailedDeposits:    failedDeposits,
			SharesIssued:      sharesIssued,
		},
	}, nil
}
