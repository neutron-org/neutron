package keeper

import (
	"context"

    "github.com/neutron-org/neutron/x/contractmanager/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)


func (k msgServer) CreateFailure(goCtx context.Context,  msg *types.MsgCreateFailure) (*types.MsgCreateFailureResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // Check if the value already exists
    _, isFound := k.GetFailure(
        ctx,
        msg.Index,
        )
    if isFound {
        return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "index already set")
    }

    var failure = types.Failure{
        Creator: msg.Creator,
        Index: msg.Index,
        ContractAddress: msg.ContractAddress,
        AckId: msg.AckId,
        AckType: msg.AckType,
        
    }

   k.SetFailure(
   		ctx,
   		failure,
   	)
	return &types.MsgCreateFailureResponse{}, nil
}

func (k msgServer) UpdateFailure(goCtx context.Context,  msg *types.MsgUpdateFailure) (*types.MsgUpdateFailureResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // Check if the value exists
    valFound, isFound := k.GetFailure(
        ctx,
        msg.Index,
    )
    if !isFound {
        return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
    }

    // Checks if the the msg creator is the same as the current owner
    if msg.Creator != valFound.Creator {
        return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
    }

    var failure = types.Failure{
		Creator: msg.Creator,
		Index: msg.Index,
        ContractAddress: msg.ContractAddress,
		AckId: msg.AckId,
		AckType: msg.AckType,
		
	}

	k.SetFailure(ctx, failure)

	return &types.MsgUpdateFailureResponse{}, nil
}

func (k msgServer) DeleteFailure(goCtx context.Context,  msg *types.MsgDeleteFailure) (*types.MsgDeleteFailureResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // Check if the value exists
    valFound, isFound := k.GetFailure(
        ctx,
        msg.Index,
    )
    if !isFound {
        return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
    }

    // Checks if the the msg creator is the same as the current owner
    if msg.Creator != valFound.Creator {
        return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
    }

	k.RemoveFailure(
	    ctx,
	msg.Index,
    )

	return &types.MsgDeleteFailureResponse{}, nil
}