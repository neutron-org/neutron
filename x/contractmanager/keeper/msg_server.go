package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/v6/x/contractmanager/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams updates the module parameters
func (k Keeper) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgUpdateParams")
	}

	authority := k.GetAuthority()
	if authority != req.Authority {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// ResubmitFailure resubmits the failure after contract acknowledgement failed
func (k Keeper) ResubmitFailure(goCtx context.Context, req *types.MsgResubmitFailure) (*types.MsgResubmitFailureResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgResubmitFailure")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, errors.Wrap(err, "sender in resubmit request is not in correct address format")
	}

	if !k.wasmKeeper.HasContractInfo(ctx, sender) {
		return nil, errors.Wrap(types.ErrNotContractResubmission, "sender in resubmit request is not a smart contract")
	}

	failure, err := k.GetFailure(ctx, sender, req.FailureId)
	if err != nil {
		return nil, errors.Wrap(sdkerrors.ErrNotFound, "no failure with given FailureId found to resubmit")
	}

	if err := k.resubmitFailure(ctx, sender, failure); err != nil {
		return nil, err
	}

	return &types.MsgResubmitFailureResponse{}, nil
}
